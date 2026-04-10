package renderer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glincker/stacklit/internal/schema"
)

func makeTestIndex() *schema.Index {
	return &schema.Index{
		Tech: schema.Tech{
			PrimaryLanguage: "go",
			Languages: map[string]schema.LangStats{
				"go": {Files: 10, Lines: 500},
			},
		},
		Modules: map[string]schema.ModuleInfo{
			"src/api": {
				Purpose: "API handlers",
				Files:   5,
			},
			"src/auth": {
				Purpose: "Authentication",
				Files:   3,
			},
			"src/db": {
				Purpose: "Database access layer",
				Files:   4,
			},
		},
		Dependencies: schema.Dependencies{
			Edges: [][2]string{
				{"src/api", "src/auth"},
				{"src/api", "src/db"},
				{"src/auth", "src/db"},
			},
		},
	}
}

func TestRenderMermaid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "diagram.mmd")

	idx := makeTestIndex()
	if err := WriteMermaid(idx, path); err != nil {
		t.Fatalf("WriteMermaid returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	content := string(data)

	// Must start with graph LR
	if !strings.HasPrefix(content, "graph LR\n") {
		t.Errorf("expected content to start with 'graph LR\\n', got: %q", content[:min(len(content), 20)])
	}

	// Must contain classDef
	if !strings.Contains(content, "classDef") {
		t.Errorf("expected classDef in output, got:\n%s", content)
	}

	// Must contain node IDs for each module
	for _, id := range []string{"src_api", "src_auth", "src_db"} {
		if !strings.Contains(content, id) {
			t.Errorf("expected node ID %q in output", id)
		}
	}

	// Must contain --> arrows
	if !strings.Contains(content, "-->") {
		t.Errorf("expected '-->' edges in output, got:\n%s", content)
	}
}

func TestRenderMermaidEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.mmd")

	idx := &schema.Index{
		Tech: schema.Tech{
			PrimaryLanguage: "",
			Languages:       map[string]schema.LangStats{},
		},
		Modules:      map[string]schema.ModuleInfo{},
		Dependencies: schema.Dependencies{},
	}

	if err := WriteMermaid(idx, path); err != nil {
		t.Fatalf("WriteMermaid with empty index returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	content := string(data)

	if !strings.HasPrefix(content, "graph LR\n") {
		t.Errorf("expected empty diagram to start with 'graph LR\\n', got: %q", content)
	}
}

func TestSanitizeMermaidID(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"src/auth", "src_auth"},
		{"@test/api", "test_api"},
		{"my-module", "my_module"},
		{"foo.bar", "foo_bar"},
		{"simple", "simple"},
		{"a//b", "a_b"},
	}
	for _, tc := range cases {
		got := sanitizeMermaidID(tc.input)
		if got != tc.want {
			t.Errorf("sanitizeMermaidID(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
