package parser

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
)

// extractC extracts structured info from C, C++, or Objective-C source files
// using the tree-sitter AST. The Language field is set based on the file extension.
func extractC(root *gts.Node, lang *gts.Language, src []byte, path string) *FileInfo {
	info := &FileInfo{
		Language: cLanguage(path),
		TypeDefs: make(map[string]string),
	}

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		switch node.Type(lang) {
		case "preproc_include":
			// #include <stdio.h> or #include "myheader.h"
			// The path is in a string_literal or system_lib_string child.
			for i := 0; i < node.ChildCount(); i++ {
				c := node.Child(i)
				t := c.Type(lang)
				if t == "string_literal" || t == "system_lib_string" {
					text := nodeText(c, src)
					// Strip enclosing <> or ""
					text = strings.Trim(text, "<>\"")
					if text != "" {
						info.Imports = append(info.Imports, text)
					}
					break
				}
			}
			return gts.WalkSkipChildren

		case "function_definition":
			// Only file-scope functions (depth == 1).
			if depth != 1 {
				return gts.WalkSkipChildren
			}
			declNode := childByType(node, lang, "function_declarator")
			if declNode == nil {
				// Try pointer_declarator wrapping a function_declarator.
				ptrNode := childByType(node, lang, "pointer_declarator")
				if ptrNode != nil {
					declNode = childByType(ptrNode, lang, "function_declarator")
				}
			}
			if declNode != nil {
				nameNode := childByType(declNode, lang, "identifier")
				if nameNode != nil {
					name := nodeText(nameNode, src)
					params := ""
					paramsNode := childByType(declNode, lang, "parameter_list")
					if paramsNode != nil {
						params = truncSig(nodeText(paramsNode, src), 60)
					}
					info.Exports = append(info.Exports, name+params)
					if cIsMain(name, params) {
						info.IsEntrypoint = true
					}
				}
			}
			return gts.WalkSkipChildren

		case "declaration":
			// File-scope declarations: function declarations (prototypes), typedefs.
			if depth != 1 {
				return gts.WalkSkipChildren
			}
			// Look for function_declarator to catch prototypes.
			cExtractDeclaration(node, lang, src, info)
			return gts.WalkSkipChildren

		case "struct_specifier", "union_specifier", "enum_specifier":
			// Only top-level struct/union/enum declarations.
			if depth > 2 {
				return gts.WalkSkipChildren
			}
			nameNode := childByType(node, lang, "type_identifier")
			if nameNode == nil {
				return gts.WalkSkipChildren
			}
			name := nodeText(nameNode, src)
			if name == "" {
				return gts.WalkSkipChildren
			}
			info.Exports = append(info.Exports, node.Type(lang)+" "+name)

			// Extract field names for structs and unions.
			if node.Type(lang) == "struct_specifier" || node.Type(lang) == "union_specifier" {
				bodyNode := childByType(node, lang, "field_declaration_list")
				if bodyNode != nil {
					var fields []string
					for i := 0; i < bodyNode.ChildCount(); i++ {
						field := bodyNode.Child(i)
						if field.Type(lang) == "field_declaration" {
							fn := cFieldName(field, lang, src)
							if fn != "" {
								fields = append(fields, fn)
							}
						}
					}
					if len(fields) > 0 {
						info.TypeDefs[name] = strings.Join(fields, ", ")
					}
				}
			}
			return gts.WalkSkipChildren

		case "type_definition":
			// typedef declarations.
			if depth != 1 {
				return gts.WalkSkipChildren
			}
			// The aliased name is typically the last type_identifier child.
			var lastName string
			for i := 0; i < node.ChildCount(); i++ {
				c := node.Child(i)
				if c.Type(lang) == "type_identifier" {
					lastName = nodeText(c, src)
				}
			}
			if lastName != "" {
				info.Exports = append(info.Exports, "typedef "+lastName)
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

// cLanguage maps file extension to a language label.
func cLanguage(filename string) string {
	switch {
	case extIs(filename, ".cpp", ".cxx", ".cc", ".hpp", ".hxx"):
		return "cpp"
	case extIs(filename, ".m", ".mm"):
		return "objective-c"
	default:
		return "c"
	}
}

// cIsMain checks whether a function name/params represents int main(...).
func cIsMain(name, params string) bool {
	return name == "main"
}

// cExtractDeclaration handles a top-level declaration node, looking for
// function prototypes and other named declarations to add to exports.
func cExtractDeclaration(node *gts.Node, lang *gts.Language, src []byte, info *FileInfo) {
	// Walk direct children for function_declarator or pointer_declarator.
	for i := 0; i < node.ChildCount(); i++ {
		c := node.Child(i)
		switch c.Type(lang) {
		case "function_declarator":
			nameNode := childByType(c, lang, "identifier")
			if nameNode != nil {
				name := nodeText(nameNode, src)
				params := ""
				paramsNode := childByType(c, lang, "parameter_list")
				if paramsNode != nil {
					params = truncSig(nodeText(paramsNode, src), 60)
				}
				info.Exports = append(info.Exports, name+params)
			}
		case "pointer_declarator":
			inner := childByType(c, lang, "function_declarator")
			if inner != nil {
				nameNode := childByType(inner, lang, "identifier")
				if nameNode != nil {
					name := nodeText(nameNode, src)
					params := ""
					paramsNode := childByType(inner, lang, "parameter_list")
					if paramsNode != nil {
						params = truncSig(nodeText(paramsNode, src), 60)
					}
					info.Exports = append(info.Exports, name+params)
				}
			}
		}
	}
}

// cFieldName extracts the field name from a field_declaration node.
func cFieldName(field *gts.Node, lang *gts.Language, src []byte) string {
	// Check for field_identifier directly.
	fn := childByType(field, lang, "field_identifier")
	if fn != nil {
		return nodeText(fn, src)
	}
	// May be wrapped in a pointer_declarator.
	ptrNode := childByType(field, lang, "pointer_declarator")
	if ptrNode != nil {
		fn = childByType(ptrNode, lang, "field_identifier")
		if fn != nil {
			return nodeText(fn, src)
		}
	}
	return ""
}
