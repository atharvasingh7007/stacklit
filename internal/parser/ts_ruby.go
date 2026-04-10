package parser

import (
	"strings"

	gts "github.com/odvcencio/gotreesitter"
)

// extractRuby extracts imports, exports, and type definitions from a Ruby source file.
func extractRuby(root *gts.Node, lang *gts.Language, src []byte, path string) *FileInfo {
	info := &FileInfo{
		Language: "ruby",
		TypeDefs: make(map[string]string),
	}

	// Track current class/module context for method association.
	// We do a two-pass approach: first collect all classes and top-level methods,
	// then populate TypeDefs.
	type classEntry struct {
		name    string
		methods []string
	}
	var classes []classEntry
	var topMethods []string

	gts.Walk(root, func(node *gts.Node, depth int) gts.WalkAction {
		switch node.Type(lang) {
		case "call":
			// require / require_relative calls.
			// Tree-sitter Ruby: call node has method child and arguments child.
			method := childByType(node, lang, "identifier")
			if method == nil {
				return gts.WalkSkipChildren
			}
			methodName := nodeText(method, src)
			if methodName != "require" && methodName != "require_relative" {
				return gts.WalkSkipChildren
			}
			// Find argument: string node.
			args := childByType(node, lang, "argument_list")
			if args == nil {
				return gts.WalkSkipChildren
			}
			str := childByType(args, lang, "string")
			if str == nil {
				// try string_content directly in args
				str = childByType(args, lang, "string_literal")
			}
			if str != nil {
				val := rubyStringValue(str, lang, src)
				if val != "" {
					info.Imports = append(info.Imports, val)
				}
			}
			return gts.WalkSkipChildren

		case "class", "module":
			// Collect class/module name and their methods.
			name := rubyDefName(node, lang, src)
			if name == "" {
				return gts.WalkContinue
			}
			info.Exports = append(info.Exports, name)
			methods := rubyClassMethods(node, lang, src)
			classes = append(classes, classEntry{name: name, methods: methods})
			// Skip into class body is handled by WalkContinue but we don't
			// want to double-count nested classes. Use WalkSkipChildren and
			// recurse manually only into the body for methods (already done above).
			return gts.WalkSkipChildren

		case "method", "singleton_method":
			// Top-level method (not inside a class).
			name := rubyMethodName(node, lang, src)
			if name != "" && !strings.HasPrefix(name, "_") {
				info.Exports = append(info.Exports, name)
				topMethods = append(topMethods, name)
			}
			return gts.WalkSkipChildren
		}
		return gts.WalkContinue
	})

	info.Imports = dedup(info.Imports)
	info.Exports = dedup(info.Exports)

	for _, ce := range classes {
		if len(ce.methods) > 0 {
			info.TypeDefs[ce.name] = strings.Join(ce.methods, "; ")
		} else {
			info.TypeDefs[ce.name] = ce.name
		}
	}

	if len(info.TypeDefs) == 0 {
		info.TypeDefs = nil
	}

	return info
}

// rubyStringValue extracts the string content from a Ruby string node.
func rubyStringValue(node *gts.Node, lang *gts.Language, src []byte) string {
	// Try string_content child first.
	sc := childByType(node, lang, "string_content")
	if sc != nil {
		return nodeText(sc, src)
	}
	// Fall back to stripping quotes from raw text.
	raw := nodeText(node, src)
	raw = strings.Trim(raw, `'"`)
	return raw
}

// rubyDefName extracts the name from a class or module node.
func rubyDefName(node *gts.Node, lang *gts.Language, src []byte) string {
	// class/module name is typically a constant (identifier with uppercase).
	name := childByType(node, lang, "constant")
	if name != nil {
		return nodeText(name, src)
	}
	// scope resolution: Foo::Bar
	scope := childByType(node, lang, "scope_resolution")
	if scope != nil {
		return nodeText(scope, src)
	}
	return ""
}

// rubyMethodName extracts the name from a method definition node.
func rubyMethodName(node *gts.Node, lang *gts.Language, src []byte) string {
	id := childByType(node, lang, "identifier")
	if id != nil {
		return nodeText(id, src)
	}
	return ""
}

// rubyClassMethods collects method names defined inside a class or module body.
func rubyClassMethods(node *gts.Node, lang *gts.Language, src []byte) []string {
	var methods []string
	// Find the body_statement child.
	body := childByType(node, lang, "body_statement")
	if body == nil {
		return methods
	}
	for i := 0; i < body.ChildCount(); i++ {
		child := body.Child(i)
		switch child.Type(lang) {
		case "method", "singleton_method":
			name := rubyMethodName(child, lang, src)
			if name != "" {
				methods = append(methods, name)
			}
		}
	}
	return methods
}
