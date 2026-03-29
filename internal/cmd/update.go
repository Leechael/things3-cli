package cmd

import (
	"fmt"

	"github.com/Leechael/things3--cli/internal/client"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	cmd := newUpdateCommand("update --id <todo-uuid>", "Update a to-do via things:///update")
	cmd.Hidden = true
	return cmd
}

func newTodoUpdateCmd() *cobra.Command {
	return newUpdateCommand("update --id <todo-uuid>", "Update a to-do")
}

func newUpdateTodoCmd() *cobra.Command {
	return newUpdateCommand("update-todo --id <todo-uuid>", "Update a to-do")
}

func newUpdateCommand(use string, short string) *cobra.Command {
	params := client.UpdateToDoParams{}
	var (
		completed bool
		canceled  bool
		duplicate bool
		reveal    bool

		title                 string
		notes                 string
		prependNotes          string
		appendNotes           string
		when                  string
		deadline              string
		tags                  string
		addTags               string
		checklistItems        string
		prependChecklistItems string
		appendChecklistItems  string
		list                  string
		listID                string
		heading               string
		headingID             string
		creationDate          string
		completionDate        string
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
			params.ChecklistItems = stringPointerIfChanged(flags, "checklist-items", checklistItems)
			params.PrependChecklistItems = stringPointerIfChanged(flags, "prepend-checklist-items", prependChecklistItems)
			params.AppendChecklistItems = stringPointerIfChanged(flags, "append-checklist-items", appendChecklistItems)
			params.List = stringPointerIfChanged(flags, "list", list)
			params.ListID = stringPointerIfChanged(flags, "list-id", listID)
			params.Heading = stringPointerIfChanged(flags, "heading", heading)
			params.HeadingID = stringPointerIfChanged(flags, "heading-id", headingID)
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

			result, err := c.UpdateToDo(params)
			if err != nil {
				return err
			}
			return f.Print(cmd.OutOrStdout(), result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&params.ID, "id", "", "To-do UUID")
	flags.StringVar(&title, "title", "", "Replace title")
	flags.StringVar(&notes, "notes", "", "Replace notes")
	flags.StringVar(&prependNotes, "prepend-notes", "", "Prepend notes")
	flags.StringVar(&appendNotes, "append-notes", "", "Append notes")
	flags.StringVar(&when, "when", "", "Update when field")
	flags.StringVar(&deadline, "deadline", "", "Update deadline (use empty value to clear)")
	flags.StringVar(&tags, "tags", "", "Replace tags")
	flags.StringVar(&addTags, "add-tags", "", "Add tags")
	flags.StringVar(&checklistItems, "checklist-items", "", "Replace checklist items")
	flags.StringVar(&prependChecklistItems, "prepend-checklist-items", "", "Prepend checklist items")
	flags.StringVar(&appendChecklistItems, "append-checklist-items", "", "Append checklist items")
	flags.StringVar(&list, "list", "", "Move to target list by name")
	flags.StringVar(&listID, "list-id", "", "Move to target list by UUID")
	flags.StringVar(&heading, "heading", "", "Move to heading by name")
	flags.StringVar(&headingID, "heading-id", "", "Move to heading by UUID")
	flags.StringVar(&creationDate, "creation-date", "", "Set creation date (ISO8601)")
	flags.StringVar(&completionDate, "completion-date", "", "Set completion date (ISO8601)")
	flags.BoolVar(&completed, "completed", false, "Set completed status")
	flags.BoolVar(&canceled, "canceled", false, "Set canceled status")
	flags.BoolVar(&duplicate, "duplicate", false, "Duplicate then modify")
	flags.BoolVar(&reveal, "reveal", false, "Reveal updated to-do in Things")

	return cmd
}
