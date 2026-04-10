# Gallery — Real Stacklit Outputs

Pre-built `stacklit.json` indexes for popular open-source projects. Browse these to see what Stacklit extracts from real codebases.

| Project | Language | Files | Lines | Index tokens | Compact map |
|---------|----------|-------|-------|-------------|-------------|
| [Express.js](express/) | JavaScript | 141 | 21,346 | 3,765 | ~200 |
| [FastAPI](fastapi/) | Python | 1,131 | 108,075 | 4,142 | ~280 |
| [Gin](gin/) | Go | 100 | 23,829 | 3,361 | ~220 |
| [Axum](axum/) | Rust | 300 | 43,997 | 14,371 | ~250 |

FastAPI has 108,000 lines of code. Its entire structure -- modules, dependencies, exports, types, activity -- fits in 4,142 tokens (full index) or ~280 tokens (compact map). An agent exploring manually would burn 400,000+ tokens.

## Reproduce

```bash
git clone --depth 1 https://github.com/expressjs/express.git
cd express && stacklit init
```

## Contribute

Have a favorite project? Generate and PR its index:

```bash
git clone --depth 1 <repo-url>
cd <repo> && stacklit init
cp stacklit.json DEPENDENCIES.md ../stacklit/examples/<name>/
```
