package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// configFileFrameworks maps config file names (relative to root) to framework names.
var configFileFrameworks = []struct {
	file string
	name string
}{
	{"next.config.js", "Next.js"},
	{"next.config.ts", "Next.js"},
	{"next.config.mjs", "Next.js"},
	{"nuxt.config.ts", "Nuxt"},
	{"nuxt.config.js", "Nuxt"},
	{"svelte.config.js", "SvelteKit"},
	{"vite.config.ts", "Vite"},
	{"vite.config.js", "Vite"},
	{"webpack.config.js", "Webpack"},
	{"tailwind.config.js", "Tailwind CSS"},
	{"tailwind.config.ts", "Tailwind CSS"},
	{"postcss.config.js", "PostCSS"},
	{"prisma/schema.prisma", "Prisma"},
	{"drizzle.config.ts", "Drizzle"},
	{"docker-compose.yml", "Docker"},
	{"docker-compose.yaml", "Docker"},
	{"Dockerfile", "Docker"},
	{".eslintrc.js", "ESLint"},
	{"eslint.config.js", "ESLint"},
	{"eslint.config.mjs", "ESLint"},
	{"jest.config.js", "Jest"},
	{"jest.config.ts", "Jest"},
	{"vitest.config.ts", "Vitest"},
	{"playwright.config.ts", "Playwright"},
	{".github/workflows", "GitHub Actions"},
}

// importFrameworks maps import substrings to framework names.
var importFrameworks = []struct {
	substr string
	name   string
}{
	{"express", "Express"},
	{"fastify", "Fastify"},
	{"koa", "Koa"},
	{"hono", "Hono"},
	{"react", "React"},
	{"vue", "Vue"},
	{"svelte", "Svelte"},
	{"@angular/core", "Angular"},
	{"flask", "Flask"},
	{"django", "Django"},
	{"fastapi", "FastAPI"},
	{"gin-gonic/gin", "Gin"},
	{"labstack/echo", "Echo"},
	{"gofiber/fiber", "Fiber"},
	{"gorilla/mux", "Gorilla Mux"},
	{"net/http", "net/http"},
	{"database/sql", "database/sql"},
	{"gorm.io/gorm", "GORM"},
	{"sqlx", "sqlx"},
	{"nats", "NATS"},
	{"redis", "Redis"},
	{"mongodb", "MongoDB"},
	{"mongo", "MongoDB"},
	{"graphql", "GraphQL"},
}

// packageJSONDeps maps package.json dependency keys to framework names.
var packageJSONDeps = []struct {
	pkg  string
	name string
}{
	{"express", "Express"},
	{"fastify", "Fastify"},
	{"koa", "Koa"},
	{"hono", "Hono"},
	{"react", "React"},
	{"react-dom", "React"},
	{"vue", "Vue"},
	{"svelte", "Svelte"},
	{"@angular/core", "Angular"},
	{"next", "Next.js"},
	{"nuxt", "Nuxt"},
	{"@sveltejs/kit", "SvelteKit"},
	{"vite", "Vite"},
	{"webpack", "Webpack"},
	{"tailwindcss", "Tailwind CSS"},
	{"postcss", "PostCSS"},
	{"prisma", "Prisma"},
	{"@prisma/client", "Prisma"},
	{"drizzle-orm", "Drizzle"},
	{"graphql", "GraphQL"},
	{"jest", "Jest"},
	{"vitest", "Vitest"},
	{"@playwright/test", "Playwright"},
	{"eslint", "ESLint"},
	{"redis", "Redis"},
	{"mongoose", "MongoDB"},
	{"mongodb", "MongoDB"},
}

// DetectFrameworks detects frameworks from config file existence, imports, and package.json.
// root is the project root directory; imports is the flat list of all imports from all parsed files.
// Returns a deduplicated, sorted list of framework names.
func DetectFrameworks(root string, imports []string) []string {
	seen := map[string]bool{}

	// 1. Config file existence checks.
	for _, cf := range configFileFrameworks {
		path := filepath.Join(root, filepath.FromSlash(cf.file))
		if _, err := os.Stat(path); err == nil {
			seen[cf.name] = true
		}
	}

	// 2. Import string checks.
	for _, imp := range imports {
		for _, rule := range importFrameworks {
			if strings.Contains(imp, rule.substr) {
				seen[rule.name] = true
			}
		}
	}

	// 3. package.json dependency checks.
	pkgPath := filepath.Join(root, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			allDeps := make(map[string]string, len(pkg.Dependencies)+len(pkg.DevDependencies))
			for k, v := range pkg.Dependencies {
				allDeps[k] = v
			}
			for k, v := range pkg.DevDependencies {
				allDeps[k] = v
			}
			for _, rule := range packageJSONDeps {
				if _, ok := allDeps[rule.pkg]; ok {
					seen[rule.name] = true
				}
			}
		}
	}

	// Collect and sort.
	result := make([]string, 0, len(seen))
	for name := range seen {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}
