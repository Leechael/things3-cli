package client

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Leechael/things3--cli/internal/model"
	_ "modernc.org/sqlite"
)

const (
	defaultDBPatternNew    = "~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite"
	defaultDBPatternLegacy = "~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite"
)

// CommandRunner executes local commands.
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
}

type osCommandRunner struct{}

func (osCommandRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

// Config configures the client.
type Config struct {
	Token         string
	DBPath        string
	CommandRunner CommandRunner
	HTTPClient    *http.Client
}

// Client wraps Things3 local integrations.
type Client struct {
	token      string
	dbPath     string
	runner     CommandRunner
	httpClient *http.Client
}

// APIError is used for stable exit code mapping.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error %d: %s", e.StatusCode, e.Message)
}

// New creates a new client.
func New(cfg Config) (*Client, error) {
	runner := cfg.CommandRunner
	if runner == nil {
		runner = osCommandRunner{}
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		token:      cfg.Token,
		dbPath:     cfg.DBPath,
		runner:     runner,
		httpClient: httpClient,
	}, nil
}

// ProbeHTTP is primarily used in tests and diagnostics.
func (c *Client) ProbeHTTP(rawURL string) error {
	resp, err := c.httpClient.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return &APIError{StatusCode: resp.StatusCode, Message: http.StatusText(resp.StatusCode)}
	}

	return nil
}

// GetStatus checks local Things3 integration prerequisites.
func (c *Client) GetStatus() (*model.Status, error) {
	db, dbPath, err := c.openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var dbVersion string
	if err := db.QueryRow("SELECT value FROM Meta WHERE key = 'databaseVersion'").Scan(&dbVersion); err != nil {
		return nil, fmt.Errorf("read database version: %w", err)
	}
	dbVersion = normalizeDatabaseVersion(dbVersion)

	_, err = exec.LookPath("open")
	urlCommandAvailable := err == nil

	return &model.Status{
		OK:                        urlCommandAvailable,
		DatabasePath:              dbPath,
		DatabaseVersion:           dbVersion,
		URLSchemeCommandAvailable: urlCommandAvailable,
		TokenConfigured:           c.token != "",
	}, nil
}

// ListToDoParams controls todo list filtering.
type ListToDoParams struct {
	ID             string
	Status         string
	ProjectID      string
	ProjectName    string
	AreaID         string
	AreaName       string
	HeadingID      string
	Tag            string
	Tags           string
	Search         string
	IncludeTrashed bool
	Limit          int
	Offset         int
}

func (p ListToDoParams) encode() url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "id", p.ID)
	setIfNotEmpty(values, "status", p.Status)
	setIfNotEmpty(values, "project-id", p.ProjectID)
	setIfNotEmpty(values, "project", p.ProjectName)
	setIfNotEmpty(values, "area-id", p.AreaID)
	setIfNotEmpty(values, "area", p.AreaName)
	setIfNotEmpty(values, "heading-id", p.HeadingID)
	setIfNotEmpty(values, "tag", p.Tag)
	setIfNotEmpty(values, "tags", p.Tags)
	setIfNotEmpty(values, "search", p.Search)
	if p.IncludeTrashed {
		values.Set("include-trashed", "true")
	}
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		values.Set("offset", strconv.Itoa(p.Offset))
	}
	return values
}

