package detect

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// envPatterns matches env var accesses in Go, JS/TS, and Python source code.
var envPatterns = []*regexp.Regexp{
	// Go
	regexp.MustCompile(`os\.Getenv\("([A-Z_][A-Z0-9_]+)"\)`),
	regexp.MustCompile(`os\.LookupEnv\("([A-Z_][A-Z0-9_]+)"\)`),
	// JS/TS
	regexp.MustCompile(`process\.env\.([A-Z_][A-Z0-9_]+)`),
	regexp.MustCompile(`process\.env\["([A-Z_][A-Z0-9_]+)"\]`),
	// Python
	regexp.MustCompile(`os\.environ\["([A-Z_][A-Z0-9_]+)"\]`),
	regexp.MustCompile(`os\.getenv\("([A-Z_][A-Z0-9_]+)"\)`),
	regexp.MustCompile(`os\.environ\.get\("([A-Z_][A-Z0-9_]+)"\)`),
}

// commonEnvVars is a set of OS-level env vars to skip.
var commonEnvVars = map[string]bool{
	"HOME": true, "PATH": true, "USER": true, "SHELL": true,
	"PWD": true, "TERM": true, "LANG": true, "TMPDIR": true,
	"TMP": true, "TEMP": true, "LOGNAME": true,
}

// DetectEnvVars scans file contents for env var references and reads .env.example
// / .env.sample if present at root. Returns a deduplicated, sorted list of keys.
func DetectEnvVars(root string, contents map[string][]byte) []string {
	seen := map[string]bool{}

	// 1. Scan source code patterns across all file contents.
	for _, content := range contents {
		src := string(content)
		for _, re := range envPatterns {
			for _, match := range re.FindAllStringSubmatch(src, -1) {
				key := match[1]
				if !commonEnvVars[key] {
					seen[key] = true
				}
			}
		}
	}

	// 2. Parse .env.example or .env.sample.
	for _, name := range []string{".env.example", ".env.sample"} {
		path := filepath.Join(root, name)
		if data, err := os.ReadFile(path); err == nil {
			scanner := bufio.NewScanner(bytes.NewReader(data))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				if idx := strings.IndexByte(line, '='); idx > 0 {
					key := strings.TrimSpace(line[:idx])
					if key != "" && !commonEnvVars[key] {
						seen[key] = true
					}
				}
			}
		}
	}

	// 3. Parse docker-compose.yml environment section keys.
	if data, err := os.ReadFile(filepath.Join(root, "docker-compose.yml")); err == nil {
		parseDockerComposeEnvVars(data, seen)
	}

	// Build sorted slice.
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// parseDockerComposeEnvVars extracts KEY or KEY=value entries from the
// environment: sections of a docker-compose file (plain text scan, no YAML lib).
func parseDockerComposeEnvVars(data []byte, seen map[string]bool) {
	inEnv := false
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect "environment:" block.
		if strings.HasSuffix(trimmed, "environment:") {
			inEnv = true
			continue
		}
		// A new top-level or sibling key ends the environment block.
		if inEnv && len(line) > 0 && line[0] != ' ' && line[0] != '\t' && line[0] != '-' {
			inEnv = false
		}
		if !inEnv {
			continue
		}
		// Strip leading whitespace and list marker.
		entry := strings.TrimLeft(trimmed, "- ")
		if entry == "" || strings.HasPrefix(entry, "#") {
			continue
		}
		var key string
		if idx := strings.IndexByte(entry, '='); idx > 0 {
			key = strings.TrimSpace(entry[:idx])
		} else if idx := strings.IndexByte(entry, ':'); idx > 0 {
			key = strings.TrimSpace(entry[:idx])
		} else {
			key = strings.TrimSpace(entry)
		}
		if key != "" && !commonEnvVars[key] {
			seen[key] = true
		}
	}
}
