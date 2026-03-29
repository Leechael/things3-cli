package client

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

type fakeRunner struct {
	lastName string
	lastArgs []string
	output   []byte
	err      error
}

func (r *fakeRunner) Run(name string, args ...string) ([]byte, error) {
	r.lastName = name
	r.lastArgs = args
	return r.output, r.err
}

func TestProbeHTTPWithHTTptest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c, err := New(Config{HTTPClient: srv.Client()})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	err = c.ProbeHTTP(srv.URL)
	if err == nil {
		t.Fatalf("expected error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", apiErr.StatusCode)
	}
}

func TestListToDosFromSQLite(t *testing.T) {
	t.Parallel()

	dbPath := setupTestDB(t)
	c, err := New(Config{DBPath: dbPath, CommandRunner: &fakeRunner{}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	resp, err := c.ListToDos(ListToDoParams{Limit: 10})
	if err != nil {
		t.Fatalf("list todos: %v", err)
	}
	if resp.Count != 1 {
		t.Fatalf("expected 1 todo, got %d", resp.Count)
	}
	if got := resp.Results[0].Title; got != "Buy milk" {
		t.Fatalf("unexpected title: %s", got)
	}
	tags := resp.Results[0].Tags
	if len(tags) != 2 || !containsString(tags, "Errand") || !containsString(tags, "Important") {
		t.Fatalf("unexpected tags: %#v", tags)
	}
	if got := resp.Results[0].Project; got != "Home" {
		t.Fatalf("unexpected project: %s", got)
	}
}

func TestUpdateRequiresToken(t *testing.T) {
	t.Parallel()

	c, err := New(Config{CommandRunner: &fakeRunner{}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = c.UpdateToDo(UpdateToDoParams{ID: "todo-1"})
	if err == nil {
		t.Fatalf("expected auth error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", apiErr.StatusCode)
	}
}

func TestListToDosFiltersByNameAndMultipleTagsAND(t *testing.T) {
	t.Parallel()

	dbPath := setupTestDB(t)
	c, err := New(Config{DBPath: dbPath, CommandRunner: &fakeRunner{}})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	matched, err := c.ListToDos(ListToDoParams{
		ProjectName: "Home",
		AreaName:    "Personal",
		Tags:        "Errand,Important",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("list todos with matching tags: %v", err)
	}
	if matched.Count != 1 {
		t.Fatalf("expected 1 todo for AND match, got %d", matched.Count)
	}

	notMatched, err := c.ListToDos(ListToDoParams{
		ProjectName: "Home",
		AreaName:    "Personal",
		Tags:        "Errand,NotExists",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("list todos with non-matching tags: %v", err)
	}
	if notMatched.Count != 0 {
		t.Fatalf("expected 0 todos for AND mismatch, got %d", notMatched.Count)
	}
}

func TestAddToDoDispatchesURL(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	c, err := New(Config{CommandRunner: runner})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	reveal := true
	result, err := c.AddToDo(AddToDoParams{Title: "Test task", Reveal: &reveal})
	if err != nil {
		t.Fatalf("add todo: %v", err)
	}
	if !result.Dispatched {
		t.Fatalf("expected dispatched=true")
	}
	if runner.lastName != "open" {
		t.Fatalf("expected open command, got %s", runner.lastName)
	}
	if len(runner.lastArgs) < 2 {
		t.Fatalf("unexpected args: %#v", runner.lastArgs)
	}
	if !strings.Contains(runner.lastArgs[1], "things:///add?") {
		t.Fatalf("unexpected URL: %s", runner.lastArgs[1])
	}
}

func TestDeleteToDoDispatchesAppleScript(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	c, err := New(Config{CommandRunner: runner})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	result, err := c.DeleteToDo(DeleteToDoParams{Name: "Buy milk"})
	if err != nil {
		t.Fatalf("delete todo: %v", err)
	}
	if !result.Dispatched {
		t.Fatalf("expected dispatched=true")
	}
	if runner.lastName != "osascript" {
		t.Fatalf("expected osascript command, got %s", runner.lastName)
	}
	if len(runner.lastArgs) != 2 || runner.lastArgs[0] != "-e" {
		t.Fatalf("unexpected args: %#v", runner.lastArgs)
	}
	if !strings.Contains(runner.lastArgs[1], "delete to do named \"Buy milk\"") {
		t.Fatalf("unexpected script: %s", runner.lastArgs[1])
	}
}

func TestAreaCRUDDispatchesAppleScript(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	c, err := New(Config{CommandRunner: runner})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if _, err := c.CreateArea(CreateAreaParams{Name: "Health", TagNames: "Personal"}); err != nil {
		t.Fatalf("create area: %v", err)
	}
	if runner.lastName != "osascript" || !strings.Contains(runner.lastArgs[1], "make new area") {
		t.Fatalf("unexpected create script: %#v", runner.lastArgs)
	}

	newName := "Wellness"
	tags := "Personal,Important"
	if _, err := c.UpdateArea(UpdateAreaParams{Name: "Health", NewName: &newName, TagNames: &tags}); err != nil {
		t.Fatalf("update area: %v", err)
	}
	if !strings.Contains(runner.lastArgs[1], "set name of area") || !strings.Contains(runner.lastArgs[1], "set tag names of area") {
		t.Fatalf("unexpected update script: %s", runner.lastArgs[1])
	}

	if _, err := c.DeleteArea(DeleteAreaParams{Name: "Wellness"}); err != nil {
		t.Fatalf("delete area: %v", err)
	}
	if !strings.Contains(runner.lastArgs[1], "delete area named") {
		t.Fatalf("unexpected delete script: %s", runner.lastArgs[1])
	}
}

func TestTagCRUDDispatchesAppleScript(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	c, err := New(Config{CommandRunner: runner})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if _, err := c.CreateTag(CreateTagParams{Name: "Errand", ParentName: "Places"}); err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if runner.lastName != "osascript" || !strings.Contains(runner.lastArgs[1], "make new tag") {
		t.Fatalf("unexpected create script: %#v", runner.lastArgs)
	}

	newName := "Shopping"
	parentName := "Context"
	if _, err := c.UpdateTag(UpdateTagParams{Name: "Errand", NewName: &newName, ParentName: &parentName}); err != nil {
		t.Fatalf("update tag: %v", err)
	}
	if !strings.Contains(runner.lastArgs[1], "set name of tag") || !strings.Contains(runner.lastArgs[1], "set parent tag of tag") {
		t.Fatalf("unexpected update script: %s", runner.lastArgs[1])
	}

	if _, err := c.DeleteTag(DeleteTagParams{Name: "Shopping"}); err != nil {
		t.Fatalf("delete tag: %v", err)
	}
	if !strings.Contains(runner.lastArgs[1], "delete tag") {
		t.Fatalf("unexpected delete script: %s", runner.lastArgs[1])
	}
}

func TestDeleteProjectByIDUsesSQLiteLookup(t *testing.T) {
	t.Parallel()

	dbPath := setupTestDB(t)
	runner := &fakeRunner{}
	c, err := New(Config{DBPath: dbPath, CommandRunner: runner})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if _, err := c.DeleteProject(DeleteProjectParams{ID: "project-1"}); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if runner.lastName != "osascript" {
		t.Fatalf("expected osascript command, got %s", runner.lastName)
	}
	if !strings.Contains(runner.lastArgs[1], "delete project named \"Home\"") {
		t.Fatalf("unexpected script: %s", runner.lastArgs[1])
	}
}

func TestDeleteTagByIDUsesSQLiteLookup(t *testing.T) {
	t.Parallel()

	dbPath := setupTestDB(t)
	runner := &fakeRunner{}
	c, err := New(Config{DBPath: dbPath, CommandRunner: runner})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	if _, err := c.DeleteTag(DeleteTagParams{ID: "tag-1"}); err != nil {
		t.Fatalf("delete tag: %v", err)
	}
	if runner.lastName != "osascript" {
		t.Fatalf("expected osascript command, got %s", runner.lastName)
	}
	if !strings.Contains(runner.lastArgs[1], "delete tag \"Errand\"") {
		t.Fatalf("unexpected script: %s", runner.lastArgs[1])
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func setupTestDB(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "things-test.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	stmts := []string{
		`CREATE TABLE Meta (key TEXT PRIMARY KEY, value TEXT);`,
		`CREATE TABLE TMArea (uuid TEXT PRIMARY KEY, title TEXT, "index" INTEGER);`,
		`CREATE TABLE TMTag (uuid TEXT PRIMARY KEY, title TEXT, shortcut TEXT, parent TEXT, "index" INTEGER);`,
		`CREATE TABLE TMTask (
			uuid TEXT PRIMARY KEY,
			type INTEGER,
			title TEXT,
			notes TEXT,
			status INTEGER,
			start INTEGER,
			startDate INTEGER,
			deadline INTEGER,
			project TEXT,
			area TEXT,
			heading TEXT,
			trashed INTEGER,
			rt1_recurrenceRule BLOB,
			"index" INTEGER,
			untrashedLeafActionsCount INTEGER,
			openUntrashedLeafActionsCount INTEGER
		);`,
		`CREATE TABLE TMTaskTag (tasks TEXT, tags TEXT);`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec schema: %v", err)
		}
	}

	inserts := []string{
		`INSERT INTO Meta(key, value) VALUES('databaseVersion', '26');`,
		`INSERT INTO TMArea(uuid, title, "index") VALUES('area-1', 'Personal', 1);`,
		`INSERT INTO TMTag(uuid, title, shortcut, parent, "index") VALUES('tag-1', 'Errand', '', '', 1);`,
		`INSERT INTO TMTag(uuid, title, shortcut, parent, "index") VALUES('tag-2', 'Important', '', '', 2);`,
		`INSERT INTO TMTask(uuid, type, title, notes, status, start, startDate, deadline, project, area, heading, trashed, rt1_recurrenceRule, "index", untrashedLeafActionsCount, openUntrashedLeafActionsCount)
		 VALUES('project-1', 1, 'Home', '', 0, 1, NULL, NULL, NULL, 'area-1', NULL, 0, NULL, 1, 2, 1);`,
		`INSERT INTO TMTask(uuid, type, title, notes, status, start, startDate, deadline, project, area, heading, trashed, rt1_recurrenceRule, "index", untrashedLeafActionsCount, openUntrashedLeafActionsCount)
		 VALUES('todo-1', 0, 'Buy milk', '2 liters', 0, 0, NULL, NULL, 'project-1', 'area-1', NULL, 0, NULL, 2, 0, 0);`,
		`INSERT INTO TMTaskTag(tasks, tags) VALUES('todo-1', 'tag-1');`,
		`INSERT INTO TMTaskTag(tasks, tags) VALUES('todo-1', 'tag-2');`,
	}

	for _, stmt := range inserts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec insert: %v", err)
		}
	}

	if err := db.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("stat db: %v", err)
	}

	return dbPath
}
