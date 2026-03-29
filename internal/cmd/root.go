package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	helpdocs "github.com/Leechael/things3--cli/docs/help"
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
		Long:          "Things3 CLI — manage to-dos, projects, areas, and tags from the command line.",
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

	topics := helpTopicList()
	if len(topics) > 0 {
		_, _ = fmt.Fprintln(out, "Help Topics:")
		for _, line := range topics {
			_, _ = fmt.Fprintf(out, "  %-24s %s\n", line.use, line.short)
		}
		_, _ = fmt.Fprintln(out)
	}

	_, _ = fmt.Fprintf(out, "Use \"%s [command] --help\" for more information about a command.\n", cmd.CommandPath())
	_, _ = fmt.Fprintf(out, "Use \"%s help <topic>\" for topic guides and reference.\n", cmd.CommandPath())
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
		Short: "Show command help or topic guides",
		Long:  "Use \"things3-cli help <topic>\" for topic guides. Use \"things3-cli help <command>\" for command usage.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return root.Help()
			}

			if len(args) == 1 {
				arg := strings.ToLower(strings.TrimSpace(args[0]))

				if strings.HasPrefix(arg, "errcode-") {
					code := strings.TrimPrefix(arg, "errcode-")
					if doc, ok := errorHelpDoc(code); ok {
						_, err := fmt.Fprintln(cmd.OutOrStdout(), doc)
						return err
					}
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "no help topic for errcode %q\n", code)
					return fmt.Errorf("no help topic for errcode %q", code)
				}

				if doc, ok := topicHelpDoc(arg); ok {
					_, err := fmt.Fprintln(cmd.OutOrStdout(), doc)
					return err
				}
			}

			target, _, err := root.Find(args)
			if err == nil && target != nil && target.Name() != "help" {
				return target.Help()
			}

			available := availableTopics()
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "unknown help topic %q\n\nAvailable topics: %s\n", args[0], strings.Join(available, ", "))
			return fmt.Errorf("unknown help topic %q", args[0])
		},
	}

	cmd.SetOut(root.OutOrStdout())
	cmd.SetErr(root.ErrOrStderr())
	return cmd
}

func topicHelpDoc(topic string) (string, bool) {
	aliases := map[string]string{
		"todo":    "todos",
		"project": "projects",
		"area":    "areas",
		"tag":     "tags",
	}
	if canonical, ok := aliases[topic]; ok {
		topic = canonical
	}
	data, err := helpdocs.FS.ReadFile("topics/" + topic + ".md")
	if err != nil {
		return "", false
	}
	return string(data), true
}

func errorHelpDoc(code string) (string, bool) {
	data, err := helpdocs.FS.ReadFile("errors/" + code + ".md")
	if err != nil {
		return "", false
	}
	return string(data), true
}

func helpTopicList() []helpLine {
	entries, err := helpdocs.FS.ReadDir("topics")
	if err != nil {
		return nil
	}
	var lines []helpLine
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		desc := topicFirstLine("topics/" + entry.Name())
		lines = append(lines, helpLine{use: name, short: desc})
	}
	return lines
}

func topicFirstLine(path string) string {
	data, err := helpdocs.FS.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}

func availableTopics() []string {
	lines := helpTopicList()
	names := make([]string, len(lines))
	for i, l := range lines {
		names[i] = l.use
	}
	return names
}
