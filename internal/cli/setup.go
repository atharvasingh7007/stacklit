package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/glincker/stacklit/internal/derive"
	"github.com/glincker/stacklit/internal/engine"
	"github.com/glincker/stacklit/internal/git"
	"github.com/glincker/stacklit/internal/schema"
	"github.com/glincker/stacklit/internal/setup"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup [claude|cursor|aider]",
		Short: "Auto-configure AI coding tools for this repo",
		Long: `Detects installed AI coding tools and configures them to use Stacklit.

For each tool, it:
  1. Generates/updates stacklit.json if needed
  2. Injects a compact codebase map into the tool's config file
  3. Configures MCP server integration (Claude Code, Cursor)
  4. Installs a git hook to keep the map fresh

Examples:
  stacklit setup          Auto-detect and configure all found tools
  stacklit setup claude   Configure Claude Code only
  stacklit setup cursor   Configure Cursor only
  stacklit setup aider    Configure Aider only`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ""
			if len(args) > 0 {
				target = args[0]
			}

			// Ensure stacklit.json exists and is fresh.
			idx, err := ensureIndex()
			if err != nil {
				return err
			}

			switch target {
			case "claude":
				return setupClaude(idx)
			case "cursor":
				return setupCursor(idx)
			case "aider":
				return setupAider(idx)
			case "":
				return setupAll(idx)
			default:
				return fmt.Errorf("unknown tool %q (use: claude, cursor, aider)", target)
			}
		},
	}
	return cmd
}

func ensureIndex() (*schema.Index, error) {
	// Try to load existing index.
	data, err := os.ReadFile("stacklit.json")
	if err == nil {
		var idx schema.Index
		if err := json.Unmarshal(data, &idx); err == nil {
			fmt.Println("[stacklit] using existing stacklit.json")
			return &idx, nil
		}
	}

	// Generate fresh index.
	fmt.Println("[stacklit] scanning codebase...")
	result, err := engine.Run(engine.Options{Root: ".", Quiet: false})
	if err != nil {
		return nil, fmt.Errorf("scanning codebase: %w", err)
	}
	return result.Index, nil
}

func setupClaude(idx *schema.Index) error {
	fmt.Println("\nConfiguring Claude Code...")
	if err := setup.ConfigureClaude(idx, true); err != nil {
		return err
	}
	return installHookQuiet()
}

func setupCursor(idx *schema.Index) error {
	fmt.Println("\nConfiguring Cursor...")
	if err := setup.ConfigureCursor(idx); err != nil {
		return err
	}
	return installHookQuiet()
}

func setupAider(idx *schema.Index) error {
	fmt.Println("\nConfiguring Aider...")
	if err := setup.ConfigureAider(idx); err != nil {
		return err
	}
	return installHookQuiet()
}

func setupAll(idx *schema.Index) error {
	tools := setup.DetectTools()

	detected := 0
	for _, t := range tools {
		if t.Detected {
			detected++
		}
	}

	if detected == 0 {
		fmt.Println("No AI coding tools detected. Generating codebase map to stdout:")
		fmt.Println()
		fmt.Print(derive.CompactMap(idx))
		fmt.Println("\nCopy the above into your CLAUDE.md, .cursorrules, or agent config.")
		return nil
	}

	fmt.Printf("Detected %d AI tool(s)\n", detected)

	for _, t := range tools {
		if !t.Detected {
			continue
		}
		switch t.Name {
		case "claude":
			if err := setupClaude(idx); err != nil {
				fmt.Printf("  warning: claude setup failed: %v\n", err)
			}
		case "cursor":
			if err := setupCursor(idx); err != nil {
				fmt.Printf("  warning: cursor setup failed: %v\n", err)
			}
		case "aider":
			if err := setupAider(idx); err != nil {
				fmt.Printf("  warning: aider setup failed: %v\n", err)
			}
		}
	}

	if err := installHookQuiet(); err != nil {
		fmt.Printf("  warning: git hook install failed: %v\n", err)
	}

	fmt.Println("\nDone. Your AI tools now have instant codebase context.")
	return nil
}

func installHookQuiet() error {
	root, err := os.Getwd()
	if err != nil {
		return nil
	}
	if err := git.InstallHook(root); err != nil {
		return nil
	}
	fmt.Println("  installed git hook for auto-refresh")
	return nil
}

