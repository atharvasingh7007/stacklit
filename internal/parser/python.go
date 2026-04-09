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

	// module-level function definitions — capture full signature up to colon
	pyDef    = regexp.MustCompile(`(?m)^def\s+(\w+)\s*\(`)
	pyDefSig = regexp.MustCompile(`(?m)^(def\s+\w+\s*\([^)]*\)(?:\s*->\s*\S+)?)`)

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

	// Build name→full-sig map for defs.
	defNames := pyDef.FindAllSubmatch(content, -1)
	defSigs := pyDefSig.FindAllSubmatch(content, -1)
	defSigMap := make(map[string]string, len(defSigs))
	for i, ds := range defSigs {
		if i < len(defNames) {
			name := string(defNames[i][1])
			sig := strings.TrimSpace(string(ds[1]))
			if len(sig) > 80 {
				sig = sig[:80]
			}
			defSigMap[name] = sig
		}
	}

	addExport := func(name, sig string) {
		if !strings.HasPrefix(name, "_") && !seenExport[name] {
			seenExport[name] = true
			info.Exports = append(info.Exports, sig)
		}
	}

	for _, m := range defNames {
		name := string(m[1])
		sig, ok := defSigMap[name]
		if !ok || sig == "" {
			sig = name
		}
		addExport(name, sig)
	}

	for _, m := range pyClass.FindAllSubmatch(content, -1) {
		name := string(m[1])
		addExport(name, "class "+name)
	}

	// Detect entrypoint.
	if pyMain.Match(content) {
		info.IsEntrypoint = true
	}

	return info, nil
}
