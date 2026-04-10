package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/glincker/stacklit/internal/derive"
	"github.com/glincker/stacklit/internal/schema"
)

// Tool represents a detected AI coding tool.
type Tool struct {
	Name     string
	Detected bool
}

// DetectTools checks which AI coding tools are available.
func DetectTools() []Tool {
	tools := []Tool{
		{Name: "claude", Detected: detectClaude()},
		{Name: "cursor", Detected: detectCursor()},
		{Name: "aider", Detected: detectAider()},
	}
	return tools
}

func detectClaude() bool {
	// Check for claude CLI in PATH.
	if _, err := exec.LookPath("claude"); err == nil {
		return true
	}
	// Check for CLAUDE.md in current directory (project already uses Claude).
	if _, err := os.Stat("CLAUDE.md"); err == nil {
		return true
	}
	return false
}

func detectCursor() bool {
	// Check for .cursor/ directory or .cursorrules file.
	if _, err := os.Stat(".cursor"); err == nil {
		return true
	}
	if _, err := os.Stat(".cursorrules"); err == nil {
		return true
	}
	return false
}

func detectAider() bool {
	if _, err := exec.LookPath("aider"); err == nil {
		return true
	}
	if _, err := os.Stat(".aider.conf.yml"); err == nil {
		return true
	}
	return false
}

// ConfigureClaude injects the compact map into CLAUDE.md and optionally sets up MCP.
func ConfigureClaude(idx *schema.Index, withMCP bool) error {
	block := derive.InjectableBlock(idx)

	if err := derive.InjectIntoFile("CLAUDE.md", block); err != nil {
		return fmt.Errorf("injecting into CLAUDE.md: %w", err)
	}
	fmt.Println("  updated CLAUDE.md with codebase map")

	if withMCP {
		if err := configureMCPJSON(); err != nil {
			return fmt.Errorf("configuring MCP: %w", err)
		}
		fmt.Println("  added MCP server to .mcp.json")
	}

	return nil
}

// ConfigureCursor injects the compact map into .cursorrules and sets up MCP.
func ConfigureCursor(idx *schema.Index) error {
	block := derive.InjectableBlock(idx)

	if err := derive.InjectIntoFile(".cursorrules", block); err != nil {
		return fmt.Errorf("injecting into .cursorrules: %w", err)
	}
	fmt.Println("  updated .cursorrules with codebase map")

	// Set up MCP for Cursor.
	if err := configureCursorMCP(); err != nil {
		return fmt.Errorf("configuring Cursor MCP: %w", err)
	}
	fmt.Println("  added MCP server to .cursor/mcp.json")

	return nil
}

// ConfigureAider injects the compact map into .aider.conf.yml conventions.
func ConfigureAider(idx *schema.Index) error {
	block := derive.InjectableBlock(idx)

	if err := derive.InjectIntoFile(".aider.conf.yml", block); err != nil {
		return fmt.Errorf("injecting into .aider.conf.yml: %w", err)
	}
	fmt.Println("  updated .aider.conf.yml with codebase map")

	return nil
}

// configureMCPJSON sets up .mcp.json for Claude Code.
func configureMCPJSON() error {
	return writeMCPConfig(".mcp.json")
}

// configureCursorMCP sets up .cursor/mcp.json for Cursor.
func configureCursorMCP() error {
	if err := os.MkdirAll(".cursor", 0755); err != nil {
		return err
	}
	return writeMCPConfig(filepath.Join(".cursor", "mcp.json"))
}

// writeMCPConfig writes or updates an MCP config file with the stacklit server entry.
func writeMCPConfig(path string) error {
	type MCPServer struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	type MCPConfig struct {
		MCPServers map[string]MCPServer `json:"mcpServers"`
	}

	var cfg MCPConfig

	// Read existing config if present.
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	if cfg.MCPServers == nil {
		cfg.MCPServers = make(map[string]MCPServer)
	}

	cfg.MCPServers["stacklit"] = MCPServer{
		Command: "stacklit",
		Args:    []string{"serve"},
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(out, '\n'), 0644)
}
