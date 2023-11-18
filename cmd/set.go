package cmd

import (
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set resources",
	Long: `
Set the values for a specific resource
`,
}

func init() {
	rootCmd.AddCommand(setCmd)
}
