# Things3 Entity Operations Matrix

What each entity type can and cannot do under different integration methods. This is the core constraint reference for CLI tool design.

## Overview

| Entity | SQLite Read | URL Scheme Create | URL Scheme Update | URL Scheme Delete | AppleScript | Shortcuts (3.17+) |
|--------|:-----------:|:-----------------:|:-----------------:|:-----------------:|:-----------:|:-----------------:|
| To-Do | ✅ | ✅ add | ✅ update | ❌ | All | All (incl. Reminder) |
| Project | ✅ | ✅ add-project | ✅ update-project | ❌ | All | All |
| Area | ✅ | ❌ | ❌ | ❌ | All | Edit/Remove |
| Tag | ✅ | ❌ | ❌ | ❌ | All (incl. hierarchy) | Edit |
| Heading | ✅ | JSON inline create only | ❌ | ❌ | ❌ | Create+Edit+Archive |
| Checklist Item | ✅ | Bulk set | Bulk replace/append | ❌ | ❌ | Per-item operations |
| Contact | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

## To-Do Operations

### Read (SQLite)

```sql
-- All incomplete to-dos
SELECT * FROM TMTask WHERE type = 0 AND status = 0 AND trashed = 0

-- By project
SELECT * FROM TMTask WHERE type = 0 AND project = '<uuid>'

-- By area
SELECT * FROM TMTask WHERE type = 0 AND area = '<uuid>'

-- By tag
SELECT t.* FROM TMTask t
JOIN TMTaskTag tt ON t.uuid = tt.tasks
JOIN TMTag tag ON tt.tags = tag.uuid
WHERE tag.title = 'TagName' AND t.type = 0

-- Search
SELECT * FROM TMTask WHERE type = 0 AND (title LIKE '%keyword%' OR notes LIKE '%keyword%')

-- To-dos with checklist (requires a secondary query on TMChecklistItem)
SELECT * FROM TMChecklistItem WHERE task = '<todo-uuid>'
```

### Create (URL Scheme)

```
things:///add?title=Title&notes=Notes&when=today&deadline=2026-12-31&tags=Tag1,Tag2&list=ProjectName&checklist-items=Item1%0aItem2
```

Supports all attributes: title, notes, when, deadline, tags, checklist-items, list/list-id, heading/heading-id, completed, canceled, creation-date, completion-date

Bulk create: `titles` parameter (separated by `%0a`), or the `json` command

### Update (URL Scheme)

```
things:///update?auth-token=TOKEN&id=UUID&title=NewTitle&append-notes=Extra
```

**Requires auth-token**. Supports:
- Replace: title, notes, when, deadline, tags, checklist-items
- Append: prepend-notes, append-notes, add-tags, prepend-checklist-items, append-checklist-items
- Status: completed, canceled
- Move: list/list-id, heading/heading-id
- Special: duplicate (copy then modify), reveal

**What cannot be done**:
- ❌ Delete a to-do
- ❌ Modify when/deadline/completed/completion-date on a repeating to-do
- ❌ Duplicate a repeating to-do
- ❌ Operate on individual checklist items (only bulk replace or bulk append)

### AppleScript Supplement

```applescript
-- Delete (not possible via URL Scheme)
tell application "Things3"
  delete to do named "Task to remove"
end tell

-- Get selected to-dos (UI interaction)
tell application "Things3"
  set selected to selected to dos
end tell
```

---

## Project Operations

### Read (SQLite)

```sql
-- All active projects
SELECT * FROM TMTask WHERE type = 1 AND status = 0 AND trashed = 0

-- All contents under a project (to-dos + headings)
SELECT * FROM TMTask WHERE project = '<project-uuid>' AND trashed = 0 ORDER BY "index"

-- Projects in a specific area
SELECT * FROM TMTask WHERE type = 1 AND area = '<area-uuid>' AND trashed = 0
```

### Create (URL Scheme)

```
things:///add-project?title=ProjectName&notes=Notes&when=today&area=AreaName&to-dos=Task1%0aTask2
```

The `to-dos` parameter creates sub-tasks. For more complex structures, use the `json` command:

```json
[{
  "type": "project",
  "attributes": {
    "title": "My Project",
    "items": [
      {"type": "to-do", "attributes": {"title": "Task 1"}},
      {"type": "heading", "attributes": {"title": "Phase 2"}},
      {"type": "to-do", "attributes": {"title": "Task 2"}}
    ]
  }
}]
```

### Update (URL Scheme)

```
things:///update-project?auth-token=TOKEN&id=UUID&title=NewTitle&add-tags=Important
```

