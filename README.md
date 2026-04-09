# stacklit

**Stop burning tokens. Start knowing your codebase.**

stacklit generates `stacklit.json` — a token-efficient, machine-readable index of any codebase. One command. Committed to the repo. Readable by any AI agent without running a server.

Think of it like `.gitignore` tells git what to ignore, `stacklit.json` tells AI agents how to navigate your code.

## The problem

AI coding agents spend most of their time just searching a codebase to build context before writing code. A 2,000-line file read just to find one function signature wastes ~6,000 tokens. Five parallel agents on the same repo tokenize the same files independently. Every new session starts from zero.

## The solution

```bash
npx stacklit init   # or: go install github.com/glincker/stacklit/cmd/stacklit@latest
```

One command generates three files:

| File | Purpose | Committed? |
|------|---------|------------|
| `stacklit.json` | Machine-readable codebase index (~1,500 tokens for a 500-file repo) | Yes |
| `stacklit.mmd` | Mermaid dependency diagram (renders in GitHub) | Yes |
| `stacklit.html` | Interactive visual map (4 views) | No (gitignored) |

## What stacklit.json looks like

```json
{
  "$schema": "https://stacklit.dev/schema/v1.json",
  "version": "1",
  "project": {
    "name": "my-app",
    "type": "monorepo",
    "workspaces": ["packages/api", "packages/web"]
  },
  "tech": {
    "primary_language": "typescript",
    "languages": { "typescript": { "files": 120, "lines": 18000 } }
  },
  "modules": {
    "src/auth": {
      "purpose": "Authentication and session management",
      "files": 8,
      "lines": 1200,
      "exports": ["AuthProvider", "useSession", "loginAction"],
      "depends_on": ["src/db", "src/config"],
      "depended_by": ["src/api"],
      "activity": "high"
    }
  },
  "dependencies": {
    "edges": [["src/api", "src/auth"], ["src/auth", "src/db"]],
    "most_depended": ["src/db", "src/config"]
  },
  "hints": {
    "add_feature": "Create handler in src/api/, add route in src/index.ts",
    "test_command": "npm test"
  }
}
```

A 500-file repo produces ~1,500 tokens of index. That's what an agent wastes reading ONE large file today.

## Interactive visual map

`stacklit.html` includes 4 views — open it with `stacklit view`:

- **Graph** — Force-directed dependency map. See how modules relate at a glance.
- **Tree** — Collapsible directory hierarchy with file counts and line counts.
- **Table** — Sortable module table. Find where the complexity lives.
- **Flow** — Top-down directed dependency flow. Trace imports from entrypoints to leaves.

Self-contained HTML. Works offline. No server needed.

## Language support

| Language | Parser | Accuracy |
|----------|--------|----------|
| Go | stdlib `go/parser` (AST) | High |
| TypeScript / JavaScript | Regex | Good |
| Python | Regex | Good |
| Rust, Java, C#, Ruby, PHP, Swift, Kotlin, C/C++ | Generic (line count only) | Basic |

Tree-sitter support planned for Phase 2.

## Monorepo support

stacklit auto-detects monorepo structures:

- pnpm workspaces
- npm/yarn workspaces
- Go workspaces (`go.work`)
- Turborepo, Nx, Lerna
- Cargo workspaces
- Convention-based (`apps/`, `packages/`, `services/`)

Each workspace is detected and reported in `stacklit.json`.

## CLI commands

```
stacklit init                    # Scan codebase, generate all outputs
stacklit init --hook             # Also install git post-commit hook
stacklit init --workspace <path> # Scan specific monorepo workspace

stacklit generate                # Regenerate from current source
stacklit generate --quiet        # Silent mode (for hooks/CI)

stacklit view                    # Regenerate HTML and open in browser

stacklit --version
stacklit --help
```

## AI agent integration

### Claude Code

Add to your project's `CLAUDE.md`:

```markdown
## Codebase Index
This repo has a stacklit.json. Read it before exploring files.
Use the modules section to locate code. Use hints for conventions.
```

### Cursor / Copilot

Add to `.cursorrules` or `.github/copilot-instructions.md`:

```
This repo has a stacklit.json index at the root.
Read it first to understand the codebase structure before exploring files.
```

### Any MCP-compatible agent

stacklit.json is a plain JSON file. Any agent that can read files can use it. No special integration needed.

## Git integration

### Post-commit hook

```bash
stacklit init --hook
```

Installs a post-commit hook that auto-regenerates `stacklit.json` and `stacklit.mmd` when source files change. Uses Merkle hashing to skip regeneration when only docs/configs changed.

### GitHub Action

```yaml
name: Update Stacklit Index
on:
  push:
    branches: [main]
    paths-ignore: ['stacklit.json', 'stacklit.mmd', '**.md']

jobs:
  stacklit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go install github.com/glincker/stacklit/cmd/stacklit@latest
      - run: stacklit generate --quiet
      - uses: stefanzweifel/git-auto-commit-action@v5
        with:
          commit_message: "chore: update stacklit index"
          file_pattern: "stacklit.json stacklit.mmd"
```

## Install

### Go

```bash
go install github.com/glincker/stacklit/cmd/stacklit@latest
```

### From source

```bash
git clone https://github.com/glincker/stacklit.git
cd stacklit
make build
```

npm and Homebrew distribution coming soon.

## How it works

1. **Walk** — Finds all source files, respects `.gitignore` + built-in ignore list
2. **Parse** — Extracts imports, exports, and metadata per file (Go AST, regex for TS/Python)
3. **Graph** — Groups files into modules by directory, resolves cross-module dependencies
4. **Detect** — Identifies monorepo structure, frameworks, entry points
5. **Git** — Analyzes commit history for file activity heatmap
6. **Render** — Writes `stacklit.json`, `stacklit.mmd`, and `stacklit.html`

Total time: under 100ms for most repos. Under 3 seconds for 10,000+ file repos.

## Why not just use Repomix / code2prompt?

Those tools dump your entire codebase into one file. A 500-file repo becomes 500,000+ tokens. stacklit produces a 1,500-token index of the same repo.

stacklit understands structure. It knows which modules depend on which, what's actively changing, and where to add new features. A flat dump doesn't.

## Why not use Codebase Memory MCP / Axon?

Those are runtime MCP servers that require installing and running a process. stacklit produces a committed JSON file that travels with the repo via git. No server, no setup, no per-machine configuration.

## Contributing

PRs welcome. See the existing code structure — it's straightforward Go with no external dependencies beyond cobra and go-gitignore.

```bash
make build   # Build binary
make test    # Run all tests
make clean   # Remove build artifacts
```

## License

MIT
