package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Leechael/things3--cli/internal/client"
	"github.com/Leechael/things3--cli/internal/output"
	"github.com/spf13/cobra"
)

const (
	ExitOK       = 0
	ExitError    = 1
	ExitAuth     = 2
	ExitNotFound = 3
)

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	var (
		token   string
		jsonOut bool
		plain   bool
		jqExpr  string
	)

	cmd := &cobra.Command{
		Use:           "things3-cli",
		Short:         "Things3 command line interface",
		Long:          "Things3 CLI (SQLite read + URL Scheme write + AppleScript management)",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if jqExpr != "" && !jsonOut {
				return fmt.Errorf("--jq requires --json")
			}
			if jsonOut && plain {
				return fmt.Errorf("--json and --plain cannot be used together")
			}
			return nil
		},
	}

	pf := cmd.PersistentFlags()
	pf.StringVar(&token, "token", "", "Things URL auth token (falls back to THINGS_API_TOKEN)")
	pf.BoolVar(&jsonOut, "json", false, "Output JSON")
	pf.BoolVar(&plain, "plain", false, "Output plain tab-separated format without headers")
	pf.StringVar(&jqExpr, "jq", "", "Filter JSON output using gojq (requires --json)")

	cmd.AddGroup(
		&cobra.Group{ID: "todos", Title: "Todo commands:"},
		&cobra.Group{ID: "projects", Title: "Projects commands:"},
		&cobra.Group{ID: "areas", Title: "Area commands:"},
		&cobra.Group{ID: "tags", Title: "Tag commands:"},
		&cobra.Group{ID: "general", Title: "General commands:"},
	)

	addTodoCmd := newAddTodoCmd()
	addTodoCmd.GroupID = "todos"

	lsTodoCmd := newLSTodoCmd()
	lsTodoCmd.GroupID = "todos"

	inboxCmd := newInboxCmd()
	inboxCmd.GroupID = "todos"

	todayCmd := newTodayCmd()
	todayCmd.GroupID = "todos"

	upcomingCmd := newUpcomingCmd()
	upcomingCmd.GroupID = "todos"

	anytimeCmd := newAnytimeCmd()
	anytimeCmd.GroupID = "todos"

	somedayCmd := newSomedayCmd()
	somedayCmd.GroupID = "todos"

	getTodoCmd := newGetTodoCmd()
	getTodoCmd.GroupID = "todos"

	updateTodoCmd := newUpdateTodoCmd()
	updateTodoCmd.GroupID = "todos"

	deleteTodoCmd := newDeleteTodoCmd()
	deleteTodoCmd.GroupID = "todos"

	projectsCmd := newProjectCmd()
	projectsCmd.GroupID = "projects"

	areasCmd := newAreaCmd()
	areasCmd.GroupID = "areas"

	statusCmd := newStatusCmd()
	statusCmd.GroupID = "general"

	tagCmd := newTagCmd()
	tagCmd.GroupID = "tags"

	showCmd := newShowCmd()
	showCmd.GroupID = "general"

	searchCmd := newSearchCmd()
	searchCmd.GroupID = "general"

	versionCmd := newVersionCmd()
	versionCmd.GroupID = "general"

	jsonCmd := newJSONCmd()
	jsonCmd.GroupID = "general"

	helpCmd := newTopicHelpCmd(cmd)
	helpCmd.GroupID = "general"
	cmd.SetHelpCommand(helpCmd)

	legacyAddCmd := newAddCmd()
	legacyAddCmd.Hidden = true

	legacyAddProjectCmd := newAddProjectCmd()
	legacyAddProjectCmd.Hidden = true

	legacyUpdateCmd := newUpdateCmd()
	legacyUpdateCmd.Hidden = true

	legacyUpdateProjectCmd := newUpdateProjectCmd()
	legacyUpdateProjectCmd.Hidden = true

	cmd.AddCommand(
		statusCmd,
		addTodoCmd,
		lsTodoCmd,
		inboxCmd,
		todayCmd,
		upcomingCmd,
		anytimeCmd,
		somedayCmd,
		getTodoCmd,
		updateTodoCmd,
		deleteTodoCmd,
		projectsCmd,
		areasCmd,
		tagCmd,
		showCmd,
		searchCmd,
		versionCmd,
		jsonCmd,
		legacyAddCmd,
		legacyAddProjectCmd,
		legacyUpdateCmd,
		legacyUpdateProjectCmd,
	)

	defaultHelpFunc := cmd.HelpFunc()
	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		if c != cmd {
			defaultHelpFunc(c, args)
			return
		}
		renderRootHelp(c)
	})

	return cmd
}

func getClient(cmd *cobra.Command) (*client.Client, error) {
	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		token = os.Getenv("THINGS_API_TOKEN")
	}
	return client.New(client.Config{Token: token})
}

func getFormatter(cmd *cobra.Command) (*output.Formatter, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	plain, _ := cmd.Flags().GetBool("plain")
	jqExpr, _ := cmd.Flags().GetString("jq")
	return output.New(jsonOut, plain, jqExpr)
}

