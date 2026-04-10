package parser

import (
	"os"
	"strings"
	"testing"
)


func TestPythonParserParse_Imports(t *testing.T) {
	p := &TreeSitterParser{}

	content, err := os.ReadFile("../../testdata/python-project/app.py")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	info, err := p.Parse("../../testdata/python-project/app.py", content)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if info.Language != "python" {
		t.Errorf("Language = %q, want %q", info.Language, "python")
	}

	wantImports := []string{"os", "sys", "flask", ".models", ".config"}
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

func TestPythonParserParse_Exports(t *testing.T) {
	p := &TreeSitterParser{}

	content, err := os.ReadFile("../../testdata/python-project/app.py")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	info, err := p.Parse("../../testdata/python-project/app.py", content)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// Public names should be exported (as full signatures containing the name).
	wantExports := []string{"Application", "create_app"}
	for _, want := range wantExports {
		found := false
		for _, got := range info.Exports {
			if strings.Contains(got, want) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("export %q not found in %v", want, info.Exports)
		}
	}

	// Private names should NOT be exported.
	for _, got := range info.Exports {
		if strings.Contains(got, "_internal_helper") {
			t.Errorf("private name %q should not be in exports", got)
		}
	}
}

func TestPythonParserParse_Entrypoint(t *testing.T) {
	p := &TreeSitterParser{}

	content, err := os.ReadFile("../../testdata/python-project/app.py")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	info, err := p.Parse("../../testdata/python-project/app.py", content)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if !info.IsEntrypoint {
		t.Error("IsEntrypoint = false, want true for file with if __name__ == '__main__'")
	}
}

func TestPythonParserParse_NoEntrypoint(t *testing.T) {
	p := &TreeSitterParser{}

	content := []byte(`def helper():
    pass

class MyClass:
    pass
`)

	info, err := p.Parse("lib.py", content)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if info.IsEntrypoint {
		t.Error("IsEntrypoint = true, want false for file without __main__ guard")
	}
}

func TestPythonParserParse_LineCount(t *testing.T) {
	p := &TreeSitterParser{}

	content := []byte("import os\nimport sys\n\nprint('hello')\n")
	info, err := p.Parse("script.py", content)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if info.LineCount != 4 {
		t.Errorf("LineCount = %d, want 4", info.LineCount)
	}
}
