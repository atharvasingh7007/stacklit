package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GLINCKER/stacklit/internal/git"
	"github.com/GLINCKER/stacklit/internal/graph"
	"github.com/GLINCKER/stacklit/internal/monorepo"
	"github.com/GLINCKER/stacklit/internal/parser"
	"github.com/GLINCKER/stacklit/internal/renderer"
	"github.com/GLINCKER/stacklit/internal/schema"
	"github.com/GLINCKER/stacklit/internal/walker"
)

// Options configures an engine Run.
type Options struct {
	Root        string
	Workspace   string
	InstallHook bool
	Quiet       bool
}

// Result holds the output paths and assembled index from a Run.
type Result struct {
	JSONPath    string
	HTMLPath    string
	MermaidPath string
	Index       *schema.Index
	Duration    time.Duration
}

// purposeMap maps common directory names to human-readable descriptions.
var purposeMap = map[string]string{
	"auth":       "Authentication and authorization",
	"api":        "API endpoints and handlers",
	"db":         "Database access layer",
	"models":     "Data models and types",
	"config":     "Configuration management",
	"components": "UI components",
	"hooks":      "React hooks",
	"cmd":        "Application entrypoints",
	"internal":   "Private application packages",
	"pkg":        "Public packages",
	"lib":        "Shared library code",
	"utils":      "Utility functions",
	"services":   "Business logic services",
	"middleware":  "HTTP middleware",
	"cli":        "Command-line interface",
	"schema":     "Data schema definitions",
	"renderer":   "Output renderers",
	"walker":     "File system walker",
	"graph":      "Dependency graph",
	"engine":     "Core orchestration engine",
	"git":        "Git integration",
	"assets":     "Static assets",
	"parser":     "Source code parsers",
	"monorepo":   "Monorepo detection",
}

// inferPurpose returns a human-readable description for a module path.
func inferPurpose(name string) string {
	// Use the last path segment for lookup.
	base := filepath.Base(name)
	if desc, ok := purposeMap[base]; ok {
		return desc
	}
	// Fall back to the name itself, capitalised.
	if base == "." || base == "" || base == "root" {
		return "Root package"
	}
	return strings.Title(strings.ReplaceAll(base, "_", " ")) //nolint:staticcheck
}

// detectTestCommand inspects root for well-known build files and returns the
// appropriate test command string.
func detectTestCommand(root string) string {
	checks := []struct {
		file string
		cmd  string
	}{
		{"package.json", "npm test"},
		{"go.mod", "go test ./..."},
		{"requirements.txt", "pytest"},
		{"pyproject.toml", "pytest"},
		{"Cargo.toml", "cargo test"},
		{"Makefile", "make test"},
	}
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(root, c.file)); err == nil {
			return c.cmd
		}
	}
	return ""
}

// Run executes the full stacklit pipeline and returns the result.
func Run(opts Options) (*Result, error) {
	start := time.Now()

	// 1. Resolve absolute root.
	root, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, fmt.Errorf("resolving root: %w", err)
	}

	// 2. Detect monorepo layout.
	mono, err := monorepo.Detect(root)
	if err != nil {
		// Non-fatal — treat as single repo.
		mono = &monorepo.Result{Type: "single"}
	}
	if !opts.Quiet && mono.Type == "monorepo" {
		fmt.Printf("[stacklit] monorepo detected: %s (%d workspaces)\n", mono.Tool, len(mono.Workspaces))
	}

	// 3. Walk the filesystem.
	files, err := walker.Walk(root)
	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", root, err)
	}
	if !opts.Quiet {
		fmt.Printf("[stacklit] found %d files\n", len(files))
	}

	// 4. Parse all files.
	parsed, parseErrs := parser.ParseAll(files)
	if !opts.Quiet {
		fmt.Printf("[stacklit] parsed %d files (%d errors)\n", len(parsed), len(parseErrs))
	}

	// 5. Build dependency graph.
	g := graph.Build(parsed)

	// 6. Get git activity.
	activity, err := git.GetActivity(root, 90)
	if err != nil {
		// Non-fatal.
		activity = &git.Activity{}
	}

	// 7. Assemble schema.Index.
	idx := assembleIndex(root, mono, files, parsed, g, activity)

	// 8. Compute Merkle hash.
	contents := make(map[string][]byte, len(files))
	for _, f := range files {
		data, readErr := os.ReadFile(f)
		if readErr == nil {
			contents[f] = data
		}
	}
	idx.MerkleHash = git.ComputeMerkle(files, contents)

	// 9. Write outputs.
	jsonPath := filepath.Join(root, "stacklit.json")
	mmdPath := filepath.Join(root, "stacklit.mmd")
	htmlPath := filepath.Join(root, "stacklit.html")

	if err := renderer.WriteJSON(idx, jsonPath); err != nil {
		return nil, fmt.Errorf("writing JSON: %w", err)
	}
	if err := renderer.WriteMermaid(idx, mmdPath); err != nil {
		return nil, fmt.Errorf("writing Mermaid: %w", err)
	}
	if err := renderer.WriteHTML(idx, htmlPath); err != nil {
		return nil, fmt.Errorf("writing HTML: %w", err)
	}

	// 10. Install git hook if requested.
	if opts.InstallHook {
		if hookErr := git.InstallHook(root); hookErr != nil && !opts.Quiet {
			fmt.Printf("[stacklit] warning: could not install hook: %v\n", hookErr)
		}
	}

	dur := time.Since(start)

	// 11. Print summary.
	if !opts.Quiet {
		fmt.Printf("[stacklit] done in %s — wrote %s, %s, %s\n",
			dur.Round(time.Millisecond), jsonPath, mmdPath, htmlPath)
	}

	return &Result{
		JSONPath:    jsonPath,
		HTMLPath:    htmlPath,
		MermaidPath: mmdPath,
		Index:       idx,
		Duration:    dur,
	}, nil
}

