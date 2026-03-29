package cmd

import (
	"github.com/Leechael/things3--cli/internal/client"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	params := client.ShowParams{}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Navigate in Things via things:///show",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}

			result, err := c.Show(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.ID, "id", "", "Built-in list id or item UUID")
	flags.StringVar(&params.Query, "query", "", "Quick find query")
	flags.StringVar(&params.Filter, "filter", "", "Comma-separated tag filter")

	return cmd
}
