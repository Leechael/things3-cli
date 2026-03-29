package cmd

import (
	"fmt"
	"os"

	"github.com/Leechael/things3--cli/internal/client"
	"github.com/spf13/cobra"
)

func newJSONCmd() *cobra.Command {
	params := client.JSONParams{}
	var (
		dataFile string
		reveal   bool
	)

	cmd := &cobra.Command{
		Use:   "json",
		Short: "Run batch operations via things:///json",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dataFile != "" {
				content, err := os.ReadFile(dataFile)
				if err != nil {
					return fmt.Errorf("read --data-file: %w", err)
				}
				params.Data = string(content)
			}
			if params.Data == "" {
				return fmt.Errorf("either --data or --data-file is required")
			}
			params.Reveal = boolPointerIfChanged(cmd.Flags(), "reveal", reveal)

			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}

			result, err := c.JSON(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.Data, "data", "", "JSON payload string")
	flags.StringVar(&dataFile, "data-file", "", "Path to JSON payload file")
	flags.BoolVar(&reveal, "reveal", false, "Reveal first created project")

	return cmd
}
