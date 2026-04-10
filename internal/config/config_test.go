package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefault(t *testing.T) {
	cfg := Load(t.TempDir())
	if cfg.MaxDepth != 4 {
		t.Errorf("expected default max_depth=4, got %d", cfg.MaxDepth)
	}
	if cfg.MaxModules != 200 {
		t.Errorf("expected default max_modules=200, got %d", cfg.MaxModules)
	}
	if cfg.MaxExports != 10 {
		t.Errorf("expected default max_exports=10, got %d", cfg.MaxExports)
	}
	if cfg.Output.JSON != "stacklit.json" {
		t.Errorf("expected default output.json=stacklit.json, got %q", cfg.Output.JSON)
	}
	if cfg.Output.Mermaid != "DEPENDENCIES.md" {
		t.Errorf("expected default output.mermaid=DEPENDENCIES.md, got %q", cfg.Output.Mermaid)
	}
	if cfg.Output.HTML != "stacklit.html" {
		t.Errorf("expected default output.html=stacklit.html, got %q", cfg.Output.HTML)
	}
}

func TestLoadCustom(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".stacklitrc.json"),
		[]byte(`{"max_depth": 6, "ignore": ["custom/"]}`), 0644)
	cfg := Load(dir)
	if cfg.MaxDepth != 6 {
		t.Errorf("expected max_depth=6, got %d", cfg.MaxDepth)
	}
	if len(cfg.Ignore) != 1 {
		t.Errorf("expected 1 ignore pattern, got %d: %v", len(cfg.Ignore), cfg.Ignore)
	}
	// Unset fields should still use defaults.
	if cfg.MaxModules != 200 {
		t.Errorf("expected default max_modules=200, got %d", cfg.MaxModules)
	}
}

func TestLoadMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".stacklitrc.json"), []byte(`not json`), 0644)
	cfg := Load(dir)
	// Should fall back to defaults without panicking.
	if cfg.MaxDepth != 4 {
		t.Errorf("expected default max_depth=4 after malformed file, got %d", cfg.MaxDepth)
	}
}
