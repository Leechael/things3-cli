package cmd

import (
	"fmt"

	"github.com/Leechael/things3--cli/internal/client"
	"github.com/spf13/cobra"
)

func newTodoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "todos",
		Aliases: []string{"todo"},
		Short:   "Manage to-do items",
		Hidden:  true,
	}

	cmd.AddCommand(newTodoCreateCmd())
	cmd.AddCommand(newTodoListCmd())
	cmd.AddCommand(newTodoGetCmd())
	cmd.AddCommand(newTodoUpdateCmd())
	cmd.AddCommand(newTodoDeleteCmd())

	return cmd
}

func newLSTodoCmd() *cobra.Command {
	return newTodoListCommand("ls", "List to-dos", []string{"list"})
}

func newTodoListCmd() *cobra.Command {
	return newTodoListCommand("list", "List to-dos", []string{"ls"})
}

func newTodoListCommand(use string, short string, aliases []string) *cobra.Command {
	params := client.ListToDoParams{}

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

			result, err := c.ListToDos(params)
			if err != nil {
				return err
			}

			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.Status, "status", "", "Filter by status: incomplete|completed|canceled")
	flags.StringVar(&params.ProjectID, "project-id", "", "Filter by project UUID")
	flags.StringVar(&params.ProjectName, "project", "", "Filter by project name")
	flags.StringVar(&params.AreaID, "area-id", "", "Filter by area UUID")
	flags.StringVar(&params.AreaName, "area", "", "Filter by area name")
	flags.StringVar(&params.HeadingID, "heading-id", "", "Filter by heading UUID")
	flags.StringVar(&params.Tag, "tag", "", "Filter by one tag title")
	flags.StringVar(&params.Tags, "tags", "", "Filter by multiple tag titles (comma-separated, AND match)")
	flags.StringVar(&params.Search, "search", "", "Search in title and notes")
	flags.BoolVar(&params.IncludeTrashed, "include-trashed", false, "Include trashed items")
	flags.IntVar(&params.Limit, "limit", 200, "Maximum number of results")
	flags.IntVar(&params.Offset, "offset", 0, "Result offset")

	return cmd
}

func newGetTodoCmd() *cobra.Command {
	return newTodoGetCommand("get-todo <id>", "Get a to-do by UUID")
}

func newTodoGetCmd() *cobra.Command {
	return newTodoGetCommand("get <id>", "Get a to-do by UUID")
}

func newTodoGetCommand(use string, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
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

			result, err := c.GetToDo(args[0])
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}
}

func newDeleteTodoCmd() *cobra.Command {
	return newTodoDeleteCommand("delete-todo", "Delete a to-do via AppleScript")
}

func newTodoDeleteCmd() *cobra.Command {
	return newTodoDeleteCommand("delete", "Delete a to-do via AppleScript")
}

func newTodoDeleteCommand(use string, short string) *cobra.Command {
	params := client.DeleteToDoParams{}

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
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

			result, err := c.DeleteToDo(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.ID, "id", "", "To-do UUID")
	flags.StringVar(&params.Name, "name", "", "To-do title (used directly by AppleScript)")

	return cmd
}
