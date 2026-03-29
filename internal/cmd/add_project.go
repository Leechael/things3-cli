package cmd

import (
	"github.com/Leechael/things3--cli/internal/client"
	"github.com/spf13/cobra"
)

func newAddProjectCmd() *cobra.Command {
	cmd := newAddProjectCommand("add-project", "Create a project")
	cmd.Hidden = true
	return cmd
}

func newProjectCreateCmd() *cobra.Command {
	return newAddProjectCommand("create", "Create a project")
}

func newAddProjectCommand(use string, short string) *cobra.Command {
	params := client.AddProjectParams{}
	var (
		completed bool
		canceled  bool
		reveal    bool
	)

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}

			flags := cmd.Flags()
			params.Completed = boolPointerIfChanged(flags, "completed", completed)
			params.Canceled = boolPointerIfChanged(flags, "canceled", canceled)
			params.Reveal = boolPointerIfChanged(flags, "reveal", reveal)

			result, err := c.AddProject(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.Title, "title", "", "Project title")
	flags.StringVar(&params.Notes, "notes", "", "Project notes")
	flags.StringVar(&params.When, "when", "", "When value")
	flags.StringVar(&params.Deadline, "deadline", "", "Deadline yyyy-mm-dd")
	flags.StringVar(&params.Tags, "tags", "", "Comma-separated tag titles")
	flags.StringVar(&params.Area, "area", "", "Target area name")
	flags.StringVar(&params.AreaID, "area-id", "", "Target area UUID")
	flags.StringVar(&params.ToDos, "to-dos", "", "Child to-dos separated by newline")
	flags.StringVar(&params.CreationDate, "creation-date", "", "Creation date (ISO8601)")
	flags.StringVar(&params.CompletionDate, "completion-date", "", "Completion date (ISO8601)")
	flags.BoolVar(&completed, "completed", false, "Set completed status")
	flags.BoolVar(&canceled, "canceled", false, "Set canceled status")
	flags.BoolVar(&reveal, "reveal", false, "Reveal created project in Things")

	return cmd
}
