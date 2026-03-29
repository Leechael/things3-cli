package model

// Status represents local Things3 integration status.
type Status struct {
	OK                        bool   `json:"ok"`
	DatabasePath              string `json:"database_path"`
	DatabaseVersion           string `json:"database_version"`
	URLSchemeCommandAvailable bool   `json:"url_scheme_command_available"`
	TokenConfigured           bool   `json:"token_configured"`
}

// ToDo represents a Things to-do read from SQLite.
type ToDo struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Notes       string   `json:"notes,omitempty"`
	Status      string   `json:"status"`
	Start       string   `json:"start,omitempty"`
	StartDate   string   `json:"start_date,omitempty"`
	Deadline    string   `json:"deadline,omitempty"`
	ProjectID   string   `json:"project_id,omitempty"`
	Project     string   `json:"project,omitempty"`
	AreaID      string   `json:"area_id,omitempty"`
	Area        string   `json:"area,omitempty"`
	HeadingID   string   `json:"heading_id,omitempty"`
	Heading     string   `json:"heading,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	IsRecurring bool     `json:"is_recurring"`
}

// StartDateOrStart returns start_date when available, otherwise start bucket label.
func (t ToDo) StartDateOrStart() string {
	if t.StartDate != "" {
		return t.StartDate
	}
	return t.Start
}

// Project represents a Things project read from SQLite.
type Project struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Notes         string   `json:"notes,omitempty"`
	Status        string   `json:"status"`
	StartDate     string   `json:"start_date,omitempty"`
	Deadline      string   `json:"deadline,omitempty"`
	AreaID        string   `json:"area_id,omitempty"`
	Area          string   `json:"area,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	TaskCount     int      `json:"task_count"`
	OpenTaskCount int      `json:"open_task_count"`
	DoneTaskCount int      `json:"done_task_count"`
	IsRecurring   bool     `json:"is_recurring"`
}

// Area represents a Things area read from SQLite.
type Area struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
}

// Tag represents a Things tag read from SQLite.
type Tag struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Shortcut string `json:"shortcut,omitempty"`
	ParentID string `json:"parent_id,omitempty"`
	Parent   string `json:"parent,omitempty"`
}

// CreateToDoRequest follows Things add command fields.
type CreateToDoRequest struct {
	Title          string `json:"title,omitempty"`
	Titles         string `json:"titles,omitempty"`
	Notes          string `json:"notes,omitempty"`
	When           string `json:"when,omitempty"`
	Deadline       string `json:"deadline,omitempty"`
	Tags           string `json:"tags,omitempty"`
	ChecklistItems string `json:"checklist_items,omitempty"`
	List           string `json:"list,omitempty"`
	ListID         string `json:"list_id,omitempty"`
	Heading        string `json:"heading,omitempty"`
	HeadingID      string `json:"heading_id,omitempty"`
	CreationDate   string `json:"creation_date,omitempty"`
	CompletionDate string `json:"completion_date,omitempty"`
	UseClipboard   string `json:"use_clipboard,omitempty"`
	Completed      *bool  `json:"completed,omitempty"`
	Canceled       *bool  `json:"canceled,omitempty"`
	ShowQuickEntry *bool  `json:"show_quick_entry,omitempty"`
	Reveal         *bool  `json:"reveal,omitempty"`
}

// UpdateToDoRequest uses pointers for partial update semantics.
type UpdateToDoRequest struct {
	ID                    string  `json:"id"`
	Title                 *string `json:"title,omitempty"`
	Notes                 *string `json:"notes,omitempty"`
	PrependNotes          *string `json:"prepend_notes,omitempty"`
	AppendNotes           *string `json:"append_notes,omitempty"`
	When                  *string `json:"when,omitempty"`
	Deadline              *string `json:"deadline,omitempty"`
	Tags                  *string `json:"tags,omitempty"`
	AddTags               *string `json:"add_tags,omitempty"`
	ChecklistItems        *string `json:"checklist_items,omitempty"`
	PrependChecklistItems *string `json:"prepend_checklist_items,omitempty"`
	AppendChecklistItems  *string `json:"append_checklist_items,omitempty"`
	List                  *string `json:"list,omitempty"`
	ListID                *string `json:"list_id,omitempty"`
	Heading               *string `json:"heading,omitempty"`
	HeadingID             *string `json:"heading_id,omitempty"`
	CreationDate          *string `json:"creation_date,omitempty"`
	CompletionDate        *string `json:"completion_date,omitempty"`
	Completed             *bool   `json:"completed,omitempty"`
	Canceled              *bool   `json:"canceled,omitempty"`
	Duplicate             *bool   `json:"duplicate,omitempty"`
	Reveal                *bool   `json:"reveal,omitempty"`
}

// CreateProjectRequest follows Things add-project command fields.
type CreateProjectRequest struct {
	Title          string `json:"title,omitempty"`
	Notes          string `json:"notes,omitempty"`
	When           string `json:"when,omitempty"`
	Deadline       string `json:"deadline,omitempty"`
	Tags           string `json:"tags,omitempty"`
	Area           string `json:"area,omitempty"`
	AreaID         string `json:"area_id,omitempty"`
	ToDos          string `json:"to_dos,omitempty"`
	CreationDate   string `json:"creation_date,omitempty"`
	CompletionDate string `json:"completion_date,omitempty"`
	Completed      *bool  `json:"completed,omitempty"`
	Canceled       *bool  `json:"canceled,omitempty"`
	Reveal         *bool  `json:"reveal,omitempty"`
}

// UpdateProjectRequest uses pointers for partial update semantics.
type UpdateProjectRequest struct {
	ID             string  `json:"id"`
	Title          *string `json:"title,omitempty"`
	Notes          *string `json:"notes,omitempty"`
	PrependNotes   *string `json:"prepend_notes,omitempty"`
	AppendNotes    *string `json:"append_notes,omitempty"`
	When           *string `json:"when,omitempty"`
	Deadline       *string `json:"deadline,omitempty"`
	Tags           *string `json:"tags,omitempty"`
	AddTags        *string `json:"add_tags,omitempty"`
	Area           *string `json:"area,omitempty"`
	AreaID         *string `json:"area_id,omitempty"`
	CreationDate   *string `json:"creation_date,omitempty"`
	CompletionDate *string `json:"completion_date,omitempty"`
	Completed      *bool   `json:"completed,omitempty"`
	Canceled       *bool   `json:"canceled,omitempty"`
	Duplicate      *bool   `json:"duplicate,omitempty"`
	Reveal         *bool   `json:"reveal,omitempty"`
}
