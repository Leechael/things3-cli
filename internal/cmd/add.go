package cmd

import (
	"fmt"

	"github.com/Leechael/things3--cli/internal/client"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	cmd := newAddCommand("add", "Create a to-do")
	cmd.Hidden = true
	return cmd
}

func newTodoCreateCmd() *cobra.Command {
	return newAddCommand("create", "Create a to-do")
}

func newAddTodoCmd() *cobra.Command {
	return newAddCommand("add-todo", "Create a to-do")
}

func newAddCommand(use string, short string) *cobra.Command {
	params := client.AddToDoParams{}
	var (
		completed      bool
		canceled       bool
		showQuickEntry bool
		reveal         bool

		project   string
		projectID string
		area      string
		areaID    string
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
			params.ShowQuickEntry = boolPointerIfChanged(flags, "show-quick-entry", showQuickEntry)
			params.Reveal = boolPointerIfChanged(flags, "reveal", reveal)

			if err := applyListTargetFlags(&params, project, projectID, area, areaID); err != nil {
				return err
			}

			result, err := c.AddToDo(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.Title, "title", "", "To-do title")
	flags.StringVar(&params.Titles, "titles", "", "Batch titles separated by newline")
	flags.StringVar(&params.Notes, "notes", "", "To-do notes")
	flags.StringVar(&params.When, "when", "", "When value: today|tomorrow|evening|anytime|someday|date")
	flags.StringVar(&params.Deadline, "deadline", "", "Deadline date yyyy-mm-dd")
	flags.StringVar(&params.Tags, "tags", "", "Comma-separated tag titles (multiple supported)")
	flags.StringVar(&params.ChecklistItems, "checklist-items", "", "Checklist items separated by newline")
	flags.StringVar(&params.List, "list", "", "Target list/project/area name")
	flags.StringVar(&params.ListID, "list-id", "", "Target list/project/area UUID")
	flags.StringVar(&project, "project", "", "Target project name")
	flags.StringVar(&projectID, "project-id", "", "Target project UUID")
	flags.StringVar(&area, "area", "", "Target area name")
	flags.StringVar(&areaID, "area-id", "", "Target area UUID")
	flags.StringVar(&params.Heading, "heading", "", "Target heading name")
	flags.StringVar(&params.HeadingID, "heading-id", "", "Target heading UUID")
	flags.StringVar(&params.UseClipboard, "use-clipboard", "", "Use clipboard: replace-title|replace-notes|replace-checklist-items")
	flags.StringVar(&params.CreationDate, "creation-date", "", "Creation date (ISO8601)")
	flags.StringVar(&params.CompletionDate, "completion-date", "", "Completion date (ISO8601)")
	flags.BoolVar(&completed, "completed", false, "Set completed status")
	flags.BoolVar(&canceled, "canceled", false, "Set canceled status")
	flags.BoolVar(&showQuickEntry, "show-quick-entry", false, "Show Quick Entry dialog")
	flags.BoolVar(&reveal, "reveal", false, "Reveal created to-do in Things")

	return cmd
}

func applyListTargetFlags(params *client.AddToDoParams, project string, projectID string, area string, areaID string) error {
	if project != "" && area != "" {
		return fmt.Errorf("--project and --area are mutually exclusive")
	}
	if projectID != "" && areaID != "" {
		return fmt.Errorf("--project-id and --area-id are mutually exclusive")
	}
	if (project != "" || projectID != "" || area != "" || areaID != "") && (params.List != "" || params.ListID != "") {
		return fmt.Errorf("--list/--list-id cannot be combined with --project/--area flags")
	}

	if project != "" {
		params.List = project
	}
	if projectID != "" {
		params.ListID = projectID
	}
	if area != "" {
		params.List = area
	}
	if areaID != "" {
		params.ListID = areaID
	}
	return nil
}
