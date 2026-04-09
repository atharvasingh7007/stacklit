package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"
)

// formatFuncSignature returns a concise signature string for a Go function declaration.
func formatFuncSignature(fn *ast.FuncDecl) string {
	var b strings.Builder
	b.WriteString(fn.Name.Name)

	// Parameters
	b.WriteString("(")
	if fn.Type.Params != nil {
		params := []string{}
		for _, field := range fn.Type.Params.List {
			typeStr := formatType(field.Type)
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					params = append(params, name.Name+" "+typeStr)
				}
			} else {
				params = append(params, typeStr)
			}
		}
		b.WriteString(strings.Join(params, ", "))
	}
	b.WriteString(")")

	// Return types
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		results := []string{}
		for _, field := range fn.Type.Results.List {
			results = append(results, formatType(field.Type))
		}
		if len(results) == 1 {
			b.WriteString(" " + results[0])
		} else {
			b.WriteString(" (" + strings.Join(results, ", ") + ")")
		}
	}

	sig := b.String()
	if len(sig) > 80 {
		sig = sig[:80]
	}
	return sig
}

// formatType returns a string representation of an AST type expression.
func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.ArrayType:
		return "[]" + formatType(t.Elt)
	case *ast.MapType:
		return "map[" + formatType(t.Key) + "]" + formatType(t.Value)
	case *ast.SelectorExpr:
		return formatType(t.X) + "." + t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + formatType(t.Elt)
	default:
		return "any"
	}
}

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

	// Extract exported names/signatures (uppercase first letter).
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if isExported(d.Name.Name) {
				info.Exports = append(info.Exports, formatFuncSignature(d))
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if isExported(s.Name.Name) {
						info.Exports = append(info.Exports, "type "+s.Name.Name)
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
