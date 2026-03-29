package cmd

import (
	"github.com/Leechael/things3--cli/internal/client"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	params := client.SearchParams{}

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Open Things search",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				params.Query = args[0]
			}

			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}

			result, err := c.Search(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	cmd.Flags().StringVar(&params.Query, "query", "", "Search query")
	return cmd
}
