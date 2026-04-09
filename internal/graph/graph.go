package graph

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/glincker/stacklit/internal/parser"
)

// maxModuleDepth is the maximum number of path segments a module name may have.
// Deeper paths are collapsed to their ancestor at this depth.
const maxModuleDepth = 3

// skipModuleSegments is the set of path segment names that indicate a module
// (or any ancestor) should be excluded from the graph.
var skipModuleSegments = map[string]bool{
	"test":        true,
	"tests":       true,
	"__tests__":   true,
	"__test__":    true,
	"testdata":    true,
	"test_data":   true,
	"fixtures":    true,
	"__fixtures__": true,
	"examples":    true,
	"example":     true,
	"docs_src":    true,
	"e2e":         true,
	"benchmarks":  true,
	"bench":       true,
	"scripts":     true,
	".github":     true,
}

// shouldSkipModule returns true if any segment of the module path matches a
// known non-essential directory name (tests, fixtures, examples, etc.).
func shouldSkipModule(name string) bool {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if skipModuleSegments[strings.ToLower(part)] {
			return true
		}
	}
	return false
}

// Module represents a logical grouping of source files by directory.
type Module struct {
	Name       string
	FileCount  int
	Files      []string // basenames of files in this module
	Exports    []string
	DependsOn  []string
	DependedBy []string
	Languages  []string
	LineCount  int
}

// Edge represents a dependency relationship between two modules.
type Edge struct {
	From  string
	To    string
	Count int // number of imports crossing this edge
}

// Graph holds the full dependency graph built from parsed files.
type Graph struct {
	modules     map[string]*Module
	edges       map[edgeKey]*Edge
	entrypoints []string
}

type edgeKey struct {
	from string
	to   string
}

// Build constructs a dependency Graph from a slice of parsed FileInfo.
func Build(files []*parser.FileInfo) *Graph {
	g := &Graph{
		modules: make(map[string]*Module),
		edges:   make(map[edgeKey]*Edge),
	}

	// First pass: create modules and collect entrypoints.
	for _, f := range files {
		mod := detectModule(f.Path)
		if shouldSkipModule(mod) {
			continue
		}
		m, ok := g.modules[mod]
		if !ok {
			m = &Module{Name: mod}
			g.modules[mod] = m
		}
		m.FileCount++
		m.LineCount += f.LineCount
		m.Exports = appendUnique(m.Exports, f.Exports...)
		m.Languages = appendUnique(m.Languages, f.Language)
		m.Files = append(m.Files, filepath.Base(f.Path))

		if f.IsEntrypoint {
			g.entrypoints = append(g.entrypoints, f.Path)
		}
	}

	// Build a set of all known module names for import resolution.
	knownModules := make([]string, 0, len(g.modules))
	for name := range g.modules {
		knownModules = append(knownModules, name)
	}

	// Second pass: resolve imports to edges.
	for _, f := range files {
		fromMod := detectModule(f.Path)
		if shouldSkipModule(fromMod) {
			continue
		}
		fileDir := filepath.Dir(f.Path)

		for _, imp := range f.Imports {
			toMod := resolveImport(imp, fileDir, knownModules)
			if toMod == "" || toMod == fromMod {
				continue
			}
			if shouldSkipModule(toMod) {
				continue
			}
			key := edgeKey{from: fromMod, to: toMod}
			e, ok := g.edges[key]
			if !ok {
				e = &Edge{From: fromMod, To: toMod}
				g.edges[key] = e
			}
			e.Count++
		}
	}

	// Build DependsOn and DependedBy from edges.
	for key := range g.edges {
		from := g.modules[key.from]
		to := g.modules[key.to]
		from.DependsOn = appendUnique(from.DependsOn, key.to)
		to.DependedBy = appendUnique(to.DependedBy, key.from)
	}

	// Sort all slice fields for determinism.
	for _, m := range g.modules {
		sort.Strings(m.Exports)
		sort.Strings(m.DependsOn)
		sort.Strings(m.DependedBy)
		sort.Strings(m.Languages)
		sort.Strings(m.Files)
	}
	sort.Strings(g.entrypoints)

	return g
}

// Modules returns all modules sorted by name.
func (g *Graph) Modules() []*Module {
	out := make([]*Module, 0, len(g.modules))
	for _, m := range g.modules {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

// Module returns a module by name, or nil if not found.
func (g *Graph) Module(name string) *Module {
	return g.modules[name]
}

// Edges returns all edges sorted by (From, To).
func (g *Graph) Edges() []Edge {
	out := make([]Edge, 0, len(g.edges))
	for _, e := range g.edges {
		out = append(out, *e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}

// MostDepended returns module names sorted by DependedBy count descending.
// Ties are broken alphabetically.
func (g *Graph) MostDepended() []string {
	mods := g.Modules()
	sort.SliceStable(mods, func(i, j int) bool {
		ci := len(mods[i].DependedBy)
		cj := len(mods[j].DependedBy)
		if ci != cj {
			return ci > cj
		}
		return mods[i].Name < mods[j].Name
	})
	names := make([]string, len(mods))
	for i, m := range mods {
		names[i] = m.Name
	}
	return names
}

// Entrypoints returns file paths that were flagged as IsEntrypoint.
func (g *Graph) Entrypoints() []string {
	return g.entrypoints
}

// detectModule returns the parent directory of a file path as the module name,
// collapsed to at most maxModuleDepth segments. Files at the root level return
// "root".
func detectModule(filePath string) string {
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" || dir == "/" {
		return "root"
	}
	dir = filepath.ToSlash(dir)
	parts := strings.Split(dir, "/")
	if len(parts) > maxModuleDepth {
		parts = parts[:maxModuleDepth]
	}
	return strings.Join(parts, "/")
}

// resolveImport attempts to match an import string to a known module name.
// Returns the matched module name, or "" if no match (external dependency).
func resolveImport(imp, fileDir string, knownModules []string) string {
	// Relative imports: ./foo or ../foo
	if strings.HasPrefix(imp, "./") || strings.HasPrefix(imp, "../") {
		resolved := filepath.Clean(filepath.Join(fileDir, imp))
		// Try exact match first, then prefix match.
		for _, mod := range knownModules {
			if mod == resolved {
				return mod
			}
		}
		// The import may point to a file inside a module directory.
		for _, mod := range knownModules {
			if strings.HasPrefix(resolved, mod+"/") || strings.HasPrefix(resolved+"/", mod+"/") {
				return mod
			}
		}
		return ""
	}

	// Absolute/package-style imports: match suffix against known module names.
	// e.g. import "internal/auth" matches module "internal/auth".
	for _, mod := range knownModules {
		if mod == imp {
			return mod
		}
		// Suffix match: "internal/auth" matches module "internal/auth".
		if strings.HasSuffix(imp, "/"+mod) || strings.HasSuffix(imp, mod) {
			return mod
		}
		// Module is a suffix of the import path (Go-style).
		if strings.HasSuffix(imp, mod) || imp == mod {
			return mod
		}
	}

	return ""
}

// appendUnique appends values to s, skipping duplicates.
func appendUnique(s []string, values ...string) []string {
	set := make(map[string]struct{}, len(s))
	for _, v := range s {
		set[v] = struct{}{}
	}
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := set[v]; !ok {
			set[v] = struct{}{}
			s = append(s, v)
		}
	}
	return s
}
