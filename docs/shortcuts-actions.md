# Things3 Apple Shortcuts Actions Technical Reference

Complete Shortcuts Actions documentation. Requires Things 3.17+, macOS 14 / iOS 17 / visionOS 26+.

Source: https://culturedcode.com/things/support/articles/9596775/

## Custom Types

### Item Type

Most actions operate on `Item` (single or multiple). An Item can be a to-do, heading, project, or area. Different item types have different available properties; use the `Type` property to determine the type.

| Property | Type | Applies To | Description |
|----------|------|------------|-------------|
| `Type` | enum | All | `To-Do`, `Heading`, `Project`, `Area` |
| `Title` | string | All | Title |
| `Parent` | Item | to-do, heading, project | A to-do's parent can be a project or area; a heading's parent is a project; a project's parent is an area |
| `Parent ID` | string | to-do, heading, project | UUID of the parent |
| `Heading` | Item | to-do | The heading the to-do belongs to |
| `Is Inbox` | boolean | to-do | Whether the item is in the Inbox |
| `Start` | enum | to-do, project | `On Date`, `Anytime`, `Someday` |
| `Start Date` | date | to-do, project | Start date (time portion is always midnight). Only has a value when Start = On Date |
| `Evening` | boolean | to-do, project | Whether scheduled for the evening. Only applies when Start = On Date and Start Date = today |
| `Reminder Date` | datetime | to-do, project | Reminder date and time. The date portion equals Start Date; the time portion is when the reminder fires |
| `Deadline` | date | to-do, project | Deadline date (time portion is always midnight) |
| `Tags` | string[] | to-do, project, area | List of directly applied tag titles |
| `All Matching Tags` | string[] | to-do, project, area | Tags + inherited tags. Useful for filtering in Find Items |
| `Status` | enum | to-do, heading, project | `Open`, `Completed`, `Canceled` |
| `Completion Date` | datetime | to-do, heading, project | Completion date. Only has a value when Status = Completed or Canceled |
| `Is Logged` | boolean | to-do, heading, project | Whether logged to the Logbook |
| `Notes` | string | to-do, project | Notes content |
| `Checklist` | string | to-do | Checklist content, each item separated by a newline |
| `Creation Date` | datetime | to-do, heading, project | Creation date |
| `Modification Date` | datetime | to-do, heading, project | Modification date |
| `ID` | string | All | UUID |

### List Type

Represents any list that can be displayed in Things:

- **Built-in lists (shown in sidebar)**: Inbox, Today, Upcoming, Anytime, Someday, Logbook
- **Built-in lists (hidden, navigable via Quick Find)**: Tomorrow, Deadlines, Repeating, All Projects, Logged Projects
- **User lists**: One per Area, one per Project, one per Tag (hidden)

Currently the `List` type can only be used with the `Open List` action and has no accessible properties.

---

## 13 Actions

### 1. Create To-Do

Creates a new to-do. All parameters are optional. If no Parent or Start is specified, the to-do goes to the Inbox by default.

| Parameter | Type | Description |
|-----------|------|-------------|
| `Title` | string | Title |
| `Parent` | Item | Target project or area |
| `Heading` | Item | Target heading (Parent must be the project that contains this heading) |
| `Start` | enum | `On Date`, `Anytime`, `Someday` |
| `Start Date` | date | Start date (only when Start = On Date). Supports natural language such as "tomorrow" or "in 7 days" |
| `Evening` | boolean | Whether to schedule for the evening (only takes effect when Start Date = today) |
| `Reminder Time` | time | Reminder time, e.g. `8:15pm` or `20:15` (fires on the Start Date) |
| `Deadline` | date | Deadline date. Supports natural language |
| `Tags` | Tag[] | Tags |
| `Status` | enum | `Open` (default), `Completed`, `Canceled` |
| `Notes` | string | Notes |
| `Checklist` | string | Checklist items, separated by newlines. Status syntax: `- [ ]` Open, `- [x]` Complete, `- [~]` Canceled |
| `Show When Run` | boolean | Whether to show the result in the Siri UI |

**Returns**: The newly created Item (to-do)

### 2. Create To-Do with Quick Entry

Same as Create To-Do, but outputs to the Quick Entry dialog for the user to edit before saving.

- iPhone/iPad/Vision: Things must be running in the foreground
- Mac: Only the Quick Entry dialog is shown
- Leave all parameters empty = invoke Quick Entry directly

**Returns**: Nothing (the to-do does not exist until the user saves it manually)