// ListToDos lists to-dos from the local SQLite database.
func (c *Client) ListToDos(params ListToDoParams) (*model.PaginatedResponse[model.ToDo], error) {
	db, _, err := c.openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	statusCode, err := mapStatus(params.Status)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 200
	}

	var query strings.Builder
	query.WriteString(`
SELECT
  t.uuid,
  t.title,
  COALESCE(t.notes, ''),
  t.status,
  t.start,
  CASE WHEN t.startDate IS NOT NULL THEN
    printf('%04d-%02d-%02d',
      (t.startDate >> 16) & 2047,
      (t.startDate >> 12) & 15,
      (t.startDate >> 7) & 31)
  END AS start_date,
  CASE WHEN t.deadline IS NOT NULL THEN
    printf('%04d-%02d-%02d',
      (t.deadline >> 16) & 2047,
      (t.deadline >> 12) & 15,
      (t.deadline >> 7) & 31)
  END AS deadline,
  t.project,
  COALESCE(p.title, ''),
  t.area,
  COALESCE(a.title, ''),
  t.heading,
  COALESCE(h.title, ''),
  COALESCE(GROUP_CONCAT(DISTINCT tag.title), ''),
  CASE WHEN t.rt1_recurrenceRule IS NOT NULL THEN 1 ELSE 0 END AS is_recurring
FROM TMTask t
LEFT JOIN TMTask p ON p.uuid = t.project
LEFT JOIN TMArea a ON a.uuid = t.area
LEFT JOIN TMTask h ON h.uuid = t.heading
LEFT JOIN TMTaskTag tt ON tt.tasks = t.uuid
LEFT JOIN TMTag tag ON tag.uuid = tt.tags
WHERE t.type = 0
`)

	args := make([]any, 0, 12)
	if !params.IncludeTrashed {
		query.WriteString(" AND IFNULL(t.trashed, 0) = 0")
	}
	if params.ID != "" {
		query.WriteString(" AND t.uuid = ?")
		args = append(args, params.ID)
	}
	if statusCode >= 0 {
		query.WriteString(" AND t.status = ?")
		args = append(args, statusCode)
	}
	if params.ProjectID != "" {
		query.WriteString(" AND t.project = ?")
		args = append(args, params.ProjectID)
	}
	if params.ProjectName != "" {
		query.WriteString(" AND p.title = ?")
		args = append(args, params.ProjectName)
	}
	if params.AreaID != "" {
		query.WriteString(" AND t.area = ?")
		args = append(args, params.AreaID)
	}
	if params.AreaName != "" {
		query.WriteString(" AND a.title = ?")
		args = append(args, params.AreaName)
	}
	if params.HeadingID != "" {
		query.WriteString(" AND t.heading = ?")
		args = append(args, params.HeadingID)
	}
	tagFilters := splitCommaValues(params.Tags)
	if params.Tag != "" {
		tagFilters = append(tagFilters, params.Tag)
	}
	tagFilters = uniqueStrings(tagFilters)
	for _, tagName := range tagFilters {
		query.WriteString(`
  AND EXISTS (
    SELECT 1
    FROM TMTaskTag ft
    JOIN TMTag ftag ON ftag.uuid = ft.tags
    WHERE ft.tasks = t.uuid AND ftag.title = ?
  )`)
		args = append(args, tagName)
	}
	if params.Search != "" {
		query.WriteString(" AND (t.title LIKE ? OR t.notes LIKE ?)")
		like := "%" + params.Search + "%"
		args = append(args, like, like)
	}

	query.WriteString(`
GROUP BY
  t.uuid, t.title, t.notes, t.status, t.start, t.startDate, t.deadline,
  t.project, p.title, t.area, a.title, t.heading, h.title, t.rt1_recurrenceRule
ORDER BY t."index"
LIMIT ? OFFSET ?`)
	args = append(args, limit, params.Offset)

	rows, err := db.Query(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query todos: %w", err)
	}
	defer rows.Close()

	items := make([]model.ToDo, 0)
	for rows.Next() {
		var (
			item         model.ToDo
			statusValue  int
			startValue   int
			startDate    sql.NullString
			deadline     sql.NullString
			projectID    sql.NullString
			projectTitle string
			areaID       sql.NullString
			areaTitle    string
			headingID    sql.NullString
			headingTitle string
			tagsRaw      string
			recurringInt int
		)

		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Notes,
			&statusValue,
			&startValue,
			&startDate,
			&deadline,
			&projectID,
			&projectTitle,
			&areaID,
			&areaTitle,
			&headingID,
			&headingTitle,
			&tagsRaw,
			&recurringInt,
		); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}

		item.Status = statusLabel(statusValue)
		item.Start = startLabel(startValue)
		item.StartDate = nullStringValue(startDate)
		item.Deadline = nullStringValue(deadline)
		item.ProjectID = nullStringValue(projectID)
		item.Project = projectTitle
		item.AreaID = nullStringValue(areaID)
		item.Area = areaTitle
		item.HeadingID = nullStringValue(headingID)
		item.Heading = headingTitle
		item.Tags = splitCommaValues(tagsRaw)
		item.IsRecurring = recurringInt == 1

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate todos: %w", err)
	}

	return &model.PaginatedResponse[model.ToDo]{
		Count:   len(items),
		Results: items,
	}, nil
}

// GetToDo reads a single to-do by UUID.
func (c *Client) GetToDo(id string) (*model.ToDo, error) {
	resp, err := c.ListToDos(ListToDoParams{ID: id, IncludeTrashed: true, Limit: 1})
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("to-do not found: %s", id)}
	}
	return &resp.Results[0], nil
}

// ListProjectParams controls project list filtering.
type ListProjectParams struct {
	ID             string
	Status         string
	AreaID         string
	Search         string
	IncludeTrashed bool
	Limit          int
	Offset         int
}

func (p ListProjectParams) encode() url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "id", p.ID)
	setIfNotEmpty(values, "status", p.Status)
	setIfNotEmpty(values, "area-id", p.AreaID)
	setIfNotEmpty(values, "search", p.Search)
	if p.IncludeTrashed {
		values.Set("include-trashed", "true")
	}
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		values.Set("offset", strconv.Itoa(p.Offset))
	}
	return values
}

