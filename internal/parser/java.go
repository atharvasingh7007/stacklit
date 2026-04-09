package parser

import (
	"regexp"
)

var (
	// import java.util.List; or import static org.junit.Assert.assertEquals;
	// captures the package part, dropping the final class name (uppercase first letter)
	javaImport = regexp.MustCompile(`(?m)^import\s+(?:static\s+)?([a-zA-Z_][a-zA-Z0-9_.]*)\.[A-Z]\w*;`)

	// public [abstract|final] class/interface/enum/record Name
	javaType = regexp.MustCompile(`(?m)^public\s+(?:abstract\s+|final\s+)?(?:class|interface|enum|record)\s+(\w+)`)

	// public static void main entrypoint
	javaMain = regexp.MustCompile(`public\s+static\s+void\s+main`)
)

// JavaParser parses Java source files using regex.
type JavaParser struct{}

// CanParse returns true for .java files.
func (j *JavaParser) CanParse(filename string) bool {
	return extIs(filename, ".java")
}

// Parse extracts imports, exports, and entrypoint from a Java file.
func (j *JavaParser) Parse(path string, content []byte) (*FileInfo, error) {
	info := &FileInfo{
		Path:      path,
		Language:  "java",
		LineCount: countLines(content),
	}

	seen := make(map[string]bool)
	addImport := func(s string) {
		if !seen[s] {
			seen[s] = true
			info.Imports = append(info.Imports, s)
		}
	}

	for _, m := range javaImport.FindAllSubmatch(content, -1) {
		addImport(string(m[1]))
	}

	// Collect public type exports
	seenExport := make(map[string]bool)
	for _, m := range javaType.FindAllSubmatch(content, -1) {
		name := string(m[1])
		if !seenExport[name] {
			seenExport[name] = true
			info.Exports = append(info.Exports, name)
		}
	}

	// Detect entrypoint
	if javaMain.Match(content) {
		info.IsEntrypoint = true
	}

	return info, nil
}
