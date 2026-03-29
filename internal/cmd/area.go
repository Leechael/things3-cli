package cmd

import (
	"fmt"

	"github.com/Leechael/things3--cli/internal/client"
	"github.com/spf13/cobra"
)

func newAreaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "areas",
		Aliases: []string{"area"},
		Short:   "Manage areas",
	}

	cmd.AddCommand(newAreaCreateCmd())
	cmd.AddCommand(newAreaListCmd())
	cmd.AddCommand(newAreaGetCmd())
	cmd.AddCommand(newAreaUpdateCmd())
	cmd.AddCommand(newAreaDeleteCmd())

	return cmd
}

func newLSAreasCmd() *cobra.Command {
	return newAreaListCommand("ls-areas", "List areas", nil)
}

func newAreaCreateCmd() *cobra.Command {
	params := client.CreateAreaParams{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an area",
		RunE: func(cmd *cobra.Command, args []string) error {
			if params.Name == "" {
				return fmt.Errorf("--name is required")
			}
			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}
			result, err := c.CreateArea(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.Name, "name", "", "Area name")
	flags.StringVar(&params.TagNames, "tags", "", "Comma-separated tag names")

	return cmd
}

func newAreaListCmd() *cobra.Command {
	return newAreaListCommand("list", "List areas", []string{"ls"})
}

func newAreaListCommand(use string, short string, aliases []string) *cobra.Command {
	params := client.ListAreaParams{}

	cmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Aliases: aliases,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}

			result, err := c.ListAreas(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.Search, "search", "", "Search by area title")
	flags.IntVar(&params.Limit, "limit", 200, "Maximum number of results")
	flags.IntVar(&params.Offset, "offset", 0, "Result offset")

	return cmd
}

func newAreaGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an area by UUID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}

			result, err := c.GetArea(args[0])
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}
}

func newAreaUpdateCmd() *cobra.Command {
	params := client.UpdateAreaParams{}
	var (
		newName string
		tags    string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an area",
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			params.NewName = stringPointerIfChanged(flags, "new-name", newName)
			params.TagNames = stringPointerIfChanged(flags, "tags", tags)

			if params.ID == "" && params.Name == "" {
				return fmt.Errorf("either --id or --name is required")
			}
			if params.NewName == nil && params.TagNames == nil {
				return fmt.Errorf("at least one of --new-name or --tags is required")
			}

			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}
			result, err := c.UpdateArea(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.ID, "id", "", "Area UUID")
	flags.StringVar(&params.Name, "name", "", "Current area name")
	flags.StringVar(&newName, "new-name", "", "New area name")
	flags.StringVar(&tags, "tags", "", "Replace area tags with comma-separated names")

	return cmd
}

func newAreaDeleteCmd() *cobra.Command {
	params := client.DeleteAreaParams{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an area",
		RunE: func(cmd *cobra.Command, args []string) error {
			if params.ID == "" && params.Name == "" {
				return fmt.Errorf("either --id or --name is required")
			}
			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}
			result, err := c.DeleteArea(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.ID, "id", "", "Area UUID")
	flags.StringVar(&params.Name, "name", "", "Area name")

	return cmd
}
