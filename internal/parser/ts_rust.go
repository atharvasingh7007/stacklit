package parser

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
)

// extractRust extracts structured info from a Rust source file using the tree-sitter AST.
func extractRust(root *gts.Node, lang *gts.Language, src []byte, path string) *FileInfo {
	info := &FileInfo{
		Language: "rust",
		TypeDefs: make(map[string]string),
	}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		switch node.Type(lang) {
		case "use_declaration":
			// use std::collections::HashMap; → take first 2 path segments
			text := strings.TrimPrefix(strings.TrimSpace(nodeText(node, src)), "use ")
			text = strings.TrimSuffix(text, ";")
			// strip braces e.g. use std::{a, b}
			if idx := strings.IndexAny(text, "{"); idx != -1 {
				text = text[:idx]
			}
			text = strings.Trim(text, ":")
			parts := strings.SplitN(text, "::", 3)
			var seg string
			if len(parts) >= 2 {
				seg = parts[0] + "::" + parts[1]
			} else {
				seg = parts[0]
			}
			seg = strings.TrimSpace(seg)
			if seg != "" {
				info.Imports = append(info.Imports, seg)
			}
			return gts.WalkSkipChildren

		case "mod_item":
			// mod submodule; or mod submodule { ... }
			nameNode := childByType(node, lang, "identifier")
			if nameNode != nil {
				name := nodeText(nameNode, src)
				if name != "" {
					info.Imports = append(info.Imports, name)
				}
			}
			return gts.WalkSkipChildren

		case "extern_crate_declaration":
			// extern crate serde;
			nameNode := childByType(node, lang, "identifier")
			if nameNode != nil {
				name := nodeText(nameNode, src)
				if name != "" {
					info.Imports = append(info.Imports, name)
				}
			}
			return gts.WalkSkipChildren

		case "function_item":
			// Check for pub visibility and extract signature
			visNode := childByType(node, lang, "visibility_modifier")
			nameNode := childByType(node, lang, "identifier")
			if nameNode != nil {
				name := nodeText(nameNode, src)
				// Detect fn main() entrypoint
				if name == "main" {
					info.IsEntrypoint = true
				}
				// Only export pub items
				if visNode != nil && strings.Contains(nodeText(visNode, src), "pub") {
					params := ""
					paramsNode := childByType(node, lang, "parameters")
					if paramsNode != nil {
						params = nodeText(paramsNode, src)
					}
					ret := ""
					retNode := childByType(node, lang, "type_identifier")
					// Look for return type after ->
					for i := 0; i < node.ChildCount(); i++ {
						c := node.Child(i)
						if c.Type(lang) == "->" {
							// next sibling is the return type
							if i+1 < node.ChildCount() {
								ret = nodeText(node.Child(i+1), src)
							}
							break
						}
					}
					_ = retNode
					sig := name + truncSig(params, 60)
					if ret != "" {
						sig += " -> " + ret
					}
					info.Exports = append(info.Exports, sig)
				}
			}
			return gts.WalkSkipChildren

		case "struct_item":
			visNode := childByType(node, lang, "visibility_modifier")
			if visNode == nil || !strings.Contains(nodeText(visNode, src), "pub") {
				return gts.WalkSkipChildren
			}
			nameNode := childByType(node, lang, "type_identifier")
			if nameNode == nil {
				return gts.WalkSkipChildren
			}
			name := nodeText(nameNode, src)
			info.Exports = append(info.Exports, "type "+name)

			// Extract field names for TypeDefs
			var fields []string
			fieldsNode := childByType(node, lang, "field_declaration_list")
			if fieldsNode != nil {
				for i := 0; i < fieldsNode.ChildCount(); i++ {
					field := fieldsNode.Child(i)
					if field.Type(lang) == "field_declaration" {
						fn := childByType(field, lang, "field_identifier")
						if fn != nil {
							fields = append(fields, nodeText(fn, src))
						}
					}
				}
			}
			if len(fields) > 0 {
				info.TypeDefs[name] = strings.Join(fields, ", ")
			}
			return gts.WalkSkipChildren

		case "enum_item":
			visNode := childByType(node, lang, "visibility_modifier")
			if visNode == nil || !strings.Contains(nodeText(visNode, src), "pub") {
				return gts.WalkSkipChildren
			}
			nameNode := childByType(node, lang, "type_identifier")
			if nameNode != nil {
				info.Exports = append(info.Exports, "type "+nodeText(nameNode, src))
			}
			return gts.WalkSkipChildren

		case "trait_item":
			visNode := childByType(node, lang, "visibility_modifier")
			if visNode == nil || !strings.Contains(nodeText(visNode, src), "pub") {
				return gts.WalkSkipChildren
			}
			nameNode := childByType(node, lang, "type_identifier")
			if nameNode == nil {
				return gts.WalkSkipChildren
			}
			name := nodeText(nameNode, src)
			info.Exports = append(info.Exports, "type "+name)

			// Extract method names for TypeDefs
			var methods []string
			bodyNode := childByType(node, lang, "declaration_list")
			if bodyNode != nil {
				for i := 0; i < bodyNode.ChildCount(); i++ {
					item := bodyNode.Child(i)
					if item.Type(lang) == "function_signature_item" || item.Type(lang) == "function_item" {
						mn := childByType(item, lang, "identifier")
						if mn != nil {
							methods = append(methods, nodeText(mn, src))
						}
					}
				}
			}
			if len(methods) > 0 {
				info.TypeDefs[name] = strings.Join(methods, ", ")
			}
			return gts.WalkSkipChildren

		case "type_item":
			visNode := childByType(node, lang, "visibility_modifier")
			if visNode == nil || !strings.Contains(nodeText(visNode, src), "pub") {
				return gts.WalkSkipChildren
			}
			nameNode := childByType(node, lang, "type_identifier")
			if nameNode != nil {
				info.Exports = append(info.Exports, "type "+nodeText(nameNode, src))
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
