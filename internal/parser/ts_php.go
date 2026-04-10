package parser

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
)

// extractPHP extracts imports, exports, and type definitions from a PHP source file.
func extractPHP(root *gts.Node, lang *gts.Language, src []byte, path string) *FileInfo {
	info := &FileInfo{
		Language: "php",
		TypeDefs: make(map[string]string),
	}

	type typeEntry struct {
		name    string
		methods []string
	}
	var types []typeEntry

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		switch node.Type(lang) {
		case "namespace_use_declaration":
			// use Foo\Bar\Baz; or use Foo\Bar\Baz as Alias;
			text := strings.TrimSpace(nodeText(node, src))
			// Strip leading "use " and trailing ";"
			text = strings.TrimPrefix(text, "use")
			text = strings.TrimSuffix(text, ";")
			text = strings.TrimSpace(text)
			// Drop "as Alias" suffix.
			if idx := strings.LastIndex(text, " as "); idx >= 0 {
				text = strings.TrimSpace(text[:idx])
			}
			if text != "" {
				info.Imports = append(info.Imports, text)
			}
			return gts.WalkSkipChildren

		case "class_declaration", "interface_declaration", "trait_declaration":
			name := phpTypeName(node, lang, src)
			if name == "" {
				return gts.WalkContinue
			}
			info.Exports = append(info.Exports, name)
			methods := phpPublicMethods(node, lang, src)
			types = append(types, typeEntry{name: name, methods: methods})
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	info.Imports = dedup(info.Imports)
	info.Exports = dedup(info.Exports)

	for _, te := range types {
		if len(te.methods) > 0 {
			info.TypeDefs[te.name] = strings.Join(te.methods, "; ")
		} else {
			info.TypeDefs[te.name] = te.name
		}
	}

	if len(info.TypeDefs) == 0 {
		info.TypeDefs = nil
	}

	return info
}

// phpTypeName extracts the name identifier from a PHP type declaration node.
func phpTypeName(node *gts.Node, lang *gts.Language, src []byte) string {
	id := childByType(node, lang, "name")
	if id != nil {
		return nodeText(id, src)
	}
	id = childByType(node, lang, "identifier")
	if id != nil {
		return nodeText(id, src)
	}
	return ""
}

// phpPublicMethods collects public method signatures from a PHP class/interface/trait body.
func phpPublicMethods(node *gts.Node, lang *gts.Language, src []byte) []string {
	var methods []string
	body := childByType(node, lang, "declaration_list")
	if body == nil {
		return methods
	}
	for i := 0; i < body.ChildCount(); i++ {
		child := body.Child(i)
		if child.Type(lang) == "method_declaration" {
			if phpHasPublicModifier(child, lang, src) {
				sig := phpMethodSig(child, lang, src)
				if sig != "" {
					methods = append(methods, sig)
				}
			}
		}
	}
	return methods
}

// phpHasPublicModifier checks for a "public" visibility modifier.
func phpHasPublicModifier(node *gts.Node, lang *gts.Language, src []byte) bool {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		t := child.Type(lang)
		if (t == "visibility_modifier" || t == "modifier") && nodeText(child, src) == "public" {
			return true
		}
	}
	return false
}

// phpMethodSig builds a short method signature string.
func phpMethodSig(node *gts.Node, lang *gts.Language, src []byte) string {
	// Collect function keyword, name, and parameter list.
	var parts []string
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		t := child.Type(lang)
		switch t {
		case "visibility_modifier", "modifier", "compound_statement", ":", "{", "}":
			continue
		case "name", "identifier", "formal_parameters", "union_type", "named_type", "primitive_type":
			parts = append(parts, nodeText(child, src))
		case "function":
			parts = append(parts, "function")
		}
	}
	sig := strings.Join(parts, " ")
	return truncSig(sig, 80)
}
