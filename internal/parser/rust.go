package parser

import (
	"regexp"
	"strings"
)

var (
	// use std::collections::HashMap; → extract first two path segments
	rustUse = regexp.MustCompile(`(?m)^use\s+([a-zA-Z_][a-zA-Z0-9_]*(?:::[a-zA-Z_][a-zA-Z0-9_]*)*)`)

	// mod submodule;
	rustMod = regexp.MustCompile(`(?m)^mod\s+(\w+)\s*;`)

	// extern crate serde;
	rustExternCrate = regexp.MustCompile(`(?m)^extern\s+crate\s+(\w+)`)

	// pub fn / pub struct / pub enum / pub trait / pub type / pub const / pub static — name only (for dedup)
	rustPub = regexp.MustCompile(`(?m)^pub\s+(?:async\s+)?(?:fn|struct|enum|trait|type|const|static)\s+(\w+)`)

	// Full pub signature: captures keyword + name + optional params + optional return
	rustPubSig = regexp.MustCompile(`(?m)^(pub\s+(?:async\s+)?(?:fn\s+\w+\s*\([^)]*\)(?:\s*->\s*\S+)?|struct\s+\w+|enum\s+\w+|trait\s+\w+|type\s+\w+\s*=\s*[^;]+|const\s+\w+\s*:\s*\S+|static\s+\w+\s*:\s*\S+))`)

	// fn main() entrypoint
	rustMain = regexp.MustCompile(`(?m)^fn\s+main\s*\(`)
)

// RustParser parses Rust source files using regex.
type RustParser struct{}

// CanParse returns true for .rs files.
func (r *RustParser) CanParse(filename string) bool {
	return extIs(filename, ".rs")
}

// Parse extracts imports, exports, and entrypoint from a Rust file.
func (r *RustParser) Parse(path string, content []byte) (*FileInfo, error) {
	info := &FileInfo{
		Path:      path,
		Language:  "rust",
		LineCount: countLines(content),
	}

	seen := make(map[string]bool)
	addImport := func(s string) {
		if !seen[s] {
			seen[s] = true
			info.Imports = append(info.Imports, s)
		}
	}

	// use paths: take first two segments only
	for _, m := range rustUse.FindAllSubmatch(content, -1) {
		full := string(m[1])
		parts := strings.SplitN(full, "::", 3)
		var segment string
		if len(parts) >= 2 {
			segment = parts[0] + "::" + parts[1]
		} else {
			segment = parts[0]
		}
		addImport(segment)
	}

	// mod declarations
	for _, m := range rustMod.FindAllSubmatch(content, -1) {
		addImport(string(m[1]))
	}

	// extern crate
	for _, m := range rustExternCrate.FindAllSubmatch(content, -1) {
		addImport(string(m[1]))
	}

	// Collect public exports with full signatures.
	seenExport := make(map[string]bool)
	nameMatches := rustPub.FindAllSubmatch(content, -1)
	sigMatches := rustPubSig.FindAllSubmatch(content, -1)
	rustSigMap := make(map[string]string, len(sigMatches))
	for i, sm := range sigMatches {
		if i < len(nameMatches) {
			name := string(nameMatches[i][1])
			sig := strings.TrimSpace(string(sm[1]))
			if len(sig) > 80 {
				sig = sig[:80]
			}
			rustSigMap[name] = sig
		}
	}
	for _, m := range nameMatches {
		name := string(m[1])
		if !seenExport[name] {
			seenExport[name] = true
			if sig, ok := rustSigMap[name]; ok && sig != "" {
				info.Exports = append(info.Exports, sig)
			} else {
				info.Exports = append(info.Exports, name)
			}
		}
	}

	// Detect entrypoint
	if rustMain.Match(content) {
		info.IsEntrypoint = true
	}

	return info, nil
}
