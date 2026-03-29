package cmd

import (
	"fmt"

	"github.com/Leechael/things3--cli/internal/client"
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tags",
		Aliases: []string{"tag"},
		Short:   "Manage tags",
	}

	cmd.AddCommand(newTagCreateCmd())
	cmd.AddCommand(newTagListCmd())
	cmd.AddCommand(newTagGetCmd())
	cmd.AddCommand(newTagUpdateCmd())
	cmd.AddCommand(newTagDeleteCmd())

	return cmd
}

func newTagCreateCmd() *cobra.Command {
	params := client.CreateTagParams{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tag",
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

			result, err := c.CreateTag(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.Name, "name", "", "Tag name")
	flags.StringVar(&params.ParentName, "parent-name", "", "Parent tag name")
	flags.StringVar(&params.ParentID, "parent-id", "", "Parent tag UUID")

	return cmd
}

func newTagListCmd() *cobra.Command {
	params := client.ListTagParams{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tags",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}

			result, err := c.ListTags(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.ParentID, "parent-id", "", "Filter by parent tag UUID")
	flags.StringVar(&params.Search, "search", "", "Search by tag title")
	flags.IntVar(&params.Limit, "limit", 200, "Maximum number of results")
	flags.IntVar(&params.Offset, "offset", 0, "Result offset")

	return cmd
}

func newTagGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a tag by UUID",
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

			result, err := c.GetTag(args[0])
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}
}

func newTagUpdateCmd() *cobra.Command {
	params := client.UpdateTagParams{}
	var (
		newName    string
		parentName string
		parentID   string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a tag",
		RunE: func(cmd *cobra.Command, args []string) error {
			if params.ID == "" && params.Name == "" {
				return fmt.Errorf("either --id or --name is required")
			}

			flags := cmd.Flags()
			params.NewName = stringPointerIfChanged(flags, "new-name", newName)
			params.ParentName = stringPointerIfChanged(flags, "parent-name", parentName)
			params.ParentID = stringPointerIfChanged(flags, "parent-id", parentID)
			if params.NewName == nil && params.ParentName == nil && params.ParentID == nil {
				return fmt.Errorf("at least one of --new-name, --parent-name, or --parent-id is required")
			}

			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}

			result, err := c.UpdateTag(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.ID, "id", "", "Tag UUID")
	flags.StringVar(&params.Name, "name", "", "Current tag name")
	flags.StringVar(&newName, "new-name", "", "New tag name")
	flags.StringVar(&parentName, "parent-name", "", "Parent tag name")
	flags.StringVar(&parentID, "parent-id", "", "Parent tag UUID")

	return cmd
}

func newTagDeleteCmd() *cobra.Command {
	params := client.DeleteTagParams{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a tag",
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

			result, err := c.DeleteTag(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.ID, "id", "", "Tag UUID")
	flags.StringVar(&params.Name, "name", "", "Tag name")

	return cmd
}
