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

func (f *Formatter) printHuman(w io.Writer, data interface{}) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	showHeader := !f.plain

	switch value := data.(type) {
	case *model.Status:
		if showHeader {
			fmt.Fprintln(tw, "FIELD\tVALUE")
		}
		fmt.Fprintf(tw, "ok\t%t\n", value.OK)
		fmt.Fprintf(tw, "database_path\t%s\n", value.DatabasePath)
		fmt.Fprintf(tw, "database_version\t%s\n", value.DatabaseVersion)
		fmt.Fprintf(tw, "url_scheme_command_available\t%t\n", value.URLSchemeCommandAvailable)
		fmt.Fprintf(tw, "token_configured\t%t\n", value.TokenConfigured)
		return tw.Flush()

	case *model.URLCommandResult:
		if showHeader {
			fmt.Fprintln(tw, "COMMAND\tDISPATCHED\tURL\tMESSAGE")
		}
		fmt.Fprintf(tw, "%s\t%t\t%s\t%s\n", value.Command, value.Dispatched, value.URL, value.Message)
		return tw.Flush()

	case *model.ToDo:
		if showHeader {
			fmt.Fprintln(tw, "ID\tTITLE\tSTATUS\tSTART\tDEADLINE\tPROJECT\tAREA\tTAGS")
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			value.ID, value.Title, value.Status, value.StartDateOrStart(), value.Deadline, value.Project, value.Area, strings.Join(value.Tags, ","))
		return tw.Flush()

	case *model.Project:
		if showHeader {
			fmt.Fprintln(tw, "ID\tTITLE\tSTATUS\tSTART_DATE\tDEADLINE\tAREA\tOPEN/TOTAL\tTAGS")
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%d/%d\t%s\n",
			value.ID, value.Title, value.Status, value.StartDate, value.Deadline, value.Area, value.OpenTaskCount, value.TaskCount, strings.Join(value.Tags, ","))
		return tw.Flush()

	case *model.Area:
		if showHeader {
			fmt.Fprintln(tw, "ID\tNAME\tTAGS")
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", value.ID, value.Name, strings.Join(value.Tags, ","))
		return tw.Flush()

	case *model.Tag:
		if showHeader {
			fmt.Fprintln(tw, "ID\tNAME\tSHORTCUT\tPARENT")
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", value.ID, value.Name, value.Shortcut, value.Parent)
		return tw.Flush()

	case *model.PaginatedResponse[model.ToDo]:
		if showHeader {
			fmt.Fprintln(tw, "ID\tTITLE\tSTATUS\tSTART\tDEADLINE\tPROJECT\tAREA\tTAGS")
		}
		for _, item := range value.Results {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				item.ID, item.Title, item.Status, item.StartDateOrStart(), item.Deadline, item.Project, item.Area, strings.Join(item.Tags, ","))
		}
		return tw.Flush()

	case *model.PaginatedResponse[model.Project]:
		if showHeader {
			fmt.Fprintln(tw, "ID\tTITLE\tSTATUS\tSTART_DATE\tDEADLINE\tAREA\tOPEN/TOTAL\tTAGS")
		}
		for _, item := range value.Results {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%d/%d\t%s\n",
				item.ID, item.Title, item.Status, item.StartDate, item.Deadline, item.Area, item.OpenTaskCount, item.TaskCount, strings.Join(item.Tags, ","))
		}
		return tw.Flush()

	case *model.PaginatedResponse[model.Area]:
		if showHeader {
			fmt.Fprintln(tw, "ID\tNAME\tTAGS")
		}
		for _, item := range value.Results {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", item.ID, item.Name, strings.Join(item.Tags, ","))
		}
		return tw.Flush()

	case *model.PaginatedResponse[model.Tag]:
		if f.plain {
			for _, item := range value.Results {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", item.ID, item.Name, item.Shortcut, item.Parent)
			}
			return tw.Flush()
		}

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
				fmt.Fprintln(tw, "ID\tNAME\tSHORTCUT")
				currentParent = parent
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\n", item.ID, item.Name, item.Shortcut)
		}
		if showHeader && len(value.Results) == 0 {
			fmt.Fprintln(tw, "ID\tNAME\tSHORTCUT\tPARENT")
		}
		return tw.Flush()

	default:
		encoder := json.NewEncoder(tw)
		return encoder.Encode(data)
	}
}
