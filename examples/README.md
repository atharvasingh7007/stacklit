# Stacklit examples

Real-world `stacklit.json` outputs from popular open source projects.

| Project | Language | Files | Lines of code | Index tokens |
|---------|----------|-------|---------------|-------------|
| [Express.js](express/) | JavaScript | 141 | 21,346 | 3,765 |
| [FastAPI](fastapi/) | Python | 1,131 | 108,075 | 4,142 |
| [Gin](gin/) | Go | 100 | 23,829 | 3,361 |
| [Axum](axum/) | Rust | 300 | 43,997 | 14,371 |

FastAPI has 108,000 lines of code. Its entire structure -- modules, dependencies, exports, types, activity -- fits in 4,142 tokens. That is less than what an agent spends reading a single large file.

## Reproduce

```bash
git clone --depth 1 https://github.com/expressjs/express.git
cd express
npx stacklit init
```