// ListProjects lists projects from SQLite.
func (c *Client) ListProjects(params ListProjectParams) (*model.PaginatedResponse[model.Project], error) {
	db, _, err := c.openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	statusCode, err := mapStatus(params.Status)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 200
	}

	var query strings.Builder
	query.WriteString(`
SELECT
  t.uuid,
  t.title,
  COALESCE(t.notes, ''),
  t.status,
  CASE WHEN t.startDate IS NOT NULL THEN
    printf('%04d-%02d-%02d',
      (t.startDate >> 16) & 2047,
      (t.startDate >> 12) & 15,
      (t.startDate >> 7) & 31)
  END AS start_date,
  CASE WHEN t.deadline IS NOT NULL THEN
    printf('%04d-%02d-%02d',
      (t.deadline >> 16) & 2047,
      (t.deadline >> 12) & 15,
      (t.deadline >> 7) & 31)
  END AS deadline,
  t.area,
  COALESCE(a.title, ''),
  COALESCE(GROUP_CONCAT(DISTINCT tag.title), ''),
  IFNULL(t.untrashedLeafActionsCount, 0),
  IFNULL(t.openUntrashedLeafActionsCount, 0),
  CASE WHEN t.rt1_recurrenceRule IS NOT NULL THEN 1 ELSE 0 END AS is_recurring
FROM TMTask t
LEFT JOIN TMArea a ON a.uuid = t.area
LEFT JOIN TMTaskTag tt ON tt.tasks = t.uuid
LEFT JOIN TMTag tag ON tag.uuid = tt.tags
WHERE t.type = 1
`)

	args := make([]any, 0, 10)
	if !params.IncludeTrashed {
		query.WriteString(" AND IFNULL(t.trashed, 0) = 0")
	}
	if params.ID != "" {
		query.WriteString(" AND t.uuid = ?")
		args = append(args, params.ID)
	}
	if statusCode >= 0 {
		query.WriteString(" AND t.status = ?")
		args = append(args, statusCode)
	}
	if params.AreaID != "" {
		query.WriteString(" AND t.area = ?")
		args = append(args, params.AreaID)
	}
	if params.Search != "" {
		query.WriteString(" AND (t.title LIKE ? OR t.notes LIKE ?)")
		like := "%" + params.Search + "%"
		args = append(args, like, like)
	}

	query.WriteString(`
GROUP BY
  t.uuid, t.title, t.notes, t.status, t.startDate, t.deadline,
  t.area, a.title, t.untrashedLeafActionsCount, t.openUntrashedLeafActionsCount,
  t.rt1_recurrenceRule
ORDER BY t."index"
LIMIT ? OFFSET ?`)
	args = append(args, limit, params.Offset)

	rows, err := db.Query(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}
	defer rows.Close()

	items := make([]model.Project, 0)
	for rows.Next() {
		var (
			item         model.Project
			statusValue  int
			startDate    sql.NullString
			deadline     sql.NullString
			areaID       sql.NullString
			areaTitle    string
			tagsRaw      string
			recurringInt int
		)

		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Notes,
			&statusValue,
			&startDate,
			&deadline,
			&areaID,
			&areaTitle,
			&tagsRaw,
			&item.TaskCount,
			&item.OpenTaskCount,
			&recurringInt,
		); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}

		item.Status = statusLabel(statusValue)
		item.StartDate = nullStringValue(startDate)
		item.Deadline = nullStringValue(deadline)
		item.AreaID = nullStringValue(areaID)
		item.Area = areaTitle
		item.Tags = splitCommaValues(tagsRaw)
		item.DoneTaskCount = item.TaskCount - item.OpenTaskCount
		item.IsRecurring = recurringInt == 1
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}

	return &model.PaginatedResponse[model.Project]{
		Count:   len(items),
		Results: items,
	}, nil
}

// GetProject reads a single project by UUID.
func (c *Client) GetProject(id string) (*model.Project, error) {
	resp, err := c.ListProjects(ListProjectParams{ID: id, IncludeTrashed: true, Limit: 1})
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("project not found: %s", id)}
	}
	return &resp.Results[0], nil
}

// ListAreaParams controls area list filtering.
type ListAreaParams struct {
	ID     string
	Search string
	Limit  int
	Offset int
}

func (p ListAreaParams) encode() url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "id", p.ID)
	setIfNotEmpty(values, "search", p.Search)
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		values.Set("offset", strconv.Itoa(p.Offset))
	}
	return values
}

// ListAreas lists areas from SQLite.
func (c *Client) ListAreas(params ListAreaParams) (*model.PaginatedResponse[model.Area], error) {
	db, _, err := c.openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	limit := params.Limit
	if limit <= 0 {
		limit = 200
	}

	var query strings.Builder
	query.WriteString(`
SELECT
  a.uuid,
  a.title,
  COALESCE(GROUP_CONCAT(DISTINCT tag.title), '')
FROM TMArea a
LEFT JOIN TMAreaTag atag ON atag.areas = a.uuid
LEFT JOIN TMTag tag ON tag.uuid = atag.tags
WHERE 1 = 1
`)

	args := make([]any, 0, 6)
	if params.ID != "" {
		query.WriteString(" AND a.uuid = ?")
		args = append(args, params.ID)
	}
	if params.Search != "" {
		query.WriteString(" AND a.title LIKE ?")
		args = append(args, "%"+params.Search+"%")
	}

	query.WriteString(`
GROUP BY a.uuid, a.title
ORDER BY a."index"
LIMIT ? OFFSET ?`)
	args = append(args, limit, params.Offset)

	rows, err := db.Query(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query areas: %w", err)
	}
	defer rows.Close()

	items := make([]model.Area, 0)
	for rows.Next() {
		var (
			item    model.Area
			tagsRaw string
		)
		if err := rows.Scan(&item.ID, &item.Name, &tagsRaw); err != nil {
			return nil, fmt.Errorf("scan area: %w", err)
		}
		item.Tags = splitCommaValues(tagsRaw)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate areas: %w", err)
	}

	return &model.PaginatedResponse[model.Area]{
		Count:   len(items),
		Results: items,
	}, nil
}

