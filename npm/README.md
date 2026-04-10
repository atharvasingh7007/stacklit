# Stacklit

**Your codebase, in 1,500 tokens.**

One command generates a committed JSON index that any AI agent can read. No server, no setup.

[![CI](https://github.com/glincker/stacklit/actions/workflows/ci.yml/badge.svg)](https://github.com/glincker/stacklit/actions)
[![npm](https://img.shields.io/npm/v/stacklit)](https://www.npmjs.com/package/stacklit)
[![downloads](https://img.shields.io/npm/dm/stacklit)](https://www.npmjs.com/package/stacklit)
[![License](https://img.shields.io/badge/license-MIT-green)](https://opensource.org/licenses/MIT)
[![platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](https://github.com/glincker/stacklit/releases)
[![node](https://img.shields.io/badge/node-%3E%3D16-brightgreen)](https://nodejs.org)

## Quick start

```bash
npx stacklit init
```

![Stacklit demo](https://raw.githubusercontent.com/glincker/stacklit/master/demo.gif)

**Without stacklit:** Agent reads 8-12 files to build context. ~400,000 tokens. 45 seconds before writing a line.

**With stacklit:** Agent reads `stacklit.json`. ~1,500 tokens. Knows every module, dependency, and convention instantly.

## What you get

| File | What it does | Committed? |
|------|-------------|------------|
| `stacklit.json` | Machine-readable codebase index | Yes |
| `DEPENDENCIES.md` | Mermaid dependency diagram (renders on GitHub) | Yes |
| `stacklit.html` | Interactive visual map with 4 views | No (gitignored) |

## Visual map

![Stacklit visual map](https://raw.githubusercontent.com/glincker/stacklit/master/stacklit-og.png)

## MCP server

```bash
stacklit serve
```

Add to Claude Desktop or Cursor MCP config:

```json
{
  "mcpServers": {
    "stacklit": {
      "command": "stacklit",
      "args": ["serve"]
    }
  }
}
```

Seven tools: `get_overview`, `get_module`, `find_module`, `list_modules`, `get_dependencies`, `get_hot_files`, `get_hints`.

## CLI

```
stacklit init                    # scan, generate all outputs, open HTML
stacklit generate                # regenerate from current source
stacklit view                    # regenerate HTML and open in browser
stacklit diff                    # check if index is stale
stacklit serve                   # start MCP server
```

## Language support

Go (AST), TypeScript, JavaScript, Python, Rust, Java (regex), plus generic fallback for any language.

## Compared to alternatives

| | Stacklit | Repomix | Aider repo-map | Codebase Memory MCP |
|---|---|---|---|---|
| Output | ~1,500 token JSON | 500k+ token dump | Ephemeral text | SQLite DB |
| Committed to repo | Yes | Too large | No | No |
| Dependency graph | Yes | No | Yes | Yes |
| Visual output | HTML (4 views) | No | No | No |
| MCP server | Yes | No | No | Yes |
| Runtime needed | No | No | Yes (Python) | Yes (C server) |

## Links

- [GitHub](https://github.com/glincker/stacklit)
- [Full documentation](https://github.com/glincker/stacklit#readme)
- [Releases](https://github.com/glincker/stacklit/releases)

## License

MIT
