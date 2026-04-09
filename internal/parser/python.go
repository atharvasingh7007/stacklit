package parser

import (
	"regexp"
	"strings"
)

var (
	// import X or import X.Y.Z
	pyImport = regexp.MustCompile(`(?m)^import\s+(\S+)`)

	// from X import ...
	pyFromImport = regexp.MustCompile(`(?m)^from\s+(\S+)\s+import`)

	// module-level function definitions
	pyDef = regexp.MustCompile(`(?m)^def\s+(\w+)\s*\(`)

	// module-level class definitions
	pyClass = regexp.MustCompile(`(?m)^class\s+(\w+)`)

	// __main__ entrypoint detection
	pyMain = regexp.MustCompile(`(?m)^if\s+__name__\s*==\s*['"]__main__['"]`)
)

// PythonParser parses Python source files using regex.
type PythonParser struct{}

// CanParse returns true for .py files.
func (p *PythonParser) CanParse(filename string) bool {
	return extIs(filename, ".py")
}

// Parse extracts imports, exports (public defs/classes), and entrypoint from a Python file.
func (p *PythonParser) Parse(path string, content []byte) (*FileInfo, error) {
	info := &FileInfo{
		Path:      path,
		Language:  "python",
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

	for _, m := range pyImport.FindAllSubmatch(content, -1) {
		addImport(string(m[1]))
	}

	for _, m := range pyFromImport.FindAllSubmatch(content, -1) {
		addImport(string(m[1]))
	}

	// Collect exports: module-level defs and classes that don't start with _.
	seenExport := make(map[string]bool)
	addExport := func(name string) {
		if !strings.HasPrefix(name, "_") && !seenExport[name] {
			seenExport[name] = true
			info.Exports = append(info.Exports, name)
		}
	}

	for _, m := range pyDef.FindAllSubmatch(content, -1) {
		addExport(string(m[1]))
	}

	for _, m := range pyClass.FindAllSubmatch(content, -1) {
		addExport(string(m[1]))
	}

	// Detect entrypoint.
	if pyMain.Match(content) {
		info.IsEntrypoint = true
	}

	return info, nil
}
