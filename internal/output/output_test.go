package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Leechael/things3--cli/internal/model"
)

func TestFormatterJSONAndJQ(t *testing.T) {
	t.Parallel()

	formatter, err := New(true, false, ".count")
	if err != nil {
		t.Fatalf("new formatter: %v", err)
	}

	payload := &model.PaginatedResponse[model.ToDo]{
		Count:   1,
		Results: []model.ToDo{{ID: "todo-1", Title: "Buy milk", Status: "incomplete"}},
	}

	var out bytes.Buffer
	if err := formatter.Print(&out, payload); err != nil {
		t.Fatalf("print: %v", err)
	}
	if strings.TrimSpace(out.String()) != "1" {
		t.Fatalf("unexpected jq output: %q", out.String())
	}
}

func TestFormatterPlainList(t *testing.T) {
	t.Parallel()

	formatter, err := New(false, true, "")
	if err != nil {
		t.Fatalf("new formatter: %v", err)
	}

	payload := &model.PaginatedResponse[model.Area]{
		Count: 1,
		Results: []model.Area{
			{ID: "area-1", Name: "Personal", Tags: []string{"Errand"}},
		},
	}

	var out bytes.Buffer
	if err := formatter.Print(&out, payload); err != nil {
		t.Fatalf("print: %v", err)
	}

	got := strings.TrimSpace(out.String())
	if strings.Contains(got, "ID") {
		t.Fatalf("plain output should not include header, got: %q", got)
	}
	if !strings.Contains(got, "area-1") || !strings.Contains(got, "Personal") || !strings.Contains(got, "Errand") {
		t.Fatalf("unexpected plain output: %q", got)
	}
}

func TestFormatterTagListGroupedByParent(t *testing.T) {
	t.Parallel()

	formatter, err := New(false, false, "")
	if err != nil {
		t.Fatalf("new formatter: %v", err)
	}

	payload := &model.PaginatedResponse[model.Tag]{
		Count: 3,
		Results: []model.Tag{
			{ID: "tag-1", Name: "Errand", Parent: "Context"},
			{ID: "tag-2", Name: "Home", Parent: "Places"},
			{ID: "tag-3", Name: "Untagged"},
		},
	}

	var out bytes.Buffer
	if err := formatter.Print(&out, payload); err != nil {
		t.Fatalf("print: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "[Context]") || !strings.Contains(got, "[Places]") || !strings.Contains(got, "[(root)]") {
		t.Fatalf("expected grouped parent sections, got: %q", got)
	}
}
