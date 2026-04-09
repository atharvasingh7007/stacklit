package parser

import "testing"

func TestGenericParserCanParse(t *testing.T) {
	p := &GenericParser{}

	cases := []string{
		"main.go",
		"app.py",
		"index.ts",
		"README.md",
		"Makefile",
		"Dockerfile",
		"some.unknown.ext",
		"noextension",
		"",
	}

	for _, filename := range cases {
		if !p.CanParse(filename) {
			t.Errorf("CanParse(%q) = false, want true (generic parser accepts all files)", filename)
		}
	}
}

func TestGenericParserParse_LanguageDetection(t *testing.T) {
	p := &GenericParser{}

	cases := []struct {
		filename string
		want     string
	}{
		{"main.go", "go"},
		{"app.py", "python"},
		{"index.ts", "typescript"},
		{"component.tsx", "tsx"},
		{"app.js", "javascript"},
		{"app.jsx", "jsx"},
		{"module.mjs", "javascript"},
		{"lib.cjs", "javascript"},
		{"Main.java", "java"},
		{"app.rb", "ruby"},
		{"main.rs", "rust"},
		{"style.css", "css"},
		{"style.scss", "scss"},
		{"index.html", "html"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"config.toml", "toml"},
		{"data.json", "json"},
		{"query.graphql", "graphql"},
		{"query.gql", "graphql"},
		{"README.md", "markdown"},
		{"script.sh", "shell"},
		{"schema.proto", "protobuf"},
		{"noextension", "unknown"},
		{"unknown.xyz", "unknown"},
	}

	for _, tc := range cases {
		info, err := p.Parse(tc.filename, []byte("hello\nworld\n"))
		if err != nil {
			t.Fatalf("Parse(%q): %v", tc.filename, err)
		}
		if info.Language != tc.want {
			t.Errorf("Language(%q) = %q, want %q", tc.filename, info.Language, tc.want)
		}
	}
}

func TestGenericParserParse_LineCount(t *testing.T) {
	p := &GenericParser{}

	cases := []struct {
		content   string
		wantLines int
	}{
		{"", 0},
		{"one line", 1},
		{"line1\nline2\n", 2},
		{"line1\nline2\nline3", 3},
		{"\n\n\n", 3},
	}

	for _, tc := range cases {
		info, err := p.Parse("test.txt", []byte(tc.content))
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if info.LineCount != tc.wantLines {
			t.Errorf("LineCount(%q) = %d, want %d", tc.content, info.LineCount, tc.wantLines)
		}
	}
}

func TestGenericParserParse_NoImportsOrExports(t *testing.T) {
	p := &GenericParser{}

	content := []byte("import os\nprint('hello')\n")
	info, err := p.Parse("script.sh", content)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if len(info.Imports) != 0 {
		t.Errorf("Imports = %v, want empty (generic parser does not analyze imports)", info.Imports)
	}
	if len(info.Exports) != 0 {
		t.Errorf("Exports = %v, want empty (generic parser does not analyze exports)", info.Exports)
	}
}
