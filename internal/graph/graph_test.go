package graph

import (
	"testing"

	"github.com/glincker/stacklit/internal/parser"
)

// assertStringSliceContains fails the test if want is not present in got.
func assertStringSliceContains(t *testing.T, got []string, want string) {
	t.Helper()
	for _, v := range got {
		if v == want {
			return
		}
	}
	t.Errorf("expected slice to contain %q, got %v", want, got)
}

// TestDetectModules verifies that 5 files across 3 directories produce 3 modules
// and that the auth module depends on db.
func TestDetectModules(t *testing.T) {
	files := []*parser.FileInfo{
		{Path: "src/auth/login.ts", Language: "TypeScript", Imports: []string{"src/db"}, LineCount: 40},
		{Path: "src/auth/register.ts", Language: "TypeScript", Imports: []string{"src/db"}, LineCount: 30},
		{Path: "src/db/client.ts", Language: "TypeScript", Exports: []string{"Client"}, LineCount: 60},
		{Path: "src/db/models.ts", Language: "TypeScript", Exports: []string{"User"}, LineCount: 50},
		{Path: "src/api/routes.ts", Language: "TypeScript", Imports: []string{"src/auth", "src/db"}, LineCount: 80},
	}

	g := Build(files)

	mods := g.Modules()
	if len(mods) != 3 {
		t.Fatalf("expected 3 modules, got %d: %v", len(mods), moduleNames(mods))
	}

	authMod := g.Module("src/auth")
	if authMod == nil {
		t.Fatal("expected module src/auth to exist")
	}
	assertStringSliceContains(t, authMod.DependsOn, "src/db")
}

// TestModuleExports verifies that exports from multiple files in the same module
// are aggregated on the module.
func TestModuleExports(t *testing.T) {
	files := []*parser.FileInfo{
		{Path: "src/utils/format.ts", Language: "TypeScript", Exports: []string{"formatDate", "formatCurrency"}, LineCount: 25},
		{Path: "src/utils/validate.ts", Language: "TypeScript", Exports: []string{"validateEmail", "validatePhone"}, LineCount: 35},
	}

	g := Build(files)

	utilsMod := g.Module("src/utils")
	if utilsMod == nil {
		t.Fatal("expected module src/utils to exist")
	}
	if utilsMod.FileCount != 2 {
		t.Errorf("expected FileCount=2, got %d", utilsMod.FileCount)
	}

	assertStringSliceContains(t, utilsMod.Exports, "formatDate")
	assertStringSliceContains(t, utilsMod.Exports, "formatCurrency")
	assertStringSliceContains(t, utilsMod.Exports, "validateEmail")
	assertStringSliceContains(t, utilsMod.Exports, "validatePhone")

	expectedLines := 60
	if utilsMod.LineCount != expectedLines {
		t.Errorf("expected LineCount=%d, got %d", expectedLines, utilsMod.LineCount)
	}
}

// TestDependencyEdges verifies that 3 modules with known imports produce the
// correct edges and MostDepended ordering.
func TestDependencyEdges(t *testing.T) {
	// Layout:
	//   api   → auth, db
	//   auth  → db
	//   db    (no deps)
	files := []*parser.FileInfo{
		{Path: "src/api/index.ts", Language: "TypeScript", Imports: []string{"src/auth", "src/db"}, LineCount: 50},
		{Path: "src/auth/service.ts", Language: "TypeScript", Imports: []string{"src/db"}, LineCount: 40},
		{Path: "src/db/pool.ts", Language: "TypeScript", Exports: []string{"Pool"}, LineCount: 30},
	}

	g := Build(files)

	edges := g.Edges()
	if len(edges) != 3 {
		t.Fatalf("expected 3 edges, got %d: %v", len(edges), edges)
	}

	// Verify MostDepended: db should be first (depended on by api and auth).
	mostDepended := g.MostDepended()
	if len(mostDepended) == 0 {
		t.Fatal("MostDepended returned empty slice")
	}
	if mostDepended[0] != "src/db" {
		t.Errorf("expected src/db to be most depended, got %q", mostDepended[0])
	}

	// db module should be depended on by both api and auth.
	dbMod := g.Module("src/db")
	if dbMod == nil {
		t.Fatal("expected module src/db to exist")
	}
	if len(dbMod.DependedBy) != 2 {
		t.Errorf("expected db.DependedBy length=2, got %d: %v", len(dbMod.DependedBy), dbMod.DependedBy)
	}
	assertStringSliceContains(t, dbMod.DependedBy, "src/api")
	assertStringSliceContains(t, dbMod.DependedBy, "src/auth")
}

// TestEntrypoints verifies that files flagged IsEntrypoint are tracked.
func TestEntrypoints(t *testing.T) {
	files := []*parser.FileInfo{
		{Path: "src/main.ts", Language: "TypeScript", IsEntrypoint: true, LineCount: 10},
		{Path: "src/auth/login.ts", Language: "TypeScript", IsEntrypoint: false, LineCount: 20},
		{Path: "src/api/index.ts", Language: "TypeScript", IsEntrypoint: true, LineCount: 15},
	}

	g := Build(files)

	eps := g.Entrypoints()
	if len(eps) != 2 {
		t.Fatalf("expected 2 entrypoints, got %d: %v", len(eps), eps)
	}
	assertStringSliceContains(t, eps, "src/main.ts")
	assertStringSliceContains(t, eps, "src/api/index.ts")
}

// moduleNames is a helper to extract names for error messages.
func moduleNames(mods []*Module) []string {
	names := make([]string, len(mods))
	for i, m := range mods {
		names[i] = m.Name
	}
	return names
}