// GetArea reads a single area by UUID.
func (c *Client) GetArea(id string) (*model.Area, error) {
	resp, err := c.ListAreas(ListAreaParams{ID: id, Limit: 1})
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("area not found: %s", id)}
	}
	return &resp.Results[0], nil
}

// CreateAreaParams controls AppleScript area creation.
type CreateAreaParams struct {
	Name     string
	TagNames string
}

// CreateArea creates a Things area via AppleScript.
func (c *Client) CreateArea(params CreateAreaParams) (*model.URLCommandResult, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return nil, fmt.Errorf("area name is required")
	}

	script := fmt.Sprintf(`tell application "Things3" to make new area with properties {name:%s`, appleScriptString(name))
	if strings.TrimSpace(params.TagNames) != "" {
		script += fmt.Sprintf(`, tag names:%s`, appleScriptString(params.TagNames))
	}
	script += "}"

	return c.runAppleScript("area-create", script)
}

// UpdateAreaParams controls AppleScript area updates.
type UpdateAreaParams struct {
	ID       string
	Name     string
	NewName  *string
	TagNames *string
}

// UpdateArea updates a Things area via AppleScript.
func (c *Client) UpdateArea(params UpdateAreaParams) (*model.URLCommandResult, error) {
	currentName := strings.TrimSpace(params.Name)
	if currentName == "" && strings.TrimSpace(params.ID) != "" {
		area, err := c.GetArea(params.ID)
		if err != nil {
			return nil, err
		}
		currentName = area.Name
	}
	if currentName == "" {
		return nil, fmt.Errorf("area name or --id is required")
	}

	newName := currentName
	commands := make([]string, 0, 2)
	if params.NewName != nil {
		trimmed := strings.TrimSpace(*params.NewName)
		if trimmed == "" {
			return nil, fmt.Errorf("new area name cannot be empty")
		}
		commands = append(commands, fmt.Sprintf("set name of area %s to %s", appleScriptString(currentName), appleScriptString(trimmed)))
		newName = trimmed
	}
	if params.TagNames != nil {
		commands = append(commands, fmt.Sprintf("set tag names of area %s to %s", appleScriptString(newName), appleScriptString(*params.TagNames)))
	}
	if len(commands) == 0 {
		return nil, fmt.Errorf("no area updates requested")
	}

	script := "tell application \"Things3\"\n"
	for _, command := range commands {
		script += "  " + command + "\n"
	}
	script += "end tell"

	return c.runAppleScript("area-update", script)
}

// DeleteAreaParams controls AppleScript area deletion.
type DeleteAreaParams struct {
	ID   string
	Name string
}

// DeleteArea deletes a Things area via AppleScript.
func (c *Client) DeleteArea(params DeleteAreaParams) (*model.URLCommandResult, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" && strings.TrimSpace(params.ID) != "" {
		area, err := c.GetArea(params.ID)
		if err != nil {
			return nil, err
		}
		name = area.Name
	}
	if name == "" {
		return nil, fmt.Errorf("area name or --id is required")
	}

	script := fmt.Sprintf("tell application \"Things3\" to delete area named %s", appleScriptString(name))
	return c.runAppleScript("area-delete", script)
}

// ListTagParams controls tag list filtering.
type ListTagParams struct {
	ID       string
	ParentID string
	Search   string
	Limit    int
	Offset   int
}

func (p ListTagParams) encode() url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "id", p.ID)
	setIfNotEmpty(values, "parent-id", p.ParentID)
	setIfNotEmpty(values, "search", p.Search)
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		values.Set("offset", strconv.Itoa(p.Offset))
	}
	return values
}

// ListTags lists tags from SQLite.
func (c *Client) ListTags(params ListTagParams) (*model.PaginatedResponse[model.Tag], error) {
	db, _, err := c.openDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	limit := params.Limit
	if limit <= 0 {
		limit = 200
	}

	var query strings.Builder
	query.WriteString(`
SELECT
  t.uuid,
  t.title,
  COALESCE(t.shortcut, ''),
  COALESCE(t.parent, ''),
  COALESCE(p.title, '')
FROM TMTag t
LEFT JOIN TMTag p ON p.uuid = t.parent
WHERE 1 = 1
`)

	args := make([]any, 0, 6)
	if params.ID != "" {
		query.WriteString(" AND t.uuid = ?")
		args = append(args, params.ID)
	}
	if params.ParentID != "" {
		query.WriteString(" AND t.parent = ?")
		args = append(args, params.ParentID)
	}
	if params.Search != "" {
		query.WriteString(" AND t.title LIKE ?")
		args = append(args, "%"+params.Search+"%")
	}

	query.WriteString(`
ORDER BY COALESCE(p.title, ''), t."index"
LIMIT ? OFFSET ?`)
	args = append(args, limit, params.Offset)

	rows, err := db.Query(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	items := make([]model.Tag, 0)
	for rows.Next() {
		var item model.Tag
		if err := rows.Scan(&item.ID, &item.Name, &item.Shortcut, &item.ParentID, &item.Parent); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}

	return &model.PaginatedResponse[model.Tag]{
		Count:   len(items),
		Results: items,
	}, nil
}

// GetTag reads a single tag by UUID.
func (c *Client) GetTag(id string) (*model.Tag, error) {
	resp, err := c.ListTags(ListTagParams{ID: id, Limit: 1})
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("tag not found: %s", id)}
	}
	return &resp.Results[0], nil
}

