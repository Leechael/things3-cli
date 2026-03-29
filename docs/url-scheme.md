# Things3 URL Scheme

Things3 provides a URL Scheme as the primary interface for external programs to interact with Things. This is the **only officially supported write method**.

Current URL Scheme version: **2**

## Prerequisites

1. Things 3 installed and running on macOS
2. Things → Settings → General → Enable Things URLs
3. Modification operations (update) require an auth-token: Things → Settings → General → Things URLs → Manage

## URL Format

```
things:///commandName?parameter1=value1&parameter2=value2
```

All commands support the x-callback-url protocol:
- `x-success`: success callback URL
- `x-error`: error callback URL
- `x-cancel`: cancel callback URL

## Execution Methods

```bash
# Open directly on macOS
open "things:///add?title=Buy%20milk"

# Run in background (without switching to the Things window)
open -g "things:///add?title=Buy%20milk"

# Via osascript
osascript -e 'open location "things:///add?title=Buy%20milk"'
```

## Data Types

| Type | Format | Example |
|------|--------|---------|
| string | URL encoded | `Buy%20milk` |
| date string | `yyyy-mm-dd` or natural language | `2026-03-15`, `today`, `tomorrow`, `in 3 days`, `next tuesday` |
| time string | local timezone | `9:30PM`, `21:30` |
| date time string | `date@time` | `2026-02-25@14:00`, `evening@6pm` |
| ISO8601 | RFC standard | `2026-03-10T14:30:00Z` |
| boolean | | `true`, `false` |
| JSON string | valid JSON | `[...]` |

## General Behavior Rules

1. **Empty parameter clears value**: Including `=` without a value (e.g., `deadline=`) clears that field. This is a general rule that applies to all clearable fields (deadline, when, tags, etc.), not just deadline.
2. **Default to Inbox**: When the `add` command does not specify `when` or `list-id`, the to-do goes to the Inbox by default.
3. **Tags must already exist**: The `tags` / `add-tags` parameters can only reference existing tags. Tag names that do not exist will be **silently ignored** and will not be created automatically.
4. **ID takes priority**: `list-id` takes priority over `list`, `heading-id` takes priority over `heading`, `area-id` takes priority over `area`.
5. **canceled takes priority**: When both completed and canceled are set, canceled takes effect.
6. **Bidirectional reopen**: Setting `completed=false` on a canceled to-do marks it as incomplete; conversely, setting `canceled=false` on a completed to-do also marks it as incomplete.
7. **Rate limit**: Maximum 250 add operations within 10 seconds.
8. **Unicode support**: Titles, notes, and other fields support Unicode characters; URL encoding is required.

## Command Reference

---

### `add` — Create a To-Do

