package cli

import (
	"fmt"

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
		fmt.Println("stacklit init: stub — not yet implemented")
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initHook, "hook", false, "Install a git pre-commit hook")
	initCmd.Flags().StringVar(&initWorkspace, "workspace", "", "Path to workspace root (default: current directory)")
	initCmd.Flags().BoolVar(&initSummary, "summary", false, "Generate a high-level summary during init")
}