// CreateTagParams controls AppleScript tag creation.
type CreateTagParams struct {
	Name       string
	ParentName string
	ParentID   string
}

// CreateTag creates a tag via AppleScript.
func (c *Client) CreateTag(params CreateTagParams) (*model.URLCommandResult, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return nil, fmt.Errorf("tag name is required")
	}

	parentName, err := c.resolveTagParentName(params.ParentName, params.ParentID)
	if err != nil {
		return nil, err
	}

	commands := []string{fmt.Sprintf("make new tag with properties {name:%s}", appleScriptString(name))}
	if parentName != "" {
		commands = append(commands, fmt.Sprintf("set parent tag of tag %s to tag %s", appleScriptString(name), appleScriptString(parentName)))
	}

	script := "tell application \"Things3\"\n"
	for _, command := range commands {
		script += "  " + command + "\n"
	}
	script += "end tell"

	return c.runAppleScript("tag-create", script)
}

// UpdateTagParams controls AppleScript tag updates.
type UpdateTagParams struct {
	ID         string
	Name       string
	NewName    *string
	ParentName *string
	ParentID   *string
}

// UpdateTag updates a tag via AppleScript.
func (c *Client) UpdateTag(params UpdateTagParams) (*model.URLCommandResult, error) {
	currentName := strings.TrimSpace(params.Name)
	if currentName == "" && strings.TrimSpace(params.ID) != "" {
		tag, err := c.GetTag(params.ID)
		if err != nil {
			return nil, err
		}
		currentName = tag.Name
	}
	if currentName == "" {
		return nil, fmt.Errorf("tag name or --id is required")
	}

	newName := currentName
	commands := make([]string, 0, 2)
	if params.NewName != nil {
		trimmed := strings.TrimSpace(*params.NewName)
		if trimmed == "" {
			return nil, fmt.Errorf("new tag name cannot be empty")
		}
		commands = append(commands, fmt.Sprintf("set name of tag %s to %s", appleScriptString(currentName), appleScriptString(trimmed)))
		newName = trimmed
	}

	if params.ParentName != nil || params.ParentID != nil {
		parentName := ""
		if params.ParentName != nil {
			parentName = strings.TrimSpace(*params.ParentName)
		}
		parentID := ""
		if params.ParentID != nil {
			parentID = strings.TrimSpace(*params.ParentID)
		}
		resolvedParentName, err := c.resolveTagParentName(parentName, parentID)
		if err != nil {
			return nil, err
		}
		if resolvedParentName == "" {
			return nil, fmt.Errorf("parent tag must be set via --parent-name or --parent-id")
		}
		commands = append(commands, fmt.Sprintf("set parent tag of tag %s to tag %s", appleScriptString(newName), appleScriptString(resolvedParentName)))
	}

	if len(commands) == 0 {
		return nil, fmt.Errorf("no tag updates requested")
	}

	script := "tell application \"Things3\"\n"
	for _, command := range commands {
		script += "  " + command + "\n"
	}
	script += "end tell"

	return c.runAppleScript("tag-update", script)
}

// DeleteTagParams controls AppleScript tag deletion.
type DeleteTagParams struct {
	ID   string
	Name string
}

// DeleteTag deletes a tag via AppleScript.
func (c *Client) DeleteTag(params DeleteTagParams) (*model.URLCommandResult, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" && strings.TrimSpace(params.ID) != "" {
		tag, err := c.GetTag(params.ID)
		if err != nil {
			return nil, err
		}
		name = tag.Name
	}
	if name == "" {
		return nil, fmt.Errorf("tag name or --id is required")
	}

	script := fmt.Sprintf("tell application \"Things3\" to delete tag %s", appleScriptString(name))
	return c.runAppleScript("tag-delete", script)
}

func (c *Client) resolveTagParentName(parentName string, parentID string) (string, error) {
	resolved := strings.TrimSpace(parentName)
	if resolved != "" && strings.TrimSpace(parentID) != "" {
		return "", fmt.Errorf("--parent-name and --parent-id are mutually exclusive")
	}
	if resolved != "" {
		return resolved, nil
	}
	if strings.TrimSpace(parentID) == "" {
		return "", nil
	}
	tag, err := c.GetTag(parentID)
	if err != nil {
		return "", err
	}
	return tag.Name, nil
}

