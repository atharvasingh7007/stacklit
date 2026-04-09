package parser

import (
	"regexp"
)

var (
	// ESM import: import [type] ... from 'module'
	tsImportFrom = regexp.MustCompile(`(?m)^import\s+(?:type\s+)?(?:.*?)\s+from\s+['"]([^'"]+)['"]`)

	// Side-effect import: import 'module'
	tsImportSideEffect = regexp.MustCompile(`(?m)^import\s+['"]([^'"]+)['"]`)

	// CJS require: require('module')
	tsRequire = regexp.MustCompile(`require\(['"]([^'"]+)['"]\)`)

	// Named exports: export [async] function|class|const|let|var|type|interface|enum Name
	tsExport = regexp.MustCompile(`(?m)^export\s+(?:async\s+)?(?:function|class|const|let|var|type|interface|enum)\s+(\w+)`)
)

// TypeScriptParser parses TypeScript and JavaScript files using regex.
type TypeScriptParser struct{}

// CanParse returns true for .ts, .tsx, .js, .jsx, .mjs, .cjs files.
func (t *TypeScriptParser) CanParse(filename string) bool {
	return extIs(filename, ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs")
}

// Parse extracts imports and exports from a TS/JS source file.
func (t *TypeScriptParser) Parse(path string, content []byte) (*FileInfo, error) {
	info := &FileInfo{
		Path:      path,
		Language:  tsLanguage(path),
		LineCount: countLines(content),
	}

	// Collect imports, deduplicating.
	seen := make(map[string]bool)
	addImport := func(s string) {
		if !seen[s] {
			seen[s] = true
			info.Imports = append(info.Imports, s)
		}
	}

	for _, m := range tsImportFrom.FindAllSubmatch(content, -1) {
		addImport(string(m[1]))
	}

	for _, m := range tsImportSideEffect.FindAllSubmatch(content, -1) {
		addImport(string(m[1]))
	}

	for _, m := range tsRequire.FindAllSubmatch(content, -1) {
		addImport(string(m[1]))
	}

	// Collect exports.
	seenExport := make(map[string]bool)
	for _, m := range tsExport.FindAllSubmatch(content, -1) {
		name := string(m[1])
		if !seenExport[name] {
			seenExport[name] = true
			info.Exports = append(info.Exports, name)
		}
	}

	return info, nil
}

// tsLanguage maps file extension to a language label.
func tsLanguage(filename string) string {
	switch {
	case extIs(filename, ".ts"):
		return "typescript"
	case extIs(filename, ".tsx"):
		return "tsx"
	case extIs(filename, ".js"):
		return "javascript"
	case extIs(filename, ".jsx"):
		return "jsx"
	case extIs(filename, ".mjs"):
		return "javascript"
	case extIs(filename, ".cjs"):
		return "javascript"
	default:
		return "javascript"
	}
}