// assembleIndex builds a schema.Index from the pipeline outputs.
func assembleIndex(
	root string,
	mono *monorepo.Result,
	files []string,
	parsed []*parser.FileInfo,
	g *graph.Graph,
	activity *git.Activity,
) *schema.Index {
	// --- Project ---
	projectName := filepath.Base(root)
	projectType := mono.Type

	// --- Tech: count languages ---
	langStats := map[string]schema.LangStats{}
	totalLines := 0
	for _, f := range parsed {
		lang := strings.ToLower(f.Language)
		ls := langStats[lang]
		ls.Files++
		ls.Lines += f.LineCount
		langStats[lang] = ls
		totalLines += f.LineCount
	}
	primaryLang := ""
	primaryCount := 0
	for lang, ls := range langStats {
		if ls.Files > primaryCount {
			primaryCount = ls.Files
			primaryLang = lang
		}
	}

	// --- Structure ---
	rawEntrypoints := g.Entrypoints()
	var entrypoints []string
	for _, ep := range rawEntrypoints {
		if !strings.HasPrefix(ep, "testdata") {
			entrypoints = append(entrypoints, ep)
		}
	}

	// --- Modules ---
	modules := map[string]schema.ModuleInfo{}
	for _, mod := range g.Modules() {
		if strings.HasPrefix(mod.Name, "testdata") {
			continue
		}
		exports := mod.Exports
		if len(exports) > 15 {
			exports = exports[:15]
		}
		modules[mod.Name] = schema.ModuleInfo{
			Purpose:    inferPurpose(mod.Name),
			Files:      mod.FileCount,
			Exports:    exports,
			DependsOn:  mod.DependsOn,
			DependedBy: mod.DependedBy,
		}
	}

	// --- Dependencies ---
	edges := g.Edges()
	var schemaEdges [][2]string
	for _, e := range edges {
		if strings.HasPrefix(e.From, "testdata") || strings.HasPrefix(e.To, "testdata") {
			continue
		}
		schemaEdges = append(schemaEdges, [2]string{e.From, e.To})
	}
	mostDepended := g.MostDepended()
	// Cap most-depended list.
	if len(mostDepended) > 10 {
		mostDepended = mostDepended[:10]
	}

	// --- Git ---
	hotFiles := make([]schema.HotFile, len(activity.HotFiles))
	for i, hf := range activity.HotFiles {
		hotFiles[i] = schema.HotFile{Path: hf.Path, Commits90d: hf.Commits90d}
	}

	// --- Hints ---
	testCmd := detectTestCommand(root)

	// --- Workspaces ---
	var workspaces []string
	if mono != nil {
		workspaces = mono.Workspaces
	}

	return &schema.Index{
		Version: "1",
		Project: schema.Project{
			Name:       projectName,
			Root:       ".",
			Type:       projectType,
			Workspaces: workspaces,
		},
		Tech: schema.Tech{
			PrimaryLanguage: primaryLang,
			Languages:       langStats,
		},
		Structure: schema.Structure{
			TotalFiles:  len(files),
			TotalLines:  totalLines,
			Entrypoints: entrypoints,
		},
		Modules: modules,
		Dependencies: schema.Dependencies{
			Edges:        schemaEdges,
			Entrypoints:  entrypoints,
			MostDepended: mostDepended,
		},
		Git: schema.GitInfo{
			HotFiles: hotFiles,
			Recent:   activity.RecentFiles,
			Stable:   activity.StableFiles,
		},
		Hints: schema.Hints{
			TestCmd: testCmd,
		},
	}
}
