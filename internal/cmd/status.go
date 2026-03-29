package cmd

import "github.com/spf13/cobra"

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check Things3 local connectivity and auth readiness",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}

			status, err := c.GetStatus()
			if err != nil {
				return err
			}

			return f.Print(cmd.OutOrStdout(), status)
		},
	}
}
