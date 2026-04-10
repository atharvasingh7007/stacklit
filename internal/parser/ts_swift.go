package parser

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
)

// extractSwift extracts imports, exports, and type definitions from a Swift source file.
func extractSwift(root *gts.Node, lang *gts.Language, src []byte, path string) *FileInfo {
	info := &FileInfo{
		Language: "swift",
		TypeDefs: make(map[string]string),
	}

	type typeEntry struct {
		name    string
		members []string
	}
	var types []typeEntry

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		switch node.Type(lang) {
		case "import_declaration":
			// import Foundation or import UIKit.UIView
			text := strings.TrimSpace(nodeText(node, src))
			text = strings.TrimPrefix(text, "import")
			text = strings.TrimSpace(text)
			if text != "" {
				info.Imports = append(info.Imports, text)
			}
			return gts.WalkSkipChildren

		case "class_declaration", "struct_declaration", "protocol_declaration", "enum_declaration":
			// Only include public/open or unqualified declarations (public by default in a module).
			if swiftHasModifier(node, lang, src, "private") || swiftHasModifier(node, lang, src, "fileprivate") {
				return gts.WalkSkipChildren
			}
			name := swiftDeclName(node, lang, src)
			if name == "" {
				return gts.WalkContinue
			}
			info.Exports = append(info.Exports, name)
			members := swiftTypeMembers(node, lang, src)
			types = append(types, typeEntry{name: name, members: members})
			return gts.WalkSkipChildren

		case "function_declaration":
			// Top-level functions only (depth 1 or 2 in the tree).
			if depth > 3 {
				return gts.WalkSkipChildren
			}
			if swiftHasModifier(node, lang, src, "private") || swiftHasModifier(node, lang, src, "fileprivate") {
				return gts.WalkSkipChildren
			}
			sig := swiftFunctionSig(node, lang, src)
			if sig != "" {
				info.Exports = append(info.Exports, sig)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	info.Imports = dedup(info.Imports)
	info.Exports = dedup(info.Exports)

	for _, te := range types {
		if len(te.members) > 0 {
			info.TypeDefs[te.name] = strings.Join(te.members, "; ")
		} else {
			info.TypeDefs[te.name] = te.name
		}
	}

	if len(info.TypeDefs) == 0 {
		info.TypeDefs = nil
	}

	return info
}

// swiftHasModifier checks if a node has a specific access control modifier.
func swiftHasModifier(node *gts.Node, lang *gts.Language, src []byte, mod string) bool {
	// Modifiers may appear as direct children with type "modifier" or "access_level_modifier".
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		t := child.Type(lang)
		if t == "modifier" || t == "access_level_modifier" || t == "attribute" {
			if strings.Contains(nodeText(child, src), mod) {
				return true
			}
		}
	}
	return false
}

// swiftDeclName extracts the type name from a Swift declaration node.
func swiftDeclName(node *gts.Node, lang *gts.Language, src []byte) string {
	id := childByType(node, lang, "type_identifier")
	if id != nil {
		return nodeText(id, src)
	}
	id = childByType(node, lang, "simple_identifier")
	if id != nil {
		return nodeText(id, src)
	}
	return ""
}

// swiftTypeMembers collects property and function names from a type body.
func swiftTypeMembers(node *gts.Node, lang *gts.Language, src []byte) []string {
	var members []string
	body := childByType(node, lang, "class_body")
	if body == nil {
		body = childByType(node, lang, "struct_body")
	}
	if body == nil {
		body = childByType(node, lang, "protocol_body")
	}
	if body == nil {
		body = childByType(node, lang, "enum_class_body")
	}
	if body == nil {
		return members
	}

	for i := 0; i < body.ChildCount(); i++ {
		child := body.Child(i)
		switch child.Type(lang) {
		case "function_declaration":
			if swiftHasModifier(child, lang, src, "private") || swiftHasModifier(child, lang, src, "fileprivate") {
				continue
			}
			sig := swiftFunctionSig(child, lang, src)
			if sig != "" {
				members = append(members, sig)
			}
		case "property_declaration", "variable_declaration":
			if swiftHasModifier(child, lang, src, "private") || swiftHasModifier(child, lang, src, "fileprivate") {
				continue
			}
			name := swiftPropertyName(child, lang, src)
			if name != "" {
				members = append(members, name)
			}
		}
	}
	return members
}

// swiftFunctionSig builds a brief function signature string.
func swiftFunctionSig(node *gts.Node, lang *gts.Language, src []byte) string {
	var parts []string
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		t := child.Type(lang)
		switch t {
		case "modifier", "access_level_modifier", "attribute", "function_body", "code_block", "{", "}":
			continue
		case "simple_identifier", "type_identifier":
			parts = append(parts, nodeText(child, src))
		case "parameter_clause", "function_value_parameters":
			parts = append(parts, nodeText(child, src))
		case "type_annotation":
			parts = append(parts, "->"+nodeText(child, src))
		case "func":
			parts = append(parts, "func")
		}
	}
	sig := strings.Join(parts, " ")
	return truncSig(sig, 80)
}

// swiftPropertyName extracts the name from a property declaration.
func swiftPropertyName(node *gts.Node, lang *gts.Language, src []byte) string {
	id := childByType(node, lang, "simple_identifier")
	if id != nil {
		return nodeText(id, src)
	}
	// pattern binding may nest the identifier.
	pb := childByType(node, lang, "pattern_binding")
	if pb != nil {
		id = childByType(pb, lang, "simple_identifier")
		if id != nil {
			return nodeText(id, src)
		}
	}
	return ""
}
