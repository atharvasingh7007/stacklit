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

	// pub fn / pub struct / pub enum / pub trait / pub type / pub const / pub static
	rustPub = regexp.MustCompile(`(?m)^pub\s+(?:async\s+)?(?:fn|struct|enum|trait|type|const|static)\s+(\w+)`)

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

	// Collect public exports
	seenExport := make(map[string]bool)
	for _, m := range rustPub.FindAllSubmatch(content, -1) {
		name := string(m[1])
		if !seenExport[name] {
			seenExport[name] = true
			info.Exports = append(info.Exports, name)
		}
	}

	// Detect entrypoint
	if rustMain.Match(content) {
		info.IsEntrypoint = true
	}

	return info, nil
}
