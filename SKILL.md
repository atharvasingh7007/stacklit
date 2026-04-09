---
name: stacklit-navigator
description: Use stacklit.json to efficiently navigate and understand codebases without wasting tokens on file exploration
---

## When to use

Use this skill when starting work on a repo that has a `stacklit.json` file at the root.

## How to use

1. Check if `stacklit.json` exists at the repo root
2. If it exists, read it FIRST before exploring any files
3. Use the `modules` section to find where code lives
4. Use the `dependencies.edges` to understand how modules connect
5. Use the `hints` section for conventions (test commands, where to add features)
6. Use `git.hot_files` to see what's actively being changed
7. Only then start reading individual source files — and read only the ones you need

## What NOT to do

- Do not read entire directories to understand the codebase
- Do not grep for patterns across all files when stacklit.json tells you which module to look in
- Do not re-explore the codebase every session — the index persists

## Keeping the index fresh

If you notice the index might be stale (you've made significant changes), run:
```
stacklit generate
```