// AddToDoParams maps to Things URL command add.
type AddToDoParams struct {
	Title          string
	Titles         string
	Notes          string
	When           string
	Deadline       string
	Tags           string
	ChecklistItems string
	List           string
	ListID         string
	Heading        string
	HeadingID      string
	UseClipboard   string
	CreationDate   string
	CompletionDate string
	Completed      *bool
	Canceled       *bool
	ShowQuickEntry *bool
	Reveal         *bool
}

func (p AddToDoParams) encode() url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "title", p.Title)
	setIfNotEmpty(values, "titles", p.Titles)
	setIfNotEmpty(values, "notes", p.Notes)
	setIfNotEmpty(values, "when", p.When)
	setIfNotEmpty(values, "deadline", p.Deadline)
	setIfNotEmpty(values, "tags", p.Tags)
	setIfNotEmpty(values, "checklist-items", p.ChecklistItems)
	setIfNotEmpty(values, "list", p.List)
	setIfNotEmpty(values, "list-id", p.ListID)
	setIfNotEmpty(values, "heading", p.Heading)
	setIfNotEmpty(values, "heading-id", p.HeadingID)
	setIfNotEmpty(values, "use-clipboard", p.UseClipboard)
	setIfNotEmpty(values, "creation-date", p.CreationDate)
	setIfNotEmpty(values, "completion-date", p.CompletionDate)
	setIfNotNilBool(values, "completed", p.Completed)
	setIfNotNilBool(values, "canceled", p.Canceled)
	setIfNotNilBool(values, "show-quick-entry", p.ShowQuickEntry)
	setIfNotNilBool(values, "reveal", p.Reveal)
	return values
}

// AddToDo dispatches things:///add.
func (c *Client) AddToDo(params AddToDoParams) (*model.URLCommandResult, error) {
	return c.runURLCommand("add", params.encode())
}

// AddProjectParams maps to Things URL command add-project.
type AddProjectParams struct {
	Title          string
	Notes          string
	When           string
	Deadline       string
	Tags           string
	Area           string
	AreaID         string
	ToDos          string
	CreationDate   string
	CompletionDate string
	Completed      *bool
	Canceled       *bool
	Reveal         *bool
}

func (p AddProjectParams) encode() url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "title", p.Title)
	setIfNotEmpty(values, "notes", p.Notes)
	setIfNotEmpty(values, "when", p.When)
	setIfNotEmpty(values, "deadline", p.Deadline)
	setIfNotEmpty(values, "tags", p.Tags)
	setIfNotEmpty(values, "area", p.Area)
	setIfNotEmpty(values, "area-id", p.AreaID)
	setIfNotEmpty(values, "to-dos", p.ToDos)
	setIfNotEmpty(values, "creation-date", p.CreationDate)
	setIfNotEmpty(values, "completion-date", p.CompletionDate)
	setIfNotNilBool(values, "completed", p.Completed)
	setIfNotNilBool(values, "canceled", p.Canceled)
	setIfNotNilBool(values, "reveal", p.Reveal)
	return values
}

// AddProject dispatches things:///add-project.
func (c *Client) AddProject(params AddProjectParams) (*model.URLCommandResult, error) {
	return c.runURLCommand("add-project", params.encode())
}

// UpdateToDoParams maps to Things URL command update.
type UpdateToDoParams struct {
	ID                    string
	Title                 *string
	Notes                 *string
	PrependNotes          *string
	AppendNotes           *string
	When                  *string
	Deadline              *string
	Tags                  *string
	AddTags               *string
	ChecklistItems        *string
	PrependChecklistItems *string
	AppendChecklistItems  *string
	List                  *string
	ListID                *string
	Heading               *string
	HeadingID             *string
	CreationDate          *string
	CompletionDate        *string
	Completed             *bool
	Canceled              *bool
	Duplicate             *bool
	Reveal                *bool
}

func (p UpdateToDoParams) encode() url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "id", p.ID)
	setIfNotNilString(values, "title", p.Title)
	setIfNotNilString(values, "notes", p.Notes)
	setIfNotNilString(values, "prepend-notes", p.PrependNotes)
	setIfNotNilString(values, "append-notes", p.AppendNotes)
	setIfNotNilString(values, "when", p.When)
	setIfNotNilString(values, "deadline", p.Deadline)
	setIfNotNilString(values, "tags", p.Tags)
	setIfNotNilString(values, "add-tags", p.AddTags)
	setIfNotNilString(values, "checklist-items", p.ChecklistItems)
	setIfNotNilString(values, "prepend-checklist-items", p.PrependChecklistItems)
	setIfNotNilString(values, "append-checklist-items", p.AppendChecklistItems)
	setIfNotNilString(values, "list", p.List)
	setIfNotNilString(values, "list-id", p.ListID)
	setIfNotNilString(values, "heading", p.Heading)
	setIfNotNilString(values, "heading-id", p.HeadingID)
	setIfNotNilString(values, "creation-date", p.CreationDate)
	setIfNotNilString(values, "completion-date", p.CompletionDate)
	setIfNotNilBool(values, "completed", p.Completed)
	setIfNotNilBool(values, "canceled", p.Canceled)
	setIfNotNilBool(values, "duplicate", p.Duplicate)
	setIfNotNilBool(values, "reveal", p.Reveal)
	return values
}

