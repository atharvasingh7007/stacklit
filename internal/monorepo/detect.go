// Package monorepo detects monorepo/workspace boundaries and reports the
// repo type along with resolved workspace paths.
package monorepo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Result holds the detection output for a given root directory.
type Result struct {
	Type       string   `json:"type"`       // "monorepo" or "single"
	Tool       string   `json:"tool"`       // e.g. "pnpm", "npm", "yarn", "go-work", "turbo", "nx", "lerna", "cargo", "convention"
	Workspaces []string `json:"workspaces"` // relative paths to workspace roots
}

// Detect inspects root and returns a Result describing the repo layout.
// Detection runs in priority order; the first match wins.
func Detect(root string) (*Result, error) {
	// 1. pnpm-workspace.yaml
	if r, err := detectPnpm(root); err == nil && r != nil {
		return r, nil
	}

	// 2. package.json "workspaces"
	if r, err := detectNpmWorkspaces(root); err == nil && r != nil {
		return r, nil
	}

	// 3. go.work
	if r, err := detectGoWork(root); err == nil && r != nil {
		return r, nil
	}

	// 4. turbo.json → delegate to package.json workspaces
	if r, err := detectTurbo(root); err == nil && r != nil {
		return r, nil
	}

	// 5. nx.json → convention dirs
	if r, err := detectNx(root); err == nil && r != nil {
		return r, nil
	}

	// 6. lerna.json → delegate to package.json workspaces
	if r, err := detectLerna(root); err == nil && r != nil {
		return r, nil
	}

	// 7. Cargo.toml [workspace]
	if r, err := detectCargo(root); err == nil && r != nil {
		return r, nil
	}

	// 8. Convention dirs
	if r, err := detectConvention(root); err == nil && r != nil {
		return r, nil
	}

	// 9. Fallback
	return &Result{Type: "single", Workspaces: []string{}}, nil
}

// resolveGlobs expands patterns like "packages/*" to actual subdirectory
// relative paths under root. Only directories are returned.
func resolveGlobs(root string, patterns []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, pattern := range patterns {
		// Build absolute glob pattern.
		absPattern := filepath.Join(root, pattern)
		matches, err := filepath.Glob(absPattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || !info.IsDir() {
				continue
			}
			rel, err := filepath.Rel(root, match)
			if err != nil {
				continue
			}
			if _, ok := seen[rel]; !ok {
				seen[rel] = struct{}{}
				result = append(result, rel)
			}
		}
	}
	return result
}

// subdirs returns the names of direct subdirectories of dir.
func subdirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs
}

// exists reports whether path exists on disk.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// readFile reads a file and returns its content, or nil on error.
func readFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}

// ── pnpm ─────────────────────────────────────────────────────────────────────

func detectPnpm(root string) (*Result, error) {
	yamlPath := filepath.Join(root, "pnpm-workspace.yaml")
	data := readFile(yamlPath)
	if data == nil {
		return nil, nil
	}

	patterns := parsePnpmYAML(string(data))
	if len(patterns) == 0 {
		return nil, nil
	}

	workspaces := resolveGlobs(root, patterns)
	return &Result{
		Type:       "monorepo",
		Tool:       "pnpm",
		Workspaces: workspaces,
	}, nil
}

// parsePnpmYAML is a minimal YAML parser that extracts the list items under
// the "packages:" key. It handles single-quoted, double-quoted, and bare values.
func parsePnpmYAML(content string) []string {
	var patterns []string
	inPackages := false

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)

		// Detect the "packages:" top-level key.
		if trimmed == "packages:" || strings.HasPrefix(trimmed, "packages:") && !strings.Contains(trimmed[len("packages:"):], " ") {
			inPackages = true
			continue
		}

		// Any non-indented, non-empty line that isn't a list item resets the section.
		if inPackages && len(line) > 0 && line[0] != ' ' && line[0] != '\t' && line[0] != '-' {
			inPackages = false
			continue
		}

		if inPackages && strings.HasPrefix(trimmed, "- ") {
			val := strings.TrimPrefix(trimmed, "- ")
			val = strings.Trim(val, `'"`)
			val = strings.TrimSpace(val)
			if val != "" {
				patterns = append(patterns, val)
			}
		}
	}
	return patterns
}

// ── npm / yarn ────────────────────────────────────────────────────────────────

// packageJSON is a minimal struct for workspace-related fields.
type packageJSON struct {
	Workspaces []string `json:"workspaces"`
}

// readPackageJSONWorkspaces parses the root package.json and returns its
// "workspaces" array, or nil if the field is absent or the file missing.
func readPackageJSONWorkspaces(root string) []string {
	data := readFile(filepath.Join(root, "package.json"))
	if data == nil {
		return nil
	}
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}
	return pkg.Workspaces
}

func detectNpmWorkspaces(root string) (*Result, error) {
	patterns := readPackageJSONWorkspaces(root)
	if len(patterns) == 0 {
		return nil, nil
	}

	workspaces := resolveGlobs(root, patterns)

	// Distinguish npm vs yarn by the presence of yarn.lock.
	tool := "npm"
	if exists(filepath.Join(root, "yarn.lock")) {
		tool = "yarn"
	}

	return &Result{
		Type:       "monorepo",
		Tool:       tool,
		Workspaces: workspaces,
	}, nil
}

// ── go.work ───────────────────────────────────────────────────────────────────

func detectGoWork(root string) (*Result, error) {
	data := readFile(filepath.Join(root, "go.work"))
	if data == nil {
		return nil, nil
	}

	workspaces := parseGoWork(root, string(data))
	return &Result{
		Type:       "monorepo",
		Tool:       "go-work",
		Workspaces: workspaces,
	}, nil
}

