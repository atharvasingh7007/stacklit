package monorepo

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDetectPnpmWorkspace verifies that the pnpm-workspace.yaml fixture is
// recognised as a pnpm monorepo with at least 3 workspace paths.
func TestDetectPnpmWorkspace(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "monorepo")

	result, err := Detect(root)
	if err != nil {
		t.Fatalf("Detect returned unexpected error: %v", err)
	}

	if result.Type != "monorepo" {
		t.Errorf("Type = %q, want %q", result.Type, "monorepo")
	}
	if result.Tool != "pnpm" {
		t.Errorf("Tool = %q, want %q", result.Tool, "pnpm")
	}
	if len(result.Workspaces) < 3 {
		t.Errorf("len(Workspaces) = %d, want >= 3; got %v", len(result.Workspaces), result.Workspaces)
	}
}

// TestDetectSingleRepo verifies that an ordinary Go project without workspace
// files is reported as a single repo.
func TestDetectSingleRepo(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "go-project")

	result, err := Detect(root)
	if err != nil {
		t.Fatalf("Detect returned unexpected error: %v", err)
	}

	if result.Type != "single" {
		t.Errorf("Type = %q, want %q", result.Type, "single")
	}
	if len(result.Workspaces) != 0 {
		t.Errorf("Workspaces = %v, want empty", result.Workspaces)
	}
}

// TestDetectGoWorkspace creates a temporary directory with a go.work file and
// two module directories, then verifies detection returns tool=go-work.
func TestDetectGoWorkspace(t *testing.T) {
	dir := t.TempDir()

	// Create two module directories.
	modA := filepath.Join(dir, "api")
	modB := filepath.Join(dir, "core")
	if err := os.MkdirAll(modA, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(modB, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a go.work file referencing both modules.
	goWork := "go 1.22\n\nuse (\n\t./api\n\t./core\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.work"), []byte(goWork), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect returned unexpected error: %v", err)
	}

	if result.Type != "monorepo" {
		t.Errorf("Type = %q, want %q", result.Type, "monorepo")
	}
	if result.Tool != "go-work" {
		t.Errorf("Tool = %q, want %q", result.Tool, "go-work")
	}
	if len(result.Workspaces) != 2 {
		t.Errorf("len(Workspaces) = %d, want 2; got %v", len(result.Workspaces), result.Workspaces)
	}
}

// TestParsePnpmYAML tests the internal YAML parser for pnpm-workspace.yaml.
func TestParsePnpmYAML(t *testing.T) {
	input := "packages:\n  - 'packages/*'\n  - 'apps/*'\n"
	got := parsePnpmYAML(input)
	want := []string{"packages/*", "apps/*"}

	if len(got) != len(want) {
		t.Fatalf("parsePnpmYAML: got %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("parsePnpmYAML[%d] = %q, want %q", i, got[i], w)
		}
	}
}

// TestParseGoWork tests go.work block-style use directive parsing.
func TestParseGoWork(t *testing.T) {
	dir := t.TempDir()

	modA := filepath.Join(dir, "svc-a")
	modB := filepath.Join(dir, "svc-b")
	_ = os.MkdirAll(modA, 0o755)
	_ = os.MkdirAll(modB, 0o755)

	content := "go 1.22\n\nuse (\n\t./svc-a\n\t./svc-b\n)\n"
	paths := parseGoWork(dir, content)

	if len(paths) != 2 {
		t.Fatalf("parseGoWork: got %v, want 2 entries", paths)
	}
}

// TestResolveGlobs verifies that glob patterns expand to real directories.
func TestResolveGlobs(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "monorepo")
	patterns := []string{"packages/*", "apps/*"}

	got := resolveGlobs(root, patterns)
	if len(got) < 3 {
		t.Errorf("resolveGlobs: got %v, want >= 3 dirs", got)
	}
}