```
things:///add?title=...&notes=...&when=...
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `title` | string | No | Title (max 4,000 characters). Ignored if `titles` is specified |
| `titles` | string | No | Multiple titles separated by `%0a`. Takes priority over `title` and `show-quick-entry` |
| `notes` | string | No | Notes (max 10,000 characters) |
| `when` | date/time | No | `today`, `tomorrow`, `evening`, `anytime`, `someday`, `yyyy-mm-dd`, `yyyy-mm-dd@hh:mm` |
| `deadline` | date | No | Deadline `yyyy-mm-dd` |
| `tags` | string | No | Comma-separated tag names; only existing tags can be used |
| `checklist-items` | string | No | Separated by `%0a`, maximum 100 items |
| `list` | string | No | Target project/area name |
| `list-id` | string | No | Target project/area UUID (takes priority over `list`) |
| `heading` | string | No | heading name within the project |
| `heading-id` | string | No | heading UUID (takes priority over `heading`) |
| `completed` | boolean | No | Mark as completed |
| `canceled` | boolean | No | Mark as canceled (takes priority over completed) |
| `show-quick-entry` | boolean | No | Show quick entry dialog instead of creating directly |
| `reveal` | boolean | No | Navigate to the newly created to-do |
| `use-clipboard` | string | No | Replace content from clipboard (see below) |
| `creation-date` | ISO8601 | No | Set creation date (future dates are ignored) |
| `completion-date` | ISO8601 | No | Set completion date (requires completed/canceled to be true) |

**`use-clipboard` behavior details**:
- `replace-title`: Clipboard content replaces the title; if the clipboard has multiple lines, the first line becomes the title, **and the remaining lines overflow into notes, replacing notes**
- `replace-notes`: Clipboard content replaces the notes
- `replace-checklist-items`: Clipboard content replaces the checklist; **each line creates one checklist item**

**Returns**: `x-things-id` — comma-separated UUIDs of created to-dos

**Examples**:
```
things:///add?title=Buy%20milk&notes=Low%20fat&when=evening&tags=Errand
things:///add?titles=Milk%0aBeer%0aCheese&list=Shopping
things:///add?title=Call%20doctor&when=next%20monday&list-id=TryhwrjdiHEXfjgNtw81yt
things:///add?title=Collect%20dry%20cleaning&when=evening@6pm
```

---

### `add-project` — Create a Project

```
things:///add-project?title=...&to-dos=...
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `title` | string | No | Project title |
| `notes` | string | No | Notes |
| `when` | date/time | No | Start date |
| `deadline` | date | No | Deadline |
| `tags` | string | No | Tags |
| `area` | string | No | Area name |
| `area-id` | string | No | Area UUID (takes priority) |
| `to-dos` | string | No | Sub-task titles separated by `%0a` |
| `completed` | boolean | No | Mark as completed (also completes all sub-tasks) |
| `canceled` | boolean | No | Mark as canceled |
| `reveal` | boolean | No | Navigate to the new project |
| `creation-date` | ISO8601 | No | Creation date (also applied to sub-tasks) |
| `completion-date` | ISO8601 | No | Completion date |

**Returns**: `x-things-id` — project UUID

**Examples**:
```
things:///add-project?title=Plan%20Birthday%20Party&area=Family
things:///add-project?title=Submit%20Tax&deadline=December%2031&area-id=Lg8UqVPXo2SbJNiBpDBBQ
```

---

### `update` — Modify a To-Do

```
things:///update?auth-token=...&id=...&title=...
```

**Required**: `id` + `auth-token`

| Parameter | Type | Description |
|-----------|------|-------------|
| `auth-token` | string | **Required** authorization token |
| `id` | string | **Required** to-do UUID |
| `title` | string | Replace title |
| `notes` | string | Replace notes |
| `prepend-notes` | string | Prepend to notes (max 10,000 characters) |
| `append-notes` | string | Append to notes (max 10,000 characters) |
| `when` | date/time | Update scheduled date |
| `deadline` | date | Update deadline (empty value clears it) |
| `tags` | string | Replace all tags |
| `add-tags` | string | Add tags (without removing existing ones) |
| `checklist-items` | string | Replace checklist |
| `prepend-checklist-items` | string | Prepend to checklist |
| `append-checklist-items` | string | Append to checklist |
| `list` / `list-id` | string | Move to another project/area |
| `heading` / `heading-id` | string | Move under a heading. Can be used together with `list`/`list-id` (first moves to the project, then places under the heading) |
| `completed` | boolean | Mark as completed |
| `canceled` | boolean | Mark as canceled |
| `duplicate` | boolean | Duplicate then modify (original unchanged) |
| `reveal` | boolean | Navigate to this to-do |
| `creation-date` | ISO8601 | Set creation date (future dates are ignored) |
| `completion-date` | ISO8601 | Set completion date (future dates are ignored; requires completed/canceled to be true) |

**`heading` behavior in update**: If the to-do is not in a project that has that heading, the parameter is ignored. You can specify `list`/`list-id` at the same time to first move to the target project and then set the heading.

**Limitations**:
- Cannot modify when/deadline/completed/completion-date of repeating tasks
- Cannot duplicate repeating tasks

**Returns**: `x-things-id` — UUID of the updated to-do

**Examples**:
```
things:///update?auth-token=TOKEN&id=SyJEz273ceSkabUbciM73A&when=today
things:///update?auth-token=TOKEN&id=SyJEz273ceSkabUbciM73A&append-notes=Details
things:///update?auth-token=TOKEN&id=SyJEz273ceSkabUbciM73A&deadline=
```