// UpdateToDo dispatches things:///update.
func (c *Client) UpdateToDo(params UpdateToDoParams) (*model.URLCommandResult, error) {
	if strings.TrimSpace(c.token) == "" {
		return nil, &APIError{StatusCode: 401, Message: "missing auth token: set --token or THINGS_API_TOKEN"}
	}
	values := params.encode()
	values.Set("auth-token", c.token)
	return c.runURLCommand("update", values)
}

// UpdateProjectParams maps to Things URL command update-project.
type UpdateProjectParams struct {
	ID             string
	Title          *string
	Notes          *string
	PrependNotes   *string
	AppendNotes    *string
	When           *string
	Deadline       *string
	Tags           *string
	AddTags        *string
	Area           *string
	AreaID         *string
	CreationDate   *string
	CompletionDate *string
	Completed      *bool
	Canceled       *bool
	Duplicate      *bool
	Reveal         *bool
}

func (p UpdateProjectParams) encode() url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "id", p.ID)
	setIfNotNilString(values, "title", p.Title)
	setIfNotNilString(values, "notes", p.Notes)
	setIfNotNilString(values, "prepend-notes", p.PrependNotes)
	setIfNotNilString(values, "append-notes", p.AppendNotes)
	setIfNotNilString(values, "when", p.When)
	setIfNotNilString(values, "deadline", p.Deadline)
	setIfNotNilString(values, "tags", p.Tags)
	setIfNotNilString(values, "add-tags", p.AddTags)
	setIfNotNilString(values, "area", p.Area)
	setIfNotNilString(values, "area-id", p.AreaID)
	setIfNotNilString(values, "creation-date", p.CreationDate)
	setIfNotNilString(values, "completion-date", p.CompletionDate)
	setIfNotNilBool(values, "completed", p.Completed)
	setIfNotNilBool(values, "canceled", p.Canceled)
	setIfNotNilBool(values, "duplicate", p.Duplicate)
	setIfNotNilBool(values, "reveal", p.Reveal)
	return values
}

// UpdateProject dispatches things:///update-project.
func (c *Client) UpdateProject(params UpdateProjectParams) (*model.URLCommandResult, error) {
	if strings.TrimSpace(c.token) == "" {
		return nil, &APIError{StatusCode: 401, Message: "missing auth token: set --token or THINGS_API_TOKEN"}
	}
	values := params.encode()
	values.Set("auth-token", c.token)
	return c.runURLCommand("update-project", values)
}

// DeleteToDoParams controls AppleScript to-do deletion.
type DeleteToDoParams struct {
	ID   string
	Name string
}

// DeleteToDo deletes a to-do via AppleScript.
func (c *Client) DeleteToDo(params DeleteToDoParams) (*model.URLCommandResult, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" && strings.TrimSpace(params.ID) != "" {
		todo, err := c.GetToDo(params.ID)
		if err != nil {
			return nil, err
		}
		name = todo.Title
	}
	if name == "" {
		return nil, fmt.Errorf("to-do name or --id is required")
	}

	script := fmt.Sprintf("tell application \"Things3\" to delete to do named %s", appleScriptString(name))
	return c.runAppleScript("todo-delete", script)
}

// DeleteProjectParams controls AppleScript project deletion.
type DeleteProjectParams struct {
	ID   string
	Name string
}

// DeleteProject deletes a project via AppleScript.
func (c *Client) DeleteProject(params DeleteProjectParams) (*model.URLCommandResult, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" && strings.TrimSpace(params.ID) != "" {
		project, err := c.GetProject(params.ID)
		if err != nil {
			return nil, err
		}
		name = project.Title
	}
	if name == "" {
		return nil, fmt.Errorf("project name or --id is required")
	}

	script := fmt.Sprintf("tell application \"Things3\" to delete project named %s", appleScriptString(name))
	return c.runAppleScript("project-delete", script)
}

// ShowParams maps to Things URL command show.
type ShowParams struct {
	ID     string
	Query  string
	Filter string
}

func (p ShowParams) encode() url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "id", p.ID)
	setIfNotEmpty(values, "query", p.Query)
	setIfNotEmpty(values, "filter", p.Filter)
	return values
}

// Show dispatches things:///show.
func (c *Client) Show(params ShowParams) (*model.URLCommandResult, error) {
	return c.runURLCommand("show", params.encode())
}

// SearchParams maps to Things URL command search.
type SearchParams struct {
	Query string
}

func (p SearchParams) encode() url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "query", p.Query)
	return values
}

// Search dispatches things:///search.
func (c *Client) Search(params SearchParams) (*model.URLCommandResult, error) {
	return c.runURLCommand("search", params.encode())
}

// Version dispatches things:///version.
func (c *Client) Version() (*model.URLCommandResult, error) {
	return c.runURLCommand("version", url.Values{})
}

// JSONParams maps to Things URL command json.
type JSONParams struct {
	Data   string
	Reveal *bool
}

func (p JSONParams) encode() url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "data", p.Data)
	setIfNotNilBool(values, "reveal", p.Reveal)
	return values
}

