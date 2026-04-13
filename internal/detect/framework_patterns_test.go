package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFrameworkPatterns_NextJS(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "next.config.ts"), []byte("export default {}"), 0644)
	os.MkdirAll(filepath.Join(dir, "app", "api"), 0755)
	os.WriteFile(filepath.Join(dir, "app", "page.tsx"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "middleware.ts"), []byte(""), 0644)

	patterns := DetectFrameworkPatterns(dir)
	found := false
	for _, p := range patterns {
		if p.Name == "Next.js" {
			found = true
			if p.Routes != "app/" {
				t.Errorf("expected routes=app/, got %s", p.Routes)
			}
			if p.API != "app/api/" {
				t.Errorf("expected api=app/api/, got %s", p.API)
			}
			if p.Middleware != "middleware.ts" {
				t.Errorf("expected middleware=middleware.ts, got %s", p.Middleware)
			}
		}
	}
	if !found {
		t.Error("Next.js pattern not detected")
	}
}

func TestDetectFrameworkPatterns_Express(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"express":"^4.18.0"}}`), 0644)
	os.MkdirAll(filepath.Join(dir, "routes"), 0755)
	os.WriteFile(filepath.Join(dir, "app.js"), []byte(""), 0644)

	patterns := DetectFrameworkPatterns(dir)
	found := false
	for _, p := range patterns {
		if p.Name == "Express" {
			found = true
			if p.Routes != "routes/" {
				t.Errorf("expected routes=routes/, got %s", p.Routes)
			}
			if p.Entry != "app.js" {
				t.Errorf("expected entry=app.js, got %s", p.Entry)
			}
		}
	}
	if !found {
		t.Error("Express pattern not detected")
	}
}

func TestDetectFrameworkPatterns_FastAPI(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("fastapi==0.100.0\nuvicorn"), 0644)
	os.MkdirAll(filepath.Join(dir, "app", "routers"), 0755)
	os.WriteFile(filepath.Join(dir, "app", "main.py"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(dir, "app", "models"), 0755)

	patterns := DetectFrameworkPatterns(dir)
	found := false
	for _, p := range patterns {
		if p.Name == "FastAPI" {
			found = true
			if p.Routes != "app/routers/" {
				t.Errorf("expected routes=app/routers/, got %s", p.Routes)
			}
			if p.Entry != "app/main.py" {
				t.Errorf("expected entry=app/main.py, got %s", p.Entry)
			}
			if p.Models != "app/models/" {
				t.Errorf("expected models=app/models/, got %s", p.Models)
			}
		}
	}
	if !found {
		t.Error("FastAPI pattern not detected")
	}
}

func TestDetectFrameworkPatterns_NoFramework(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	patterns := DetectFrameworkPatterns(dir)
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns, got %d", len(patterns))
	}
}
