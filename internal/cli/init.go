package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/GLINCKER/stacklit/internal/engine"
	"github.com/spf13/cobra"
)

var (
	initHook      bool
	initWorkspace string
	initSummary   bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize stacklit in the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := engine.Run(engine.Options{
			Root:        ".",
			Workspace:   initWorkspace,
			InstallHook: initHook,
		})
		if err != nil {
			return err
		}
		fmt.Println("\nOpening visual map...")
		openBrowser(result.HTMLPath)
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initHook, "hook", false, "Install a git pre-commit hook")
	initCmd.Flags().StringVar(&initWorkspace, "workspace", "", "Path to workspace root (default: current directory)")
	initCmd.Flags().BoolVar(&initSummary, "summary", false, "Generate a high-level summary during init")
}

// openBrowser opens url in the default system browser.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		fmt.Printf("Open %s in your browser\n", url)
		return
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("Open %s in your browser\n", url)
	}
}
