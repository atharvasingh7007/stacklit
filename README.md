# Stacklit

**108,000 lines of code. 4,000 tokens of index.**

One command makes any repo AI-agent-ready. No server, no setup.

[![CI](https://github.com/glincker/stacklit/actions/workflows/ci.yml/badge.svg)](https://github.com/glincker/stacklit/actions)
[![Release](https://img.shields.io/github/v/release/glincker/stacklit)](https://github.com/glincker/stacklit/releases)
[![npm](https://img.shields.io/npm/v/stacklit)](https://www.npmjs.com/package/stacklit)
[![License](https://img.shields.io/badge/license-MIT-green)](https://opensource.org/licenses/MIT)

## Quick start

```bash
npx stacklit init
```

![Stacklit demo](demo.gif)

## The problem

AI coding agents burn most of their context window figuring out where things live. Reading one large file to find a function signature costs thousands of tokens. Five agents on the same repo each rebuild the same mental model from scratch.

**Without stacklit:** Agent reads 8-12 files to build context. ~400,000 tokens. 45 seconds before writing a line.

**With stacklit:** Agent reads `stacklit.json`. ~4,000 tokens. Knows every module, dependency, and convention instantly.

## What you get

| File | Purpose | Committed? |
|------|---------|------------|
| `stacklit.json` | Machine-readable codebase index | Yes |
| `DEPENDENCIES.md` | Mermaid dependency diagram (renders on GitHub) | Yes |
| `stacklit.html` | Interactive visual map with 4 views | No (gitignored) |

### stacklit.json

```json
{
  "modules": {
    "src/auth": {
      "purpose": "Authentication and session management",
      "files": 8, "lines": 1200,
      "exports": ["AuthProvider", "useSession()", "loginAction()"],
      "depends_on": ["src/db", "src/config"],
      "activity": "high"
    }
  },
  "hints": {
    "add_feature": "Create handler in src/api/, add route in src/index.ts",
    "test_command": "npm test"
  }
}
```

### Token efficiency (measured)

| Project | Language | Lines of code | Index tokens |
|---------|----------|---------------|-------------|
| Express.js | JavaScript | 21,346 | 3,765 |
| FastAPI | Python | 108,075 | 4,142 |
| Gin | Go | 23,829 | 3,361 |
| Axum | Rust | 43,997 | 14,371 |

See [examples/](examples/) for real outputs.

### Visual map

![Stacklit visual map](stacklit-og.png)

Four views: **Graph** (force-directed dependency map), **Tree** (collapsible directories), **Table** (sortable modules with search), **Flow** (top-down dependency flow). Self-contained HTML, works offline.

## How it works

1. **Walk** -- Find source files, respect `.gitignore`
2. **Parse** -- Extract imports, exports, types via tree-sitter AST (11 languages)
3. **Graph** -- Group files into modules, resolve dependencies
4. **Detect** -- Identify monorepo structure, frameworks, entrypoints
5. **Git** -- 90-day commit history for file activity heatmap
6. **Render** -- Write JSON, Mermaid, and HTML

Under 100ms for most repos. Under 7 seconds for 20,000+ file repos.

## Language support

| Language | Parser | Extracts |
|----------|--------|----------|
| Go | stdlib AST | imports, exports with signatures, struct fields, interface methods |
| TypeScript/JS | tree-sitter | imports (ESM, CJS, dynamic), classes, interfaces, type aliases |
| Python | tree-sitter | imports, classes with methods, type hints, decorators |
| Rust | tree-sitter | use/mod/crate, pub items with generics, trait methods |
| Java | tree-sitter | imports, public classes, method signatures with types |
| C# | tree-sitter | using directives, public types, method signatures |
| Ruby | tree-sitter | require, classes, modules, methods |
| PHP | tree-sitter | namespace use, classes, traits, public methods |
| Kotlin | tree-sitter | imports, classes, objects, functions |
| Swift | tree-sitter | imports, structs, classes, protocols |
| C/C++ | tree-sitter | includes, functions, structs, typedefs |

Generic fallback handles any other language (line count + language detection).

## MCP server

```bash
stacklit serve
```

Seven tools: `get_overview`, `get_module`, `find_module`, `list_modules`, `get_dependencies`, `get_hot_files`, `get_hints`.

Add to Claude Desktop, Claude Code, or Cursor:

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

For Claude Code without MCP, add to `CLAUDE.md`:

```
Read stacklit.json before exploring files. Use modules to locate code, hints for conventions.
```

## Install

```bash
npx stacklit init                # npm (recommended)
go install github.com/glincker/stacklit/cmd/stacklit@latest  # go
```

Binary releases for macOS, Linux, and Windows at [GitHub Releases](https://github.com/glincker/stacklit/releases).

## CLI

```
stacklit init                    # scan, generate, open HTML
stacklit init --hook             # also install git post-commit hook
stacklit init --multi repos.txt  # polyrepo: scan multiple repos
stacklit generate                # regenerate from current source
stacklit view                    # regenerate HTML, open in browser
stacklit diff                    # check if index is stale (Merkle hash)
stacklit serve                   # start MCP server
```

## Git integration

```bash
stacklit init --hook    # auto-regenerate on commit
```

<details>
<summary>GitHub Action</summary>

```yaml
name: Update stacklit index
on:
  push:
    branches: [main]
    paths-ignore: ['stacklit.json', 'DEPENDENCIES.md', '**.md']

jobs:
  stacklit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: go install github.com/glincker/stacklit/cmd/stacklit@latest
      - run: stacklit generate --quiet
      - uses: stefanzweifel/git-auto-commit-action@v5
        with:
          commit_message: "chore: update stacklit index"
          file_pattern: "stacklit.json DEPENDENCIES.md"
```

</details>

<details>
<summary>Configuration (stacklit.toml)</summary>

```toml
ignore = ["vendor/", "generated/"]
max_depth = 3

[output]
json = "stacklit.json"
mermaid = "DEPENDENCIES.md"
html = "stacklit.html"
```

</details>

## Compared to alternatives

| | Stacklit | Repomix | Aider repo-map | Codebase Memory MCP |
|---|---|---|---|---|
| Output | ~4k token index | 500k+ token dump | Ephemeral text | SQLite DB |
| Committed to repo | Yes | Too large | No | No |
| Dependency graph | Yes | No | Yes | Yes |
| Visual output | HTML (4 views) | No | No | No |
| MCP server | Yes (7 tools) | No | No | Yes |
| Monorepo aware | Yes | No | No | No |
| Languages | 11 (tree-sitter) | N/A | Many | 66 |
| Runtime needed | No | No | Yes (Python) | Yes (C server) |
| Single binary | Yes (Go) | No (Node) | No (Python) | Yes (C) |

## Monorepo support

Auto-detects: pnpm, npm, yarn workspaces, Go workspaces, Turborepo, Nx, Lerna, Cargo workspaces, and convention directories (`apps/`, `packages/`, `services/`).

## Contributing

```bash
make build   # build binary
make test    # run all tests
```

## License

MIT
