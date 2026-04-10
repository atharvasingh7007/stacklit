package parser

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
)

// extractPython extracts imports, exports, type definitions, and entrypoint
// information from a Python source file using the tree-sitter AST.
func extractPython(root *gts.Node, lang *gts.Language, src []byte, path string) *FileInfo {
	info := &FileInfo{
		Language: "python",
	}

	// Walk the AST. Module-level nodes are direct children of the root (depth 1).
	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		nodeType := node.Type(lang)

		switch nodeType {
		case "import_statement":
			// import os, sys  OR  import os.path
			// Children include "import" keyword and dotted_name / aliased_import nodes.
			extractPyImportNames(node, lang, src, info)
			return gts.WalkSkipChildren

		case "import_from_statement":
			// from os import path  OR  from .models import ...
			// The module name is in the "module_name" field.
			mod := node.ChildByFieldName("module_name", lang)
			if mod != nil {
				name := nodeText(mod, src)
				if name != "" {
					info.Imports = dedup(append(info.Imports, name))
				}
			}
			return gts.WalkSkipChildren

		case "function_definition":
			// Only collect module-level functions (depth 1).
			if depth != 1 {
				return gts.WalkSkipChildren
			}
			nameNode := node.ChildByFieldName("name", lang)
			if nameNode == nil {
				return gts.WalkSkipChildren
			}
			funcName := nodeText(nameNode, src)
			if strings.HasPrefix(funcName, "_") {
				return gts.WalkSkipChildren
			}
			sig := buildPyFuncSig(node, lang, src, funcName)
			info.Exports = append(info.Exports, sig)
			return gts.WalkSkipChildren

		case "class_definition":
			// Only collect module-level classes (depth 1).
			if depth != 1 {
				return gts.WalkSkipChildren
			}
			nameNode := node.ChildByFieldName("name", lang)
			if nameNode == nil {
				return gts.WalkSkipChildren
			}
			className := nodeText(nameNode, src)
			if strings.HasPrefix(className, "_") {
				return gts.WalkSkipChildren
			}
			info.Exports = append(info.Exports, "class "+className)

			// Extract public methods from the class body for TypeDefs.
			methods := extractPyClassMethods(node, lang, src)
			if len(methods) > 0 {
				if info.TypeDefs == nil {
					info.TypeDefs = make(map[string]string)
				}
				info.TypeDefs[className] = strings.Join(methods, ", ")
			}
			return gts.WalkSkipChildren

		case "if_statement":
			// Detect: if __name__ == '__main__':
			if depth == 1 && isPyMainGuard(node, lang, src) {
				info.IsEntrypoint = true
			}
			return gts.WalkSkipChildren
		}

		return gts.WalkContinue
	})

	return info
}

// extractPyImportNames handles import_statement nodes and appends module names
// to info.Imports. Handles: import os, import os.path, import os as o.
func extractPyImportNames(node *gts.Node, lang *gts.Language, src []byte, info *FileInfo) {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		childType := child.Type(lang)
		switch childType {
		case "dotted_name":
			name := nodeText(child, src)
			if name != "" {
				info.Imports = dedup(append(info.Imports, name))
			}
		case "aliased_import":
			// aliased_import has a "name" field for the module being imported.
			nameNode := child.ChildByFieldName("name", lang)
			if nameNode != nil {
				name := nodeText(nameNode, src)
				if name != "" {
					info.Imports = dedup(append(info.Imports, name))
				}
			}
		}
	}
}

// buildPyFuncSig builds a function signature string like "func_name(params) -> return_type".
func buildPyFuncSig(node *gts.Node, lang *gts.Language, src []byte, funcName string) string {
	paramsNode := node.ChildByFieldName("parameters", lang)
	returnNode := node.ChildByFieldName("return_type", lang)

	params := ""
	if paramsNode != nil {
		params = nodeText(paramsNode, src)
	}

	sig := funcName + params
	if returnNode != nil {
		ret := nodeText(returnNode, src)
		if ret != "" {
			sig += " -> " + ret
		}
	}

	return truncSig(sig, 120)
}

// extractPyClassMethods returns public method names from a class_definition node.
func extractPyClassMethods(classNode *gts.Node, lang *gts.Language, src []byte) []string {
	bodyNode := classNode.ChildByFieldName("body", lang)
	if bodyNode == nil {
		return nil
	}

	var methods []string
	for i := 0; i < bodyNode.ChildCount(); i++ {
		child := bodyNode.Child(i)
		if child == nil {
			continue
		}
		if child.Type(lang) != "function_definition" {
			continue
		}
		nameNode := child.ChildByFieldName("name", lang)
		if nameNode == nil {
			continue
		}
		methodName := nodeText(nameNode, src)
		if !strings.HasPrefix(methodName, "_") {
			methods = append(methods, methodName)
		}
	}
	return methods
}

// isPyMainGuard returns true when an if_statement matches: if __name__ == '__main__'
func isPyMainGuard(node *gts.Node, lang *gts.Language, src []byte) bool {
	// The condition of the if statement is in the "condition" field.
	condNode := node.ChildByFieldName("condition", lang)
	if condNode == nil {
		return false
	}
	condText := nodeText(condNode, src)
	return strings.Contains(condText, "__name__") && strings.Contains(condText, "__main__")
}
