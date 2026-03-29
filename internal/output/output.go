package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Leechael/things3--cli/internal/model"
	"github.com/itchyny/gojq"
)

// Formatter controls stable output for humans and machines.
type Formatter struct {
	jsonOut bool
	plain   bool
	jqQuery *gojq.Query
}

// New creates a formatter instance.
func New(jsonOut, plain bool, jqExpr string) (*Formatter, error) {
	formatter := &Formatter{jsonOut: jsonOut, plain: plain}
	if jqExpr != "" {
		query, err := gojq.Parse(jqExpr)
		if err != nil {
			return nil, fmt.Errorf("invalid jq expression: %w", err)
		}
		formatter.jqQuery = query
	}
	return formatter, nil
}

// Print outputs structured data.
func (f *Formatter) Print(w io.Writer, data interface{}) error {
	if f.jsonOut {
		return f.printJSON(w, data)
	}
	if f.plain {
		return f.printPlain(w, data)
	}
	return f.printHuman(w, data)
}

// Hint prints non-data output to stderr.
func (f *Formatter) Hint(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// PrintMessage prints plain or JSON messages.
func (f *Formatter) PrintMessage(w io.Writer, msg string) error {
	if f.jsonOut {
		return json.NewEncoder(w).Encode(map[string]string{"message": msg})
	}
	_, err := fmt.Fprintln(w, msg)
	return err
}

func (f *Formatter) printJSON(w io.Writer, data interface{}) error {
	if f.jqQuery != nil {
		return f.applyJQ(w, data)
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (f *Formatter) applyJQ(w io.Writer, data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var input interface{}
	if err := json.Unmarshal(raw, &input); err != nil {
		return err
	}

	iter := f.jqQuery.Run(input)
	for {
		result, ok := iter.Next()
		if !ok {
			return nil
		}
		if err, isErr := result.(error); isErr {
			return fmt.Errorf("jq error: %w", err)
		}
		line, err := json.Marshal(result)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(line)); err != nil {
			return err
		}
	}
}

// printPlain outputs full tab-separated rows with all fields, no headers.
// Intended for scripting: stable column order, all IDs included.
func (f *Formatter) printPlain(w io.Writer, data interface{}) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	switch value := data.(type) {
	case *model.Status:
		fmt.Fprintf(tw, "ok\t%t\n", value.OK)
		fmt.Fprintf(tw, "database_path\t%s\n", value.DatabasePath)
		fmt.Fprintf(tw, "database_version\t%s\n", value.DatabaseVersion)
		fmt.Fprintf(tw, "url_scheme_command_available\t%t\n", value.URLSchemeCommandAvailable)
		fmt.Fprintf(tw, "token_configured\t%t\n", value.TokenConfigured)

	case *model.URLCommandResult:
		fmt.Fprintf(tw, "%s\t%t\t%s\t%s\n", value.Command, value.Dispatched, value.URL, value.Message)

	case *model.ToDo:
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			value.ID, value.Title, value.Status, value.StartDateOrStart(), value.Deadline, value.Project, value.Area, strings.Join(value.Tags, ","))

	case *model.Project:
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%d/%d\t%s\n",
			value.ID, value.Title, value.Status, value.StartDate, value.Deadline, value.Area, value.OpenTaskCount, value.TaskCount, strings.Join(value.Tags, ","))

	case *model.Area:
		fmt.Fprintf(tw, "%s\t%s\t%s\n", value.ID, value.Name, strings.Join(value.Tags, ","))

	case *model.Tag:
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", value.ID, value.Name, value.Shortcut, value.Parent)

	case *model.PaginatedResponse[model.ToDo]:
		for _, item := range value.Results {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				item.ID, item.Title, item.Status, item.StartDateOrStart(), item.Deadline, item.Project, item.Area, strings.Join(item.Tags, ","))
		}

	case *model.PaginatedResponse[model.Project]:
		for _, item := range value.Results {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%d/%d\t%s\n",
				item.ID, item.Title, item.Status, item.StartDate, item.Deadline, item.Area, item.OpenTaskCount, item.TaskCount, strings.Join(item.Tags, ","))
		}

	case *model.PaginatedResponse[model.Area]:
		for _, item := range value.Results {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", item.ID, item.Name, strings.Join(item.Tags, ","))
		}

	case *model.PaginatedResponse[model.Tag]:
		for _, item := range value.Results {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.ID, item.Name, item.Shortcut, item.Parent)
		}

	default:
		return json.NewEncoder(tw).Encode(data)
	}

	return tw.Flush()
}