// ExitCode maps errors to stable codes.
func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}

	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 401, 403:
			return ExitAuth
		case 404:
			return ExitNotFound
		default:
			return ExitError
		}
	}

	return ExitError
}

func renderRootHelp(cmd *cobra.Command) {
	out := cmd.OutOrStdout()

	description := strings.TrimSpace(cmd.Long)
	if description == "" {
		description = cmd.Short
	}
	if description != "" {
		_, _ = fmt.Fprintln(out, description)
		_, _ = fmt.Fprintln(out)
	}

	_, _ = fmt.Fprintln(out, "Usage:")
	_, _ = fmt.Fprintf(out, "  %s [command]\n\n", cmd.CommandPath())

	for _, group := range cmd.Groups() {
		lines := groupHelpLines(cmd, group.ID)
		if len(lines) == 0 {
			continue
		}
		_, _ = fmt.Fprintln(out, group.Title)
		for _, line := range lines {
			_, _ = fmt.Fprintf(out, "  %-24s %s\n", line.use, line.short)
		}
		_, _ = fmt.Fprintln(out)
	}

	additional := additionalHelpLines(cmd)
	if len(additional) > 0 {
		_, _ = fmt.Fprintln(out, "Additional Commands:")
		for _, line := range additional {
			_, _ = fmt.Fprintf(out, "  %-24s %s\n", line.use, line.short)
		}
		_, _ = fmt.Fprintln(out)
	}

	flagUsages := strings.TrimSpace(cmd.Flags().FlagUsages())
	if flagUsages != "" {
		_, _ = fmt.Fprintln(out, "Flags:")
		_, _ = fmt.Fprintln(out, flagUsages)
		_, _ = fmt.Fprintln(out)
	}

	_, _ = fmt.Fprintf(out, "Use \"%s [command] --help\" for more information about a command.\n", cmd.CommandPath())
}

type helpLine struct {
	use   string
	short string
}

func groupHelpLines(root *cobra.Command, groupID string) []helpLine {
	lines := make([]helpLine, 0)

	if groupID == "general" {
		for _, command := range root.Commands() {
			if command.Name() == "help" && !command.Hidden {
				lines = append(lines, helpLine{use: command.Name(), short: command.Short})
				break
			}
		}
	}

	for _, command := range root.Commands() {
		if command.GroupID != groupID || !command.IsAvailableCommand() || command.Hidden {
			continue
		}

		if isResourceGroup(groupID) {
			subcommands := visibleSubcommands(command)
			if len(subcommands) == 0 {
				lines = append(lines, helpLine{use: command.Name(), short: command.Short})
				continue
			}
			for _, subcommand := range subcommands {
				lines = append(lines, helpLine{
					use:   command.Name() + " " + subcommand.Name(),
					short: subcommand.Short,
				})
			}
			continue
		}

		if groupID == "general" && command.Name() == "help" {
			continue
		}

		lines = append(lines, helpLine{use: command.Name(), short: command.Short})
	}
	return lines
}

func additionalHelpLines(root *cobra.Command) []helpLine {
	lines := make([]helpLine, 0)
	for _, command := range root.Commands() {
		if command.GroupID != "" || !command.IsAvailableCommand() || command.Hidden {
			continue
		}
		lines = append(lines, helpLine{use: command.Name(), short: command.Short})
	}
	return lines
}

func visibleSubcommands(command *cobra.Command) []*cobra.Command {
	list := make([]*cobra.Command, 0)
	for _, subcommand := range command.Commands() {
		if !subcommand.IsAvailableCommand() || subcommand.Hidden {
			continue
		}
		list = append(list, subcommand)
	}
	return list
}

func isResourceGroup(groupID string) bool {
	return groupID == "todos" || groupID == "projects" || groupID == "areas" || groupID == "tags"
}

func newTopicHelpCmd(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [topic|command]",
		Short: "Show command help or topic guides (todos|projects|areas|tags)",
		Long:  "Examples: things3-cli help todos, things3-cli help projects, things3-cli help areas, things3-cli help tags",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return root.Help()
			}

			if len(args) == 1 {
				if doc, ok := topicHelpDoc(strings.ToLower(strings.TrimSpace(args[0]))); ok {
					_, err := fmt.Fprintln(cmd.OutOrStdout(), doc)
					return err
				}
			}

			target, _, err := root.Find(args)
			if err == nil && target != nil && target.Name() != "help" {
				return target.Help()
			}

			message := fmt.Sprintf("unknown help topic %q, available topics: todos, projects, areas, tags", args[0])
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), message)
			return errors.New(message)
		},
	}

	cmd.SetOut(root.OutOrStdout())
	cmd.SetErr(root.ErrOrStderr())
	return cmd
}

func topicHelpDoc(topic string) (string, bool) {
	switch topic {
	case "todo", "todos":
		return todosHelpDoc, true
	case "project", "projects":
		return projectsHelpDoc, true
	case "area", "areas":
		return areasHelpDoc, true
	case "tag", "tags":
		return tagsHelpDoc, true
	default:
		return "", false
	}
}