**Limitation**: Specifying a Heading is not supported (Shortcuts bug, FB11859240)

### 3. Create Heading

Creates a new heading inside a project. **Only projects support headings**.

| Parameter | Type | Description |
|-----------|------|-------------|
| `Title` | string | Heading title |
| `Project` | Item | Target project |

**Returns**: The newly created Item (heading)

> This is the **only programmatic way to create a heading independently**. The URL Scheme can only embed headings when creating a project via JSON; AppleScript does not support headings.

### 4. Create Project

Creates a new project. To add child to-dos, use Create To-Do in a subsequent step with this project as the Parent.

| Parameter | Type | Description |
|-----------|------|-------------|
| `Title` | string | Project title |
| `Area` | Item | Target area |
| `Start` | enum | `On Date`, `Anytime`, `Someday` |
| `Start Date` | date | Start date. Supports natural language |
| `Evening` | boolean | Whether to schedule for the evening |
| `Reminder Time` | time | Reminder time |
| `Deadline` | date | Deadline date. Supports natural language |
| `Tags` | Tag[] | Tags |
| `Status` | enum | `Open` (default), `Completed`, `Canceled` |
| `Notes` | string | Notes |

**Returns**: The newly created Item (project)

### 5. Find Items

Queries items in Things by criteria. **Returns at most 500 items**.

#### Filters

| Filter | Description |
|--------|-------------|
| `Type` | Filter by type: to-do, heading, project, area |
| `Title` | Title contains the specified string |
| `Parent` | Is (or is not) inside the specified project or area |
| `Parent ID` | Same as Parent, but specified by UUID (recommended to avoid bugs) |
| `Is Inbox` / `Not Is Inbox` | Whether the item is in the Inbox |
| `Start` | When property: On Date, Anytime, Someday |
| `Start Date` | Start date matches the specified date or range (Shortcuts requires a time value but ignores it) |
| `Evening` / `Not Evening` | Whether in This Evening (Not Evening includes all items not in Today) |
| `Reminder Date` | Reminder datetime matches the specified range |
| `Deadline` | Deadline matches the specified range |
| `Tags` | Has the specified tag (directly applied) |
| `All Matching Tags` | Has the specified tag (including inherited) |
| `Status` | Open, Completed, Canceled |
| `Completion Date` | Completion date matches the specified range (**time portion is significant**) |
| `Is Logged` / `Not Is Logged` | Whether in the Logbook |
| `Notes` | Notes contain the specified string |
| `Creation Date` | Creation date matches (**time portion is significant**) |
| `Modification Date` | Modification date matches (**time portion is significant**) |
| `ID` | Exact UUID match |

#### Sorting

Can sort by: `Title`, `Start Date`, `Reminder Date`, `Deadline`, `Completion Date`, `Creation Date`, `Modification Date`, `Random`

#### Filters That Are Not Possible

- ❌ Find items with no Deadline
- ❌ Find items with no Parent
- ❌ Find items with no Tags
- ❌ Filter passed-in Items by Type/Status/Start (Shortcuts bug, FB11939711)

**Returns**: A list of matching Items

### 6. Get Items

Specifies multiple specific items and returns them.

| Parameter | Type | Description |
|-----------|------|-------------|
| `Type` | enum | `To-Do`, `Heading`, `Project`, `Area` |
| `Items` | Item[] | The items to retrieve |

**Returns**: The specified list of Items

### 7. Get Selected Items

Gets the currently selected items in Things. When multiple windows are open, uses the last active window.

**Returns**: A list of selected Items

**Limitation**: Cannot get projects/areas selected in the sidebar.

### 8. Edit Items

Modifies properties of one or more items. **Cannot be undone**; shows a warning when editing more than 99 items.

#### Operation Modes

| Mode | Applicable Properties | Description |
|------|-----------------------|-------------|
| `Set` | All | Replace with the new value |
| `Append` | Title, Notes, Checklist | Append after the existing value |
| `Prepend` | Title, Notes, Checklist | Prepend before the existing value |
| `Add` | Tags | Add tags (keeping existing ones) |
| `Remove` | Tags | Remove the specified tags |
| `Remove All` | Tags | Remove all tags |

#### Editable Properties