---

### `update-project` — Modify a Project

```
things:///update-project?auth-token=...&id=...
```

**Required**: `id` + `auth-token`

| Parameter | Type | Description |
|-----------|------|-------------|
| `auth-token` | string | **Required** authorization token |
| `id` | string | **Required** project UUID |
| `title` | string | Replace title |
| `notes` | string | Replace notes |
| `prepend-notes` | string | Prepend to notes (max 10,000 characters) |
| `append-notes` | string | Append to notes (max 10,000 characters) |
| `when` | date/time | Update scheduled date |
| `deadline` | date | Update deadline (empty value clears it) |
| `tags` | string | Replace all tags |
| `add-tags` | string | Add tags (without removing existing ones) |
| `area` | string | Move to area (by name) |
| `area-id` | string | Move to area (by UUID, takes priority over `area`) |
| `completed` | boolean | Mark as completed |
| `canceled` | boolean | Mark as canceled (takes priority over completed) |
| `duplicate` | boolean | Duplicate then modify (original unchanged) |
| `reveal` | boolean | Navigate to this project |
| `creation-date` | ISO8601 | Set creation date (future dates are ignored) |
| `completion-date` | ISO8601 | Set completion date (future dates are ignored; requires completed/canceled to be true) |

**Differences from `update`**:
- No checklist-related parameters
- No heading / list related parameters (projects cannot be nested inside other projects)
- Has area / area-id parameters

**Prerequisites for completing/canceling**: Setting `completed=true` or `canceled=true` requires all sub to-dos to be completed or canceled, **and all sub headings to be archived**. Otherwise the operation is ignored.

**Limitations**:
- Cannot modify when/deadline/completed/completion-date of repeating projects
- Cannot duplicate repeating projects
- Cannot add to-dos to a project via this command (use the `list-id` parameter of the `add` command instead)
- Cannot manage headings within a project

**Returns**: `x-things-id` — UUID of the updated project

**Examples**:
```
things:///update-project?auth-token=TOKEN&id=Jvj7EW1fLoScPhaw2JomCT&when=tomorrow
things:///update-project?auth-token=TOKEN&id=Jvj7EW1fLoScPhaw2JomCT&add-tags=Important
things:///update-project?auth-token=TOKEN&id=Jvj7EW1fLoScPhaw2JomCT&prepend-notes=SFO%20to%20JFK.
things:///update-project?auth-token=TOKEN&id=Jvj7EW1fLoScPhaw2JomCT&deadline=
```

---

### `show` — Navigate

```
things:///show?id=today
things:///show?query=vacation&filter=errand
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Built-in list ID or any item UUID (takes priority over `query`) |
| `query` | string | Search by name (similar to Quick Find) |
| `filter` | string | Filter by tags (comma-separated) |

**Built-in list IDs**: `inbox`, `today`, `anytime`, `upcoming`, `someday`, `logbook`, `tomorrow`, `deadlines`, `repeating`, `all-projects`, `logged-projects`

**`query` limitations**: `query` can search by area, project, or tag name, but **cannot search for to-dos**. To navigate to a specific to-do, you must use the `id` parameter or the `search` command.

**Types supported by `id`**: Can pass the UUID of a to-do, project, area, or **tag**.

---

### `search` — Search

```
things:///search?query=vacation
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `query` | string | Search keyword (optional; opens search interface if empty) |

---

### `version` — Version Info

```
things:///version
```

Returns: `x-things-scheme-version` (currently 2) + `x-things-client-version` (app build number)

---

### `json` — JSON Batch Operations

This is the most powerful command, supporting the creation/modification of multiple complex objects in a single call.

