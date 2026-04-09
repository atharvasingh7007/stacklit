package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var generateQuiet bool

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the stacklit.json codebase index",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("stacklit generate: stub — not yet implemented")
		return nil
	},
}

func init() {
	generateCmd.Flags().BoolVar(&generateQuiet, "quiet", false, "Suppress progress output")
}