| Property | Description |
|----------|-------------|
| `Title` | Title |
| `Parent` | The project or area the item belongs to |
| `Start` | Scheduled date. Select On Date to set a specific date; supports natural language |
| `Reminder Time` | **Reminder time** (not supported by URL Scheme). The date portion is ignored; the item's Start Date is preserved |
| `Deadline` | Deadline date |
| `Tags` | Tags (supports Set/Add/Remove/Remove All) |
| `Status` | Open, Completed, Canceled |
| `Completion Date` | Completion date |
| `Notes` | Notes |
| `Checklist` | Checklist. Multiple items separated by newlines. Status: `- [ ]` Open, `- [x]` Complete, `- [~]` Canceled |
| `Creation Date` | Creation date |

**Returns**: The list of edited Items

### 9. Duplicate Items

Duplicates one or more items. **Duplicating areas is not supported**.

**Returns**: The list of newly duplicated Items

### 10. Delete Items

Deletes one or more items. **Cannot be undone**; shows a warning when deleting more than 99 items.

| Parameter | Description |
|-----------|-------------|
| `Items` | Only supports to-dos, headings, and projects. Areas are ignored |
| `Delete Immediately` | OFF = move to Trash (Mac only); ON = delete immediately. iPhone/iPad/Vision has no Trash, so this must be set to ON, otherwise an error occurs |

**Note**: Deleting a project also deletes all its contents; deleting a heading also deletes all to-dos under it.

### 11. Open List

Opens the specified list in Things.

| Parameter | Description |
|-----------|-------------|
| `List` | The list to display |
| `Filter by Tags` | Filter by tags |

### 12. Show Items

Shows the specified items in Things. A single item is shown in its parent list; multiple items are shown in a special list.

### 13. Run Things URL

Executes a Things URL Scheme command. Advantage over Safari's Open URLs: can **run in the background** without forcing Things to the foreground.

| Parameter | Description |
|-----------|-------------|
| `Things URL` | e.g. `things:///add?title=Milk`. Variables must be URL-encoded. Some commands require an auth-token |

**Limitation**: Can only be used with URLs that do not require Things to be in the foreground. `things:///show?id=today` will not work; use Shortcuts' Open URLs instead (FB11520039).

---

## Important Notes and Known Bugs

### Date Handling

- `Current Date` includes the current time. Use `Adjust Date` → `Get Start of Day` to get midnight of the current day
- **Do not** use `Format Date` to reformat a date before passing it to a Things action (may cause failures)
- **Do not** use `Ask Each Time` for date input (Shortcuts bug, especially in timezones east of GMT, FB11969030). Use `Ask for Input` instead
- **Avoid** `is in the last X days` in Find — the results are incorrect and will include all items with future dates (FB11841259). Use `is between` instead
- **Avoid** `is in the last` date ranges in Find — a Shortcuts bug since iOS 16.3 (FB12392280). Use `is after` or `is between` instead
- To check if a date is today: use `is Today`, not `is between` (FB11799692)

### Parent Filtering

- The `Parent` parameter has a bug in macOS 14 / iOS 17 where the setting is lost when the shortcut is closed and reopened. Fixed in macOS 15 / iOS 18
- Recommended: use `Get Items` to retrieve the parent, then pass its ID to the `Parent ID` parameter

### Repeating Tasks

- For repeating tasks with a deadline that appear early, the `Deadline` property returns the **date the task appears in Upcoming**, not the final deadline. The underlying deadline offset is not exposed through Shortcuts

### Edit Action Notes

- Setting a completed to-do back to Open, but if its heading has already been archived, the to-do will be marked as incomplete but remain in the Logbook (by design)

### Platform Limitations

- `Ask Each Time` does not work in the Find action on macOS 15
- The Share Sheet cannot receive Things Items, only title text
- `Delete Items` must use `Delete Immediately` on iPhone/iPad/Vision
- `Run Things URL` cannot be used with URLs that require Things to be in the foreground (such as show)

---

## Calling Shortcuts from the CLI

```bash
# Run a shortcut
shortcuts run "My Things Shortcut"

# With input
echo "input data" | shortcuts run "My Things Shortcut"

# Get output
shortcuts run "My Things Shortcut" | cat

# List all shortcuts
shortcuts list
```

Limitation: Shortcuts must be created in advance in the Shortcuts app. However, you can create a general-purpose shortcut that accepts parameters for use from the CLI.

---

## Workflow for Building Complex Projects

1. `Create Project` to create the project
2. `Create Heading` passing the project from step 1 as the Project parameter
3. `Create To-Do` passing the project from step 1 as Parent and the heading from step 2 as Heading

Example shortcut: https://www.icloud.com/shortcuts/7ddbc94c97934316beebac7fa1687e6e
