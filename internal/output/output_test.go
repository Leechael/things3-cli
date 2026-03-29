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

func TestFormatterTagListTree(t *testing.T) {
	t.Parallel()

	formatter, err := New(false, false, "")
	if err != nil {
		t.Fatalf("new formatter: %v", err)
	}

	payload := &model.PaginatedResponse[model.Tag]{
		Count: 4,
		Results: []model.Tag{
			{ID: "ctx-1", Name: "Context"},
			{ID: "tag-1", Name: "Errand", ParentID: "ctx-1", Parent: "Context"},
			{ID: "tag-2", Name: "Home", ParentID: "ctx-1", Parent: "Context"},
			{ID: "tag-3", Name: "Untagged"},
		},
	}

	var out bytes.Buffer
	if err := formatter.Print(&out, payload); err != nil {
		t.Fatalf("print: %v", err)
	}

	got := out.String()
	// Root tags have no connector prefix
	if !strings.Contains(got, "[ctx-1] Context\n") {
		t.Fatalf("expected root tag line without connector, got: %q", got)
	}
	// Non-last child uses ├──
	if !strings.Contains(got, "├── [tag-1] Errand") {
		t.Fatalf("expected non-last child with ├── connector, got: %q", got)
	}
	// Last child uses └──
	if !strings.Contains(got, "└── [tag-2] Home") {
		t.Fatalf("expected last child with └── connector, got: %q", got)
	}
	// Another root tag
	if !strings.Contains(got, "[tag-3] Untagged\n") {
		t.Fatalf("expected root tag line for Untagged, got: %q", got)
	}
}
