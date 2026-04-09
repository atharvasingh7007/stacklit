package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View the generated stacklit.json index",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("stacklit view: stub — not yet implemented")
		return nil
	},
}