**Requires auth-token**. Supported attributes are similar to update, but:
- No checklist-related parameters
- No heading-related parameters
- Has area/area-id parameter
- completed/canceled requires all sub-to-dos to be completed/canceled first

**What cannot be done**:
- ❌ Delete a project
- ❌ Add a to-do to an existing project (use the `list`/`list-id` parameter of the `add` command instead)
- ❌ Modify when/deadline/completed on a repeating project
- ❌ Manage headings within a project (cannot modify/delete after creation)

### Adding a To-Do to an Existing Project

This is a common point of confusion — use `add`, not `update-project`:

```
things:///add?title=New%20Task&list-id=<project-uuid>&heading=Phase%201
```

---

## Area Operations

### Read (SQLite)

```sql
-- All areas
SELECT * FROM TMArea

-- Areas with their tags
SELECT a.uuid, a.title, GROUP_CONCAT(tag.title, ', ') as tags
FROM TMArea a
LEFT JOIN TMAreaTag at ON a.uuid = at.areas
LEFT JOIN TMTag tag ON at.tags = tag.uuid
GROUP BY a.uuid

-- Projects under an area
SELECT * FROM TMTask WHERE type = 1 AND area = '<area-uuid>' AND trashed = 0

-- Direct to-dos under an area (not in any project)
SELECT * FROM TMTask WHERE type = 0 AND area = '<area-uuid>' AND project IS NULL AND trashed = 0
```

### Create / Update / Delete

> **⚠️ URL Scheme does not support Area management operations at all.**

There are no `add-area`, `update-area`, or `delete-area` commands.

Areas can be managed through the following methods:

**AppleScript**:
```applescript
tell application "Things3"
  -- Create
  make new area with properties {name:"Health", tag names:"Personal"}

  -- Rename
  set name of area "Health" to "Wellness"

  -- Delete (sub-items go to Trash; the Area itself disappears immediately)
  delete area named "Old Area"
end tell
```

**Apple Shortcuts** (Things 3.17+): The Edit Items action can modify area attributes; the Remove action can remove the area association from a to-do/project.

**Things App manual operation**: Right-click in the sidebar to manage

### Indirect Reference

Although areas cannot be created or modified, existing areas can be referenced when creating a to-do or project:

```
things:///add?title=Task&list=AreaName
things:///add-project?title=Project&area=AreaName
things:///add-project?title=Project&area-id=<area-uuid>
```

---

## Tag Operations

### Read (SQLite)

```sql
-- All tags
SELECT uuid, title, shortcut FROM TMTag

-- Find tasks that use a specific tag
SELECT t.* FROM TMTask t
JOIN TMTaskTag tt ON t.uuid = tt.tasks
JOIN TMTag tag ON tt.tags = tag.uuid
WHERE tag.title = 'Errand'

-- Find areas that use a specific tag
SELECT a.* FROM TMArea a
JOIN TMAreaTag at ON a.uuid = at.areas
JOIN TMTag tag ON at.tags = tag.uuid
WHERE tag.title = 'Personal'
```

### Create / Update / Delete

> **⚠️ URL Scheme does not support Tag management operations at all.**

There are no `add-tag`, `update-tag`, or `delete-tag` commands.

**Key constraint**: The `tags` / `add-tags` parameters in URL Scheme **can only reference existing tags** — they will not automatically create tags that do not exist. If a non-existent tag name is specified, it is silently ignored.

Tags can be managed through the following methods:

**AppleScript** (most complete):
```applescript
tell application "Things3"
  -- Create
  make new tag with properties {name:"Errand"}

  -- Rename
  set name of tag "Errand" to "Shopping"

  -- Set hierarchy (tags support tree-style nesting)
  set parent tag of tag "Home" to tag "Places"

  -- Delete
  delete tag "Old Tag"
end tell
```

**Apple Shortcuts** (Things 3.17+): The Edit Items action supports modifying tags, but does not support creating new tags or setting tag hierarchy.

### Using Tags in URL Scheme

```
# Set tags when creating a to-do (tags must already exist)
things:///add?title=Task&tags=Errand,Important

# Replace all tags when updating a to-do
things:///update?auth-token=TOKEN&id=UUID&tags=NewTag1,NewTag2

# Append tags when updating a to-do (does not remove existing tags)
things:///update?auth-token=TOKEN&id=UUID&add-tags=ExtraTag
```

---

## Heading Operations

### Read (SQLite)

```sql
-- All headings within a project
SELECT uuid, title FROM TMTask WHERE type = 2 AND project = '<project-uuid>' AND trashed = 0
ORDER BY "index"

-- To-dos under a heading
SELECT * FROM TMTask WHERE type = 0 AND heading = '<heading-uuid>' AND trashed = 0
ORDER BY "index"
```