// JSON dispatches things:///json.
func (c *Client) JSON(params JSONParams) (*model.URLCommandResult, error) {
	if strings.TrimSpace(params.Data) == "" {
		return nil, fmt.Errorf("json data is required")
	}
	values := params.encode()
	if strings.TrimSpace(c.token) != "" {
		values.Set("auth-token", c.token)
	}
	return c.runURLCommand("json", values)
}

func (c *Client) runURLCommand(command string, values url.Values) (*model.URLCommandResult, error) {
	fullURL := "things:///" + command
	if encoded := values.Encode(); encoded != "" {
		fullURL += "?" + encoded
	}

	output, err := c.runner.Run("open", "-g", fullURL)
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg == "" {
			msg = err.Error()
		}
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return nil, &APIError{StatusCode: 404, Message: "'open' command not found"}
		}
		return nil, &APIError{StatusCode: 1, Message: "dispatch URL command failed: " + msg}
	}

	return &model.URLCommandResult{
		Command:    command,
		URL:        fullURL,
		Dispatched: true,
		Message:    "Command dispatched to Things3.",
	}, nil
}

func (c *Client) runAppleScript(action string, script string) (*model.URLCommandResult, error) {
	output, err := c.runner.Run("osascript", "-e", script)
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg == "" {
			msg = err.Error()
		}
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return nil, &APIError{StatusCode: 404, Message: "'osascript' command not found"}
		}
		return nil, &APIError{StatusCode: 1, Message: "run AppleScript failed: " + msg}
	}

	return &model.URLCommandResult{
		Command:    action,
		URL:        "applescript://local",
		Dispatched: true,
		Message:    "AppleScript command executed.",
	}, nil
}

func (c *Client) openDB() (*sql.DB, string, error) {
	dbPath, err := c.resolveDBPath()
	if err != nil {
		return nil, "", err
	}

	dsn := fmt.Sprintf("file:%s?mode=ro", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, "", fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, "", fmt.Errorf("ping sqlite: %w", err)
	}

	return db, dbPath, nil
}

func (c *Client) resolveDBPath() (string, error) {
	if c.dbPath != "" {
		if fileExists(c.dbPath) {
			return c.dbPath, nil
		}
		return "", &APIError{StatusCode: 404, Message: "database file not found: " + c.dbPath}
	}

	if envPath := strings.TrimSpace(os.Getenv("THINGSDB")); envPath != "" {
		resolved := expandHomePath(envPath)
		if fileExists(resolved) {
			return resolved, nil
		}
		return "", &APIError{StatusCode: 404, Message: "database file not found: " + resolved}
	}

	newPattern := expandHomePath(defaultDBPatternNew)
	if matches, err := filepath.Glob(newPattern); err == nil {
		for _, match := range matches {
			if fileExists(match) {
				return match, nil
			}
		}
	}

	legacyPath := expandHomePath(defaultDBPatternLegacy)
	if fileExists(legacyPath) {
		return legacyPath, nil
	}

	return "", &APIError{StatusCode: 404, Message: "Things3 database not found; set THINGSDB or ensure Things is installed"}
}

func mapStatus(status string) (int, error) {
	s := strings.ToLower(strings.TrimSpace(status))
	switch s {
	case "", "all":
		return -1, nil
	case "incomplete", "open":
		return 0, nil
	case "canceled":
		return 2, nil
	case "completed":
		return 3, nil
	default:
		return -1, fmt.Errorf("invalid status %q, valid values: incomplete|completed|canceled", status)
	}
}

func statusLabel(value int) string {
	switch value {
	case 0:
		return "incomplete"
	case 2:
		return "canceled"
	case 3:
		return "completed"
	default:
		return "unknown"
	}
}

func startLabel(value int) string {
	switch value {
	case 0:
		return "inbox"
	case 1:
		return "anytime"
	case 2:
		return "someday"
	default:
		return ""
	}
}

func nullStringValue(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return v.String
}

func splitCommaValues(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	if len(values) == 0 {
		return nil
	}
	return values
}

func uniqueStrings(values []string) []string {
	if len(values) <= 1 {
		return values
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func appleScriptString(value string) string {
	escaped := strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(value)
	return fmt.Sprintf("\"%s\"", escaped)
}

func setIfNotEmpty(values url.Values, key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	values.Set(key, value)
}

func setIfNotNilString(values url.Values, key string, value *string) {
	if value == nil {
		return
	}
	values.Set(key, *value)
}

func setIfNotNilBool(values url.Values, key string, value *bool) {
	if value == nil {
		return
	}
	values.Set(key, strconv.FormatBool(*value))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func expandHomePath(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(homeDir, strings.TrimPrefix(path, "~/"))
}

func normalizeDatabaseVersion(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if !strings.Contains(trimmed, "<integer>") {
		return trimmed
	}
	start := strings.Index(trimmed, "<integer>")
	end := strings.Index(trimmed, "</integer>")
	if start < 0 || end <= start {
		return trimmed
	}
	value := strings.TrimSpace(trimmed[start+len("<integer>") : end])
	if value == "" {
		return trimmed
	}
	return value
}