```
things:///json?data=<URL-encoded JSON array>&auth-token=TOKEN&reveal=true
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `data` | JSON string | JSON array |
| `auth-token` | string | Required when the payload includes update operations |
| `reveal` | boolean | Navigate to the first created item |

**Returns**: `x-things-ids` — top-level UUIDs in JSON array format

#### JSON Object Structure

```json
{
  "type": "to-do" | "project" | "heading" | "checklist-item",
  "operation": "create" | "update",
  "id": "required only for update",
  "attributes": { ... }
}
```

#### To-Do Attributes (create)

```json
{
  "title": "Title",
  "notes": "Notes",
  "when": "today",
  "deadline": "2026-12-31",
  "tags": ["Tag1", "Tag2"],
  "checklist-items": [
    {"type": "checklist-item", "attributes": {"title": "Sub-item", "completed": false}}
  ],
  "list-id": "project/area UUID",
  "list": "project/area name",
  "heading-id": "heading UUID",
  "heading": "heading name",
  "completed": false,
  "canceled": false,
  "creation-date": "2026-03-10T14:30:00Z",
  "completion-date": "2026-03-10T14:30:00Z"
}
```

#### To-Do Additional Attributes (update only)

| Attribute | Type | Description |
|-----------|------|-------------|
| `prepend-notes` | string | Prepend to notes |
| `append-notes` | string | Append to notes |
| `add-tags` | string | Comma-separated tag names (note: also a comma-separated string in JSON, not an array) |
| `prepend-checklist-items` | string | Separated by `%0a`, prepend to checklist |
| `append-checklist-items` | string | Separated by `%0a`, append to checklist |

#### Project Attributes (create)

Supports an `items` array when creating (for nesting to-do and heading objects):

```json
{
  "title": "Project name",
  "notes": "Notes",
  "when": "today",
  "deadline": "2026-12-31",
  "tags": ["Tag1"],
  "area-id": "area UUID",
  "area": "area name",
  "completed": false,
  "canceled": false,
  "creation-date": "...",
  "completion-date": "...",
  "items": [
    {"type": "to-do", "attributes": {"title": "Task 1"}},
    {"type": "heading", "attributes": {"title": "Phase 2"}},
    {"type": "to-do", "attributes": {"title": "Task 2"}}
  ]
}
```

#### Project Additional Attributes (update only)

`prepend-notes`, `append-notes`, `add-tags` (same as to-do update)

#### Heading Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `title` | string | heading title |
| `archived` | boolean | Whether to archive. Default false. **Only takes effect when all to-dos under the heading are completed/canceled**, otherwise ignored |

#### Checklist Item Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `title` | string | Item title |
| `completed` | boolean | Whether completed. Default false |
| `canceled` | boolean | Whether canceled. Default false, takes priority over completed |

#### Full Example

```json
[
  {
    "type": "project",
    "attributes": {
      "title": "Vacation in Rome",
      "notes": "Some time in August.",
      "area": "Family",
      "items": [
        {"type": "to-do", "attributes": {"title": "Ask Sarah for travel guide"}},
        {"type": "heading", "attributes": {"title": "Sights"}},
        {"type": "to-do", "attributes": {"title": "Vatican City"}},
        {"type": "to-do", "attributes": {
          "title": "Research",
          "checklist-items": [
            {"type": "checklist-item", "attributes": {"title": "Hotels", "completed": true}},
            {"type": "checklist-item", "attributes": {"title": "Transport from airport"}}
          ]
        }}
      ]
    }
  },
  {
    "type": "to-do",
    "operation": "update",
    "id": "Di9deEJeUkVZaDEdbnzQZw",
    "attributes": {"deadline": "today"}
  }
]
```

---

### `add-json` (Deprecated)

Use the `json` command instead.

## How to Get IDs

- **To-do ID**: Control-click to-do → Share → Copy Link (Mac); Tap to-do → Toolbar → Share → Copy Link (iOS)
- **Project/List ID**: Control-click the list in the sidebar → Share → Copy Link
- **Tag ID**: Same as above, control-click the tag
- **Via database**: Read from the `uuid` field of the TMTask / TMArea / TMTag tables
- **Auth Token**: `SELECT uriSchemeAuthenticationToken FROM TMSettings WHERE uuid = 'RhAzEf6qDxCD5PmnZVtBZR'`

## Helper Tools

- **ThingsJSONCoder**: Official Swift helper class from Cultured Code for generating JSON command payloads — https://github.com/culturedcode/ThingsJSONCoder
