package cmd

import (
	"context"
	"fmt"
	"openhue-cli/openhue"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listLightsCmd = &cobra.Command{
	Use:   "list",
	Short: "List lights",
	Long:  `Fetches and displays all available lights from the Philips Hue bridge`,
	Run: func(cmd *cobra.Command, args []string) {

		resp, err := openhue.Api.GetLightsWithResponse(context.Background())
		cobra.CheckErr(err)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "Light ID\tLight Name\tLight Type")
		_, _ = fmt.Fprintln(w, "----\t----\t----")

		lights := (*resp.JSON200).Data
		for _, l := range *lights {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", *l.Id, *l.Metadata.Name, *l.Metadata.Archetype)
		}

		_ = w.Flush()
	},
}

func init() {
	lightsCmd.AddCommand(listLightsCmd)
}
