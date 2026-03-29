package cmd

import (
	"fmt"

	"github.com/Leechael/things3--cli/internal/client"
	"github.com/spf13/cobra"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "projects",
		Aliases: []string{"project"},
		Short:   "Manage projects",
	}

	cmd.AddCommand(newProjectCreateCmd())
	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectGetCmd())
	cmd.AddCommand(newProjectUpdateCmd())
	cmd.AddCommand(newProjectDeleteCmd())

	return cmd
}

func newLSProjectsCmd() *cobra.Command {
	return newProjectListCommand("ls-projects", "List projects (default: incomplete only)", nil, true, true)
}

func newProjectListCmd() *cobra.Command {
	return newProjectListCommand("list", "List projects", []string{"ls"}, false, false)
}

func newProjectListCommand(use string, short string, aliases []string, defaultIncomplete bool, exposeAllFlag bool) *cobra.Command {
	params := client.ListProjectParams{}
	var all bool

	cmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Aliases: aliases,
		RunE: func(cmd *cobra.Command, args []string) error {
			if exposeAllFlag && all {
				if cmd.Flags().Changed("status") {
					return fmt.Errorf("--all and --status are mutually exclusive")
				}
				params.Status = ""
			} else if defaultIncomplete && !cmd.Flags().Changed("status") {
				params.Status = "incomplete"
			}

			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}

			result, err := c.ListProjects(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.Status, "status", "", "Filter by status: incomplete|completed|canceled")
	if exposeAllFlag {
		flags.BoolVar(&all, "all", false, "Show all statuses (overrides default incomplete filter)")
	}
	flags.StringVar(&params.AreaID, "area-id", "", "Filter by area UUID")
	flags.StringVar(&params.Search, "search", "", "Search in title and notes")
	flags.BoolVar(&params.IncludeTrashed, "include-trashed", false, "Include trashed projects")
	flags.IntVar(&params.Limit, "limit", 200, "Maximum number of results")
	flags.IntVar(&params.Offset, "offset", 0, "Result offset")

	return cmd
}

func newProjectGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a project by UUID",
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

			result, err := c.GetProject(args[0])
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}
}

func newProjectDeleteCmd() *cobra.Command {
	params := client.DeleteProjectParams{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a project",
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

			result, err := c.DeleteProject(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.ID, "id", "", "Project UUID")
	flags.StringVar(&params.Name, "name", "", "Project title")

	return cmd
}
