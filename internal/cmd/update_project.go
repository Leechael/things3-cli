package cmd

import (
	"fmt"

	"github.com/Leechael/things3--cli/internal/client"
	"github.com/spf13/cobra"
)

func newUpdateProjectCmd() *cobra.Command {
	cmd := newUpdateProjectCommand("update-project --id <project-uuid>", "Update a project via things:///update-project")
	cmd.Hidden = true
	return cmd
}

func newProjectUpdateCmd() *cobra.Command {
	return newUpdateProjectCommand("update --id <project-uuid>", "Update a project")
}

func newUpdateProjectCommand(use string, short string) *cobra.Command {
	params := client.UpdateProjectParams{}
	var (
		title          string
		notes          string
		prependNotes   string
		appendNotes    string
		when           string
		deadline       string
		tags           string
		addTags        string
		area           string
		areaID         string
		creationDate   string
		completionDate string
		completed      bool
		canceled       bool
		duplicate      bool
		reveal         bool
	)

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			if params.ID == "" {
				return fmt.Errorf("--id is required")
			}

			flags := cmd.Flags()
			params.Title = stringPointerIfChanged(flags, "title", title)
			params.Notes = stringPointerIfChanged(flags, "notes", notes)
			params.PrependNotes = stringPointerIfChanged(flags, "prepend-notes", prependNotes)
			params.AppendNotes = stringPointerIfChanged(flags, "append-notes", appendNotes)
			params.When = stringPointerIfChanged(flags, "when", when)
			params.Deadline = stringPointerIfChanged(flags, "deadline", deadline)
			params.Tags = stringPointerIfChanged(flags, "tags", tags)
			params.AddTags = stringPointerIfChanged(flags, "add-tags", addTags)
			params.Area = stringPointerIfChanged(flags, "area", area)
			params.AreaID = stringPointerIfChanged(flags, "area-id", areaID)
			params.CreationDate = stringPointerIfChanged(flags, "creation-date", creationDate)
			params.CompletionDate = stringPointerIfChanged(flags, "completion-date", completionDate)
			params.Completed = boolPointerIfChanged(flags, "completed", completed)
			params.Canceled = boolPointerIfChanged(flags, "canceled", canceled)
			params.Duplicate = boolPointerIfChanged(flags, "duplicate", duplicate)
			params.Reveal = boolPointerIfChanged(flags, "reveal", reveal)

			c, err := getClient(cmd)
			if err != nil {
				return err
			}
			f, err := getFormatter(cmd)
			if err != nil {
				return err
			}

			result, err := c.UpdateProject(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.ID, "id", "", "Project UUID")
	flags.StringVar(&title, "title", "", "Replace title")
	flags.StringVar(&notes, "notes", "", "Replace notes")
	flags.StringVar(&prependNotes, "prepend-notes", "", "Prepend notes")
	flags.StringVar(&appendNotes, "append-notes", "", "Append notes")
	flags.StringVar(&when, "when", "", "Update when field")
	flags.StringVar(&deadline, "deadline", "", "Update deadline (use empty value to clear)")
	flags.StringVar(&tags, "tags", "", "Replace tags")
	flags.StringVar(&addTags, "add-tags", "", "Add tags")
	flags.StringVar(&area, "area", "", "Move to area by name")
	flags.StringVar(&areaID, "area-id", "", "Move to area by UUID")
	flags.StringVar(&creationDate, "creation-date", "", "Set creation date (ISO8601)")
	flags.StringVar(&completionDate, "completion-date", "", "Set completion date (ISO8601)")
	flags.BoolVar(&completed, "completed", false, "Set completed status")
	flags.BoolVar(&canceled, "canceled", false, "Set canceled status")
	flags.BoolVar(&duplicate, "duplicate", false, "Duplicate then modify")
	flags.BoolVar(&reveal, "reveal", false, "Reveal updated project in Things")

	return cmd
}
