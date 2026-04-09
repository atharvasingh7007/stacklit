package cli

import (
	"github.com/GLINCKER/stacklit/internal/engine"
	"github.com/spf13/cobra"
)

var generateQuiet bool

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate the stacklit.json codebase index",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := engine.Run(engine.Options{
			Root:  ".",
			Quiet: generateQuiet,
		})
		return err
	},
}

func init() {
	generateCmd.Flags().BoolVar(&generateQuiet, "quiet", false, "Suppress progress output")
}