### Create

> **Can only be created together with a project via the JSON command. Headings cannot be created independently.**

```json
[{
  "type": "project",
  "attributes": {
    "title": "My Project",
    "items": [
      {"type": "heading", "attributes": {"title": "Phase 1"}},
      {"type": "to-do", "attributes": {"title": "Task under Phase 1"}},
      {"type": "heading", "attributes": {"title": "Phase 2"}},
      {"type": "to-do", "attributes": {"title": "Task under Phase 2"}}
    ]
  }
}]
```

### Update / Delete

> **⚠️ Neither URL Scheme nor AppleScript supports modifying or deleting a Heading.**

- URL Scheme: no related commands
- AppleScript: does not support heading objects
- Apple Shortcuts (Things 3.17+): **the only programmatic approach**
  - **Create Heading**: add a heading to an existing project independently (no need to rebuild the project)
  - **Edit Items**: modify the heading title
  - **Archive**: set the heading's archived state (prerequisite: all sub-to-dos must be completed or canceled)

### Referencing an Existing Heading

A target heading can be specified when creating or moving a to-do:

```
things:///add?title=Task&list-id=<project-uuid>&heading=Phase%201
things:///update?auth-token=TOKEN&id=UUID&heading-id=<heading-uuid>
```

---

## Checklist Item Operations

### Read (SQLite)

```sql
SELECT uuid, title,
  CASE status WHEN 0 THEN 'incomplete' WHEN 3 THEN 'completed' WHEN 2 THEN 'canceled' END as status,
  stopDate, creationDate, userModificationDate
FROM TMChecklistItem
WHERE task = '<todo-uuid>'
ORDER BY "index"
```

### Create (URL Scheme)

Create along with a to-do:
```
things:///add?title=Task&checklist-items=Item1%0aItem2%0aItem3
```

Create via JSON (allows setting per-item status):
```json
{
  "type": "to-do",
  "attributes": {
    "title": "Task",
    "checklist-items": [
      {"type": "checklist-item", "attributes": {"title": "Item 1", "completed": true}},
      {"type": "checklist-item", "attributes": {"title": "Item 2"}}
    ]
  }
}
```

### Update (URL Scheme)

```
# Replace all checklist items
things:///update?auth-token=TOKEN&id=UUID&checklist-items=New1%0aNew2

# Append to the end
things:///update?auth-token=TOKEN&id=UUID&append-checklist-items=Extra1%0aExtra2

# Prepend to the beginning
things:///update?auth-token=TOKEN&id=UUID&prepend-checklist-items=First1%0aFirst2
```

**Key limitations**:
- ❌ Cannot modify the title of an individual checklist item
- ❌ Cannot mark an individual checklist item as completed (only bulk replace)
- ❌ Cannot delete an individual checklist item
- ❌ Cannot reorder checklist items
- ❌ AppleScript also does not support checklist item operations
- ✅ Apple Shortcuts supports per-item checklist item operations

### Workaround

To operate on an individual checklist item:
1. Read all current items from SQLite
2. Modify the list in your program
3. Replace entirely using the `checklist-items` parameter (completion state information will be lost, because URL Scheme's checklist-items only accepts title strings)

> This is an important limitation: **replacing a checklist via URL Scheme loses the completion state of each item**, because the `checklist-items` parameter is simply a `%0a`-separated string of titles with no state information. Only the JSON command supports setting an initial state at creation time.

---

## CLI Design Implications

Based on the above constraints, the capability boundaries of the CLI tool are:

### Fully Achievable (SQLite + URL Scheme)
- ✅ Read any entity (SQLite)
- ✅ Create a to-do (URL Scheme)
- ✅ Create a project with headings and to-dos (URL Scheme JSON)
- ✅ Update most attributes of a to-do and project (URL Scheme)
- ✅ Search and filter (SQLite)
- ✅ Navigate to any item (URL Scheme show)

### Requires AppleScript Assistance
- ⚠️ Create/rename/delete Area
- ⚠️ Create/rename/delete Tag (including hierarchy management)
- ⚠️ Delete a to-do or project

### Requires Shortcuts Assistance (Things 3.17+)
- ⚠️ Create a Heading independently (add to an existing project)
- ⚠️ Update/archive an existing Heading
- ⚠️ Operate on the title and completion state of an individual Checklist Item
- ⚠️ Set/modify Reminder Time

### Not Achievable Programmatically
- ❌ Modify core attributes (when/deadline/completed) of a repeating task
- ❌ Attach files to a to-do
- ❌ Sorting (Things does not support automatic sorting)
- ❌ Create new tags (AppleScript only; Shortcuts cannot)
