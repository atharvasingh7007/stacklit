# Stacklit

**Your codebase, in 1,500 tokens.**

One command generates a committed JSON index that any AI agent can read. No server, no setup.

## Quick start

```bash
npx stacklit init
```

This scans your codebase and generates:

- `stacklit.json` -- machine-readable codebase index (commit this)
- `stacklit.mmd` -- Mermaid dependency diagram (renders on GitHub)
- `stacklit.html` -- interactive visual map with 4 views (gitignored)

## Why

AI coding agents burn most of their context window figuring out where things live. Reading one large file to find a function signature costs thousands of tokens. Every session starts from scratch.

`stacklit.json` gives any agent a complete map of your codebase -- modules, dependencies, exports, types, activity -- in about 1,500 tokens. Commit it once, every agent benefits.

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

Six tools: `get_overview`, `get_module`, `find_module`, `get_dependencies`, `get_hot_files`, `get_hints`.

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

## Links

- [GitHub](https://github.com/glincker/stacklit)
- [Full documentation](https://github.com/glincker/stacklit#readme)
- [Releases](https://github.com/glincker/stacklit/releases)

## License

MIT
