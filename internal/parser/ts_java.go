package parser

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
)

// extractJava extracts structured info from a Java source file using the tree-sitter AST.
func extractJava(root *gts.Node, lang *gts.Language, src []byte, path string) *FileInfo {
	info := &FileInfo{
		Language: "java",
		TypeDefs: make(map[string]string),
	}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		switch node.Type(lang) {
		case "import_declaration":
			// import java.util.List; or import static org.junit.Assert.assertEquals;
			// Drop the final class name (uppercase first letter).
			text := strings.TrimPrefix(strings.TrimSpace(nodeText(node, src)), "import ")
			text = strings.TrimPrefix(text, "static ")
			text = strings.TrimSuffix(text, ";")
			text = strings.TrimSpace(text)
			// Drop final segment (class name): split on "." and rejoin all but last.
			parts := strings.Split(text, ".")
			if len(parts) > 1 {
				// Only drop if the last segment looks like a class (uppercase or *).
				last := parts[len(parts)-1]
				if last == "*" || (len(last) > 0 && last[0] >= 'A' && last[0] <= 'Z') {
					parts = parts[:len(parts)-1]
				}
			}
			pkg := strings.Join(parts, ".")
			if pkg != "" {
				info.Imports = append(info.Imports, pkg)
			}
			return gts.WalkSkipChildren

		case "class_declaration", "interface_declaration", "enum_declaration", "record_declaration":
			// Only export public types (check modifiers).
			if !javaHasPublicModifier(node, lang, src) {
				return gts.WalkSkipChildren
			}
			nameNode := childByType(node, lang, "identifier")
			if nameNode == nil {
				return gts.WalkSkipChildren
			}
			name := nodeText(nameNode, src)
			info.Exports = append(info.Exports, name)

			// Extract public method names for TypeDefs.
			methods := javaPublicMethods(node, lang, src)
			if len(methods) > 0 {
				info.TypeDefs[name] = strings.Join(methods, ", ")
			}

			// Detect public static void main entrypoint within this class.
			if javaHasMainMethod(node, lang, src) {
				info.IsEntrypoint = true
			}
			return gts.WalkSkipChildren
		}

		return gts.WalkContinue
	})

	info.Imports = dedup(info.Imports)
	info.Exports = dedup(info.Exports)
	if len(info.TypeDefs) == 0 {
		info.TypeDefs = nil
	}
	return info
}

// javaHasPublicModifier checks if a type declaration node has a "public" modifier.
func javaHasPublicModifier(node *gts.Node, lang *gts.Language, src []byte) bool {
	modifiers := childByType(node, lang, "modifiers")
	if modifiers == nil {
		return false
	}
	text := nodeText(modifiers, src)
	return strings.Contains(text, "public")
}

// javaPublicMethods returns names of public methods declared directly in a type body.
func javaPublicMethods(typeNode *gts.Node, lang *gts.Language, src []byte) []string {
	var methods []string
	bodyNode := childByType(typeNode, lang, "class_body")
	if bodyNode == nil {
		bodyNode = childByType(typeNode, lang, "interface_body")
	}
	if bodyNode == nil {
		bodyNode = childByType(typeNode, lang, "enum_body")
	}
	if bodyNode == nil {
		return nil
	}

	for i := 0; i < bodyNode.ChildCount(); i++ {
		child := bodyNode.Child(i)
		t := child.Type(lang)
		if t != "method_declaration" && t != "interface_method_declaration" {
			continue
		}
		mods := childByType(child, lang, "modifiers")
		if mods == nil || !strings.Contains(nodeText(mods, src), "public") {
			continue
		}
		nameNode := childByType(child, lang, "identifier")
		if nameNode != nil {
			methods = append(methods, nodeText(nameNode, src))
		}
	}
	return methods
}

// javaHasMainMethod checks for "public static void main" in the direct class body.
func javaHasMainMethod(typeNode *gts.Node, lang *gts.Language, src []byte) bool {
	bodyNode := childByType(typeNode, lang, "class_body")
	if bodyNode == nil {
		return false
	}
	for i := 0; i < bodyNode.ChildCount(); i++ {
		child := bodyNode.Child(i)
		if child.Type(lang) != "method_declaration" {
			continue
		}
		text := nodeText(child, src)
		if strings.Contains(text, "public") &&
			strings.Contains(text, "static") &&
			strings.Contains(text, "void") &&
			strings.Contains(text, "main") {
			return true
		}
	}
	return false
}