const todosHelpDoc = `# Todos guide

Overview
- Data model: To-Do is the core actionable item in Things.
- This CLI uses SQLite for read and URL Scheme for create/update.
- Delete is implemented via AppleScript (macOS only).

CLI commands
- things3-cli add-todo ...
- things3-cli ls-todo [full filters]
- things3-cli inbox|today|upcoming|anytime|someday [filters]
- things3-cli get-todo <id>
- things3-cli update-todo --id <id> ...
- things3-cli delete-todo --id <id> | --name <title>

Key best practices
1) Prefer ID-based operations for reliability.
2) Use --json (+ --jq) in automation scripts.
3) ls/add-todo support project/area by name directly.
4) For tag filters in ls, use --tags with comma-separated names (AND match).
5) Keep reads from SQLite and writes from URL Scheme/AppleScript.

Important constraints (from local docs + Things docs)
- URL Scheme cannot delete to-dos; AppleScript can.
- Repeating to-dos have update limitations (when/deadline/completed).
- Checklist single-item edit is not supported by URL Scheme.

Source references
- docs/things3-concepts_zh.md
- docs/url-scheme_zh.md
- docs/operations-matrix_zh.md
- Official URL Scheme: https://culturedcode.com/things/support/articles/2803573/
- Official AppleScript: https://culturedcode.com/things/support/articles/4562654/
`

const projectsHelpDoc = `# Projects guide

Overview
- Project groups multiple to-dos and can include headings.
- This CLI supports project create/read/update/delete.
- Create/update uses URL Scheme; delete uses AppleScript.

CLI commands
- things3-cli projects create ...
- things3-cli projects list|ls [filters]
- things3-cli projects get <id>
- things3-cli projects update --id <id> ...
- things3-cli projects delete --id <id> | --name <title>

Key best practices
1) Use area assignment early for better organization.
2) Use tags as context, not as primary hierarchy.
3) Prefer --id for destructive actions.
4) Validate project completion preconditions before setting completed/canceled.

Important constraints
- URL Scheme cannot delete projects.
- update-project cannot append child to-dos directly.
- Heading management is limited in URL Scheme and better handled via Shortcuts.

Source references
- docs/operations-matrix_zh.md
- docs/url-scheme_zh.md
- docs/shortcuts-actions_zh.md
- Official URL Scheme: https://culturedcode.com/things/support/articles/2803573/
- Official Shortcuts: https://culturedcode.com/things/support/articles/9596775/
`

const areasHelpDoc = `# Areas guide

Overview
- Area is a long-lived responsibility domain (e.g., Work, Personal, Health).
- URL Scheme does not manage areas directly.
- This CLI implements area create/update/delete via AppleScript.

CLI commands
- things3-cli areas create --name <name> [--tags "..."]
- things3-cli areas list|ls [filters]
- things3-cli areas get <id>
- things3-cli areas update --id <id>|--name <name> [--new-name ...] [--tags ...]
- things3-cli areas delete --id <id>|--name <name>

Key best practices
1) Use areas for stable domains, not short-term tasks.
2) Keep area names stable and concise.
3) Prefer --id for update/delete when available.
4) In scripts, pair area reads with --json output.

Important constraints
- URL Scheme cannot create/update/delete areas.
- Area operations here require macOS + AppleScript runtime.

Source references
- docs/applescript_zh.md
- docs/operations-matrix_zh.md
- docs/other-integration_zh.md
- Official AppleScript intro: https://culturedcode.com/things/support/articles/2803572/
- Official AppleScript reference: https://culturedcode.com/things/support/articles/4562654/
`

const tagsHelpDoc = `# Tags guide

Overview
- Tag is a cross-cutting context label and supports hierarchy (parent tag).
- URL Scheme can reference existing tags but cannot create/update/delete tag definitions.
- This CLI implements tag create/update/delete via AppleScript.

CLI commands
- things3-cli tags create --name <name> [--parent-name <name>|--parent-id <id>]
- things3-cli tags list|ls [filters]
- things3-cli tags get <id>
- things3-cli tags update --id <id>|--name <name> [--new-name <name>] [--parent-name <name>|--parent-id <id>]
- things3-cli tags delete --id <id>|--name <name>

Key best practices
1) Use stable top-level tags and keep hierarchy shallow.
2) tags list output is grouped by parent for readability.
3) Use --id for update/delete in scripts.
4) Avoid name collisions when multiple tags have similar names.
5) Remember URL-based create/update only applies existing tags.

Important constraints
- URL Scheme does not manage tag definitions.
- Tag CRUD here requires macOS + AppleScript runtime.

Source references
- docs/applescript_zh.md
- docs/operations-matrix_zh.md
- docs/url-scheme_zh.md
- Official AppleScript reference: https://culturedcode.com/things/support/articles/4562654/
- Official URL Scheme: https://culturedcode.com/things/support/articles/2803573/
`
