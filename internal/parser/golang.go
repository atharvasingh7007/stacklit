package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"
)

// GoParser parses Go source files using the stdlib go/parser + go/ast.
type GoParser struct{}

// CanParse returns true for .go files that are NOT test files.
func (g *GoParser) CanParse(filename string) bool {
	if !extIs(filename, ".go") {
		return false
	}
	// Exclude _test.go files.
	return !strings.HasSuffix(strings.ToLower(filename), "_test.go")
}

// Parse extracts imports and exported names from a Go source file.
func (g *GoParser) Parse(path string, content []byte) (*FileInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, content, 0)
	if err != nil {
		return nil, err
	}

	info := &FileInfo{
		Path:      path,
		Language:  "go",
		LineCount: countLines(content),
	}

	// Extract import paths.
	for _, imp := range f.Imports {
		if imp.Path != nil {
			// Strip surrounding quotes.
			importPath := strings.Trim(imp.Path.Value, `"`)
			info.Imports = append(info.Imports, importPath)
		}
	}

	// Detect main entrypoint: package main with a main() func.
	if f.Name != nil && f.Name.Name == "main" {
		for _, decl := range f.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				if fn.Name.Name == "main" {
					info.IsEntrypoint = true
					break
				}
			}
		}
	}

	// Extract exported names (uppercase first letter).
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if isExported(d.Name.Name) {
				info.Exports = append(info.Exports, d.Name.Name)
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if isExported(s.Name.Name) {
						info.Exports = append(info.Exports, s.Name.Name)
					}
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if isExported(name.Name) {
							info.Exports = append(info.Exports, name.Name)
						}
					}
				}
			}
		}
	}

	return info, nil
}

// isExported reports whether a Go identifier is exported (starts with uppercase).
func isExported(name string) bool {
	if name == "" {
		return false
	}
	r := rune(name[0])
	return unicode.IsUpper(r)
}
