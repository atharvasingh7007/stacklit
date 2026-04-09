package parser

import (
	"os"
	"strings"
	"testing"
)

func TestGoParserCanParse(t *testing.T) {
	p := &GoParser{}

	cases := []struct {
		filename string
		want     bool
	}{
		{"main.go", true},
		{"handler.go", true},
		{"foo_test.go", false},
		{"main_test.go", false},
		{"main.ts", false},
		{"main.py", false},
		{"README.md", false},
	}

	for _, tc := range cases {
		got := p.CanParse(tc.filename)
		if got != tc.want {
			t.Errorf("CanParse(%q) = %v, want %v", tc.filename, got, tc.want)
		}
	}
}

func TestGoParserParse_Imports(t *testing.T) {
	p := &GoParser{}

	content, err := os.ReadFile("../../testdata/go-project/main.go")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	info, err := p.Parse("../../testdata/go-project/main.go", content)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if info.Language != "go" {
		t.Errorf("Language = %q, want %q", info.Language, "go")
	}

	wantImports := []string{"fmt", "net/http", "github.com/example/pkg/handler"}
	for _, want := range wantImports {
		found := false
		for _, got := range info.Imports {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("import %q not found in %v", want, info.Imports)
		}
	}
}

func TestGoParserParse_Entrypoint(t *testing.T) {
	p := &GoParser{}

	content, err := os.ReadFile("../../testdata/go-project/main.go")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	info, err := p.Parse("../../testdata/go-project/main.go", content)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if !info.IsEntrypoint {
		t.Error("IsEntrypoint = false, want true for main package with main()")
	}
}

func TestGoParserParse_Exports(t *testing.T) {
	p := &GoParser{}

	content, err := os.ReadFile("../../testdata/go-project/internal/handler.go")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	info, err := p.Parse("../../testdata/go-project/internal/handler.go", content)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// Handle should be exported (as a full signature starting with "Handle"),
	// processRequest should not appear at all.
	foundHandle := false
	foundProcessRequest := false
	for _, exp := range info.Exports {
		if strings.HasPrefix(exp, "Handle") {
			foundHandle = true
		}
		if strings.Contains(exp, "processRequest") {
			foundProcessRequest = true
		}
	}

	if !foundHandle {
		t.Errorf("exported name starting with %q not found in %v", "Handle", info.Exports)
	}
	if foundProcessRequest {
		t.Errorf("unexported name %q should not be in exports %v", "processRequest", info.Exports)
	}
}

func TestGoParserParse_LineCount(t *testing.T) {
	p := &GoParser{}

	content := []byte("package main\n\nfunc main() {}\n")
	info, err := p.Parse("fake.go", content)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if info.LineCount != 3 {
		t.Errorf("LineCount = %d, want 3", info.LineCount)
	}
}
