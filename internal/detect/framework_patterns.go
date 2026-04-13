package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// FrameworkPattern holds detected structural patterns for a framework.
type FrameworkPattern struct {
	Name       string   `json:"name"`
	Config     []string `json:"config_files,omitempty"`
	Routes     string   `json:"routes,omitempty"`
	API        string   `json:"api,omitempty"`
	Middleware string   `json:"middleware,omitempty"`
	Models     string   `json:"models,omitempty"`
	Entry      string   `json:"entry,omitempty"`
}

// fileExists returns true if the file at root/rel exists and is a regular file.
func fileExists(root, rel string) bool {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil && !info.IsDir()
}

// dirExists returns true if the directory at root/rel exists.
func dirExists(root, rel string) bool {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil && info.IsDir()
}

// firstExisting returns the first candidate (relative path) that exists as a file under root,
// or "" if none exist.
func firstExisting(root string, candidates []string) string {
	for _, c := range candidates {
		if fileExists(root, c) {
			return c
		}
	}
	return ""
}

// firstExistingDir returns the first candidate (relative path) that exists as a directory under root,
// or "" if none exist.
func firstExistingDir(root string, candidates []string) string {
	for _, c := range candidates {
		if dirExists(root, c) {
			return c + "/"
		}
	}
	return ""
}

// hasPkgJSONDep returns true if the given package name appears in dependencies or devDependencies
// of the package.json at root.
func hasPkgJSONDep(root, pkg string) bool {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return false
	}
	var p struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if json.Unmarshal(data, &p) != nil {
		return false
	}
	if _, ok := p.Dependencies[pkg]; ok {
		return true
	}
	_, ok := p.DevDependencies[pkg]
	return ok
}

// hasRequirement returns true if the given package name (case-insensitive prefix match on lines)
// appears in requirements.txt at root.
func hasRequirement(root, pkg string) bool {
	data, err := os.ReadFile(filepath.Join(root, "requirements.txt"))
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	pkgLower := strings.ToLower(pkg)
	for _, line := range strings.Split(lower, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, pkgLower) {
			return true
		}
	}
	return false
}

// hasGoMod returns true if go.mod at root contains the given module path substring.
func hasGoMod(root, substr string) bool {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), substr)
}

// hasBuildFile returns true if the given file (pom.xml or build.gradle) at root contains substr.
func hasBuildFile(root, file, substr string) bool {
	data, err := os.ReadFile(filepath.Join(root, file))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), substr)
}

// DetectFrameworkPatterns inspects root and returns structural patterns for detected frameworks.
func DetectFrameworkPatterns(root string) []FrameworkPattern {
	var patterns []FrameworkPattern

	// --- Next.js ---
	nextConfigFiles := []string{"next.config.js", "next.config.ts", "next.config.mjs"}
	isNext := firstExisting(root, nextConfigFiles) != "" || hasPkgJSONDep(root, "next")
	if isNext {
		p := FrameworkPattern{Name: "Next.js"}
		for _, f := range nextConfigFiles {
			if fileExists(root, f) {
				p.Config = append(p.Config, f)
			}
		}
		// Routes: prefer app/ over pages/
		if dirExists(root, "app") {
			p.Routes = "app/"
		} else if dirExists(root, "pages") {
			p.Routes = "pages/"
		}
		// API dir
		if dirExists(root, "app/api") {
			p.API = "app/api/"
		} else if dirExists(root, "pages/api") {
			p.API = "pages/api/"
		}
		// Middleware
		if mid := firstExisting(root, []string{"middleware.ts", "middleware.js"}); mid != "" {
			p.Middleware = mid
		}
		patterns = append(patterns, p)
	}

	// --- Express ---
	if hasPkgJSONDep(root, "express") {
		p := FrameworkPattern{Name: "Express"}
		if dirExists(root, "routes") {
			p.Routes = "routes/"
		}
		if entry := firstExisting(root, []string{"app.js", "server.js"}); entry != "" {
			p.Entry = entry
		}
		patterns = append(patterns, p)
	}

	// --- FastAPI ---
	if hasRequirement(root, "fastapi") {
		p := FrameworkPattern{Name: "FastAPI"}
		if routesDir := firstExistingDir(root, []string{"app/routers", "routers"}); routesDir != "" {
			p.Routes = routesDir
		}
		if entry := firstExisting(root, []string{"app/main.py", "main.py"}); entry != "" {
			p.Entry = entry
		}
		if dirExists(root, "app/models") {
			p.Models = "app/models/"
		}
		patterns = append(patterns, p)
	}

	// --- Django ---
	isDjango := fileExists(root, "manage.py") || hasRequirement(root, "django")
	if isDjango {
		p := FrameworkPattern{Name: "Django"}
		p.Entry = "manage.py"
		patterns = append(patterns, p)
	}

	// --- Gin ---
	if hasGoMod(root, "gin-gonic/gin") {
		p := FrameworkPattern{Name: "Gin"}
		if entry := firstExisting(root, []string{"main.go", "cmd/server/main.go"}); entry != "" {
			p.Entry = entry
		}
		patterns = append(patterns, p)
	}

	// --- Spring Boot ---
	isSpring := hasBuildFile(root, "pom.xml", "spring-boot") ||
		hasBuildFile(root, "build.gradle", "spring-boot")
	if isSpring {
		p := FrameworkPattern{Name: "Spring Boot"}
		if dirExists(root, "src/main/java") {
			p.Routes = "src/main/java/"
		}
		patterns = append(patterns, p)
	}

	return patterns
}