// parseGoWork extracts the paths from "use" directives in go.work content.
// It handles both inline ("use ./path") and block ("use (\n  ./path\n)") forms.
func parseGoWork(root, content string) []string {
	var paths []string
	inUseBlock := false

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)

		if inUseBlock {
			if trimmed == ")" {
				inUseBlock = false
				continue
			}
			p := resolvePath(root, trimmed)
			if p != "" {
				paths = append(paths, p)
			}
			continue
		}

		if trimmed == "use (" {
			inUseBlock = true
			continue
		}

		if strings.HasPrefix(trimmed, "use ") {
			rest := strings.TrimPrefix(trimmed, "use ")
			rest = strings.TrimSpace(rest)
			p := resolvePath(root, rest)
			if p != "" {
				paths = append(paths, p)
			}
		}
	}
	return paths
}

// resolvePath converts a go.work path entry (e.g. "./api") to a relative path
// from root. Returns "" if the path does not resolve to a real directory.
func resolvePath(root, entry string) string {
	if entry == "" {
		return ""
	}
	abs := filepath.Join(root, entry)
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return ""
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return ""
	}
	return rel
}

// ── turbo ─────────────────────────────────────────────────────────────────────

func detectTurbo(root string) (*Result, error) {
	if !exists(filepath.Join(root, "turbo.json")) {
		return nil, nil
	}

	// Turbo relies on package.json workspaces for the actual workspace list.
	patterns := readPackageJSONWorkspaces(root)
	workspaces := resolveGlobs(root, patterns)

	return &Result{
		Type:       "monorepo",
		Tool:       "turbo",
		Workspaces: workspaces,
	}, nil
}

// ── nx ────────────────────────────────────────────────────────────────────────

func detectNx(root string) (*Result, error) {
	if !exists(filepath.Join(root, "nx.json")) {
		return nil, nil
	}

	workspaces := resolveConventionDirs(root)
	return &Result{
		Type:       "monorepo",
		Tool:       "nx",
		Workspaces: workspaces,
	}, nil
}

// ── lerna ─────────────────────────────────────────────────────────────────────

func detectLerna(root string) (*Result, error) {
	if !exists(filepath.Join(root, "lerna.json")) {
		return nil, nil
	}

	patterns := readPackageJSONWorkspaces(root)
	workspaces := resolveGlobs(root, patterns)

	return &Result{
		Type:       "monorepo",
		Tool:       "lerna",
		Workspaces: workspaces,
	}, nil
}

// ── Cargo ─────────────────────────────────────────────────────────────────────

func detectCargo(root string) (*Result, error) {
	data := readFile(filepath.Join(root, "Cargo.toml"))
	if data == nil {
		return nil, nil
	}

	patterns := parseCargoWorkspace(string(data))
	if len(patterns) == 0 {
		return nil, nil
	}

	workspaces := resolveGlobs(root, patterns)
	return &Result{
		Type:       "monorepo",
		Tool:       "cargo",
		Workspaces: workspaces,
	}, nil
}

// parseCargoWorkspace extracts member patterns from the [workspace] section of
// a Cargo.toml file. It handles the members = [...] array.
func parseCargoWorkspace(content string) []string {
	var patterns []string
	inWorkspace := false
	inMembers := false

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)

		if trimmed == "[workspace]" {
			inWorkspace = true
			continue
		}

		// Any new top-level section ends the workspace block.
		if inWorkspace && strings.HasPrefix(trimmed, "[") && trimmed != "[workspace]" {
			inWorkspace = false
			inMembers = false
			continue
		}

		if !inWorkspace {
			continue
		}

		// Detect "members = [" possibly with values on the same line.
		if strings.HasPrefix(trimmed, "members") {
			inMembers = true
			// Extract values that might be on the same line.
			idx := strings.Index(trimmed, "[")
			if idx != -1 {
				rest := trimmed[idx+1:]
				// Check if the array closes on this line.
				if ci := strings.Index(rest, "]"); ci != -1 {
					rest = rest[:ci]
					inMembers = false
				}
				patterns = append(patterns, extractCargoValues(rest)...)
			}
			continue
		}

		if inMembers {
			if strings.Contains(trimmed, "]") {
				// Last line of the array.
				rest := strings.Split(trimmed, "]")[0]
				patterns = append(patterns, extractCargoValues(rest)...)
				inMembers = false
			} else {
				patterns = append(patterns, extractCargoValues(trimmed)...)
			}
		}
	}
	return patterns
}

// extractCargoValues parses comma-separated quoted strings from a fragment of
// a TOML array value.
func extractCargoValues(s string) []string {
	var vals []string
	for _, part := range strings.Split(s, ",") {
		v := strings.TrimSpace(part)
		v = strings.Trim(v, `"'`)
		v = strings.TrimSpace(v)
		if v != "" {
			vals = append(vals, v)
		}
	}
	return vals
}

// ── convention ────────────────────────────────────────────────────────────────

// conventionDirs are the well-known top-level directory names that indicate a
// monorepo by convention.
var conventionDirs = []string{"apps", "packages", "services", "libs"}

func detectConvention(root string) (*Result, error) {
	workspaces := resolveConventionDirs(root)
	if len(workspaces) == 0 {
		return nil, nil
	}
	return &Result{
		Type:       "monorepo",
		Tool:       "convention",
		Workspaces: workspaces,
	}, nil
}

// resolveConventionDirs looks for conventionDirs under root and returns the
// relative paths of all direct subdirectories found within them.
func resolveConventionDirs(root string) []string {
	var workspaces []string
	for _, dir := range conventionDirs {
		abs := filepath.Join(root, dir)
		if !exists(abs) {
			continue
		}
		for _, sub := range subdirs(abs) {
			rel := filepath.Join(dir, sub)
			workspaces = append(workspaces, rel)
		}
	}
	return workspaces
}
