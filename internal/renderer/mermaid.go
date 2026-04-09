package renderer

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/GLINCKER/stacklit/internal/schema"
)

// langColors maps language names to their brand colors.
var langColors = map[string][2]string{
	"go":         {"#00ADD8", "white"},
	"typescript": {"#3178C6", "white"},
	"javascript": {"#F7DF1E", "black"},
	"python":     {"#3776AB", "white"},
	"rust":       {"#DEA584", "black"},
	"java":       {"#ED8B00", "white"},
	"ruby":       {"#CC342D", "white"},
	"c":          {"#A8B9CC", "black"},
	"cpp":        {"#00599C", "white"},
	"csharp":     {"#239120", "white"},
	"swift":      {"#FA7343", "white"},
	"kotlin":     {"#7F52FF", "white"},
	"php":        {"#777BB4", "white"},
	"scala":      {"#DC322F", "white"},
	"elixir":     {"#6E4A7E", "white"},
}

// sanitizeMermaidID converts a module path to a valid Mermaid node ID.
// e.g. "src/auth" → "src_auth", "@test/api" → "test_api"
func sanitizeMermaidID(name string) string {
	var sb strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	result := strings.Trim(sb.String(), "_")
	// Collapse consecutive underscores
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	return result
}

// truncate shortens s to at most n runes, appending "…" if truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

// WriteMermaid generates a Mermaid graph LR diagram from idx and writes it to path.
func WriteMermaid(idx *schema.Index, path string) error {
	var sb strings.Builder

	sb.WriteString("graph LR\n")

	// Collect which languages are actually used by modules.
	usedLangs := map[string]bool{}
	for _, mod := range idx.Modules {
		_ = mod // language info not on ModuleInfo; use Tech.Languages from idx
	}
	for lang := range idx.Tech.Languages {
		lang = strings.ToLower(lang)
		if _, ok := langColors[lang]; ok {
			usedLangs[lang] = true
		}
	}

	// Write classDef entries for used languages.
	for lang := range usedLangs {
		colors := langColors[lang]
		sb.WriteString(fmt.Sprintf("  classDef %s fill:%s,color:%s,stroke:%s\n",
			lang, colors[0], colors[1], colors[0]))
	}

	// Determine primary language for node class assignment.
	primaryLang := strings.ToLower(idx.Tech.PrimaryLanguage)

	// Write node definitions.
	for name, mod := range idx.Modules {
		id := sanitizeMermaidID(name)
		label := truncate(mod.Purpose, 40)
		nodeClass := primaryLang
		if _, ok := langColors[nodeClass]; !ok {
			nodeClass = ""
		}
		if nodeClass != "" {
			sb.WriteString(fmt.Sprintf("  %s[\"%s\\n%s\"]:::%s\n", id, name, label, nodeClass))
		} else {
			sb.WriteString(fmt.Sprintf("  %s[\"%s\\n%s\"]\n", id, name, label))
		}
	}

	// Write edges.
	for _, edge := range idx.Dependencies.Edges {
		from := sanitizeMermaidID(edge[0])
		to := sanitizeMermaidID(edge[1])
		sb.WriteString(fmt.Sprintf("  %s --> %s\n", from, to))
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}