// printHuman outputs a compact, readable view focused on what matters.
// No IDs. Single items use key-value format. Lists use minimal columns.
func (f *Formatter) printHuman(w io.Writer, data interface{}) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	switch value := data.(type) {
	case *model.Status:
		fmt.Fprintf(tw, "ok\t%t\n", value.OK)
		fmt.Fprintf(tw, "database_path\t%s\n", value.DatabasePath)
		fmt.Fprintf(tw, "database_version\t%s\n", value.DatabaseVersion)
		fmt.Fprintf(tw, "url_scheme_command_available\t%t\n", value.URLSchemeCommandAvailable)
		fmt.Fprintf(tw, "token_configured\t%t\n", value.TokenConfigured)
		return tw.Flush()

	case *model.URLCommandResult:
		fmt.Fprintf(tw, "command\t%s\n", value.Command)
		fmt.Fprintf(tw, "dispatched\t%t\n", value.Dispatched)
		if value.Message != "" {
			fmt.Fprintf(tw, "message\t%s\n", value.Message)
		}
		return tw.Flush()

	case *model.ToDo:
		fmt.Fprintf(tw, "id\t%s\n", value.ID)
		fmt.Fprintf(tw, "title\t%s\n", value.Title)
		fmt.Fprintf(tw, "status\t%s\n", value.Status)
		if s := value.StartDateOrStart(); s != "" {
			fmt.Fprintf(tw, "start\t%s\n", s)
		}
		if value.Deadline != "" {
			fmt.Fprintf(tw, "deadline\t%s\n", value.Deadline)
		}
		if value.Project != "" {
			fmt.Fprintf(tw, "project\t%s\n", value.Project)
		}
		if value.Area != "" {
			fmt.Fprintf(tw, "area\t%s\n", value.Area)
		}
		if len(value.Tags) > 0 {
			fmt.Fprintf(tw, "tags\t%s\n", strings.Join(value.Tags, ", "))
		}
		if value.Notes != "" {
			fmt.Fprintf(tw, "notes\t%s\n", value.Notes)
		}
		return tw.Flush()

	case *model.Project:
		fmt.Fprintf(tw, "id\t%s\n", value.ID)
		fmt.Fprintf(tw, "title\t%s\n", value.Title)
		fmt.Fprintf(tw, "status\t%s\n", value.Status)
		if value.Area != "" {
			fmt.Fprintf(tw, "area\t%s\n", value.Area)
		}
		fmt.Fprintf(tw, "tasks\t%d open / %d total\n", value.OpenTaskCount, value.TaskCount)
		if value.Deadline != "" {
			fmt.Fprintf(tw, "deadline\t%s\n", value.Deadline)
		}
		if len(value.Tags) > 0 {
			fmt.Fprintf(tw, "tags\t%s\n", strings.Join(value.Tags, ", "))
		}
		if value.Notes != "" {
			fmt.Fprintf(tw, "notes\t%s\n", value.Notes)
		}
		return tw.Flush()

	case *model.Area:
		fmt.Fprintf(tw, "id\t%s\n", value.ID)
		fmt.Fprintf(tw, "name\t%s\n", value.Name)
		if len(value.Tags) > 0 {
			fmt.Fprintf(tw, "tags\t%s\n", strings.Join(value.Tags, ", "))
		}
		return tw.Flush()

	case *model.Tag:
		fmt.Fprintf(tw, "id\t%s\n", value.ID)
		fmt.Fprintf(tw, "name\t%s\n", value.Name)
		if value.Parent != "" {
			fmt.Fprintf(tw, "parent\t%s\n", value.Parent)
		}
		if value.Shortcut != "" {
			fmt.Fprintf(tw, "shortcut\t%s\n", value.Shortcut)
		}
		return tw.Flush()

	case *model.PaginatedResponse[model.ToDo]:
		fmt.Fprintln(tw, "TITLE\tPROJECT\tDEADLINE")
		for _, item := range value.Results {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", item.Title, item.Project, item.Deadline)
		}
		return tw.Flush()

	case *model.PaginatedResponse[model.Project]:
		fmt.Fprintln(tw, "TITLE\tAREA\tOPEN/TOTAL")
		for _, item := range value.Results {
			fmt.Fprintf(tw, "%s\t%s\t%d/%d\n", item.Title, item.Area, item.OpenTaskCount, item.TaskCount)
		}
		return tw.Flush()

	case *model.PaginatedResponse[model.Area]:
		fmt.Fprintln(tw, "NAME")
		for _, item := range value.Results {
			fmt.Fprintln(tw, item.Name)
		}
		return tw.Flush()

	case *model.PaginatedResponse[model.Tag]:
		currentParent := ""
		for index, item := range value.Results {
			parent := item.Parent
			if parent == "" {
				parent = "(root)"
			}
			if index == 0 || parent != currentParent {
				if index > 0 {
					fmt.Fprintln(tw)
				}
				fmt.Fprintf(tw, "[%s]\n", parent)
				fmt.Fprintln(tw, "NAME\tSHORTCUT")
				currentParent = parent
			}
			fmt.Fprintf(tw, "%s\t%s\n", item.Name, item.Shortcut)
		}
		if len(value.Results) == 0 {
			fmt.Fprintln(tw, "NAME\tSHORTCUT")
		}
		return tw.Flush()

	default:
		return json.NewEncoder(tw).Encode(data)
	}
}
