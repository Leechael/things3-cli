# Things3 AppleScript Integration

AppleScript is the third integration method for Things3, **Mac only**. Compared to URL Scheme, AppleScript can directly read and write Things data without requiring an auth-token.

> Cultured Code recommends prioritizing Apple Shortcuts, as it is cross-platform and more fully featured (e.g., editing heading and checklist).

## How to Run

### Script Editor

The most basic method: open `/Applications/Utilities/Script Editor.app` and run scripts from there.

### Things Script Menu

1. Quit Things
2. Navigate to `~/Library/Containers/com.culturedcode.ThingsMac/Data/Library/Application Support/Cultured Code/Things Scripts/` (create it if it doesn't exist)
3. Place `.scpt` files in that folder
4. Reopen Things — an S-shaped icon will appear in the menu bar

### Global Script Menu

1. Open Script Editor → Settings → General → enable "Show Script menu in menu bar"
2. Place scripts in `~/Library/Scripts/Applications/Things/`

### Export as Standalone App

Script Editor → File → Export → set File Format to "Application"

### Via Apple Shortcuts

Use the "Run AppleScript" action in the Shortcuts app.

## Basic Operations

### Built-in Lists

```applescript
tell application "Things3"
  set myList to list "Inbox"    -- also: Today, Anytime, Upcoming, Someday, Logbook, Trash
  set items to to dos of myList
end tell
```

> Non-English systems must use the localized list names.

### To-Do Operations

```applescript
tell application "Things3"
  -- get all to-dos
  repeat with toDo in to dos
    log name of toDo
  end repeat

  -- get by name
  set callMom to to do named "Call mom"

  -- get to-dos from a specific list
  set inboxItems to to dos of list "Inbox"

  -- get to-dos from a project
  set projectItems to to dos of project "Vacation in Rome"

  -- get to-dos from an area
  set areaItems to to dos of area "Family"
end tell
```

### Create a To-Do

```applescript
tell application "Things3"
  -- basic creation
  set newToDo to make new to do with properties {name:"New to-do", due date:current date}

  -- create in a specific list
  set newToDo to make new to do with properties {name:"New to-do"} at beginning of list "Today"

  -- create in a project
  set newToDo to make new to do with properties {name:"Buy milk"} at beginning of project "Groceries"

  -- create in an area
  set newToDo to make new to do with properties {name:"Work task"} at beginning of area "Work"
end tell
```

### Modify To-Do Properties

```applescript
tell application "Things3"
  set aToDo to to do named "My Task"
  set name of aToDo to "Renamed!"
  set notes of aToDo to "www.apple.com" & linefeed & "Details here."
  set due date of aToDo to (current date) + 7 * days
  set completion date of aToDo to current date
  set tag names of aToDo to "Home, Mac"
end tell
```

Available properties: `name`, `notes`, `due date`, `creation date`, `modification date`, `completion date`, `cancellation date`, `status` (open/completed/canceled), `tag names`

### Delete a To-Do

```applescript
tell application "Things3"
  delete to do named "To-do to remove"
end tell
```

## Project Operations

```applescript
tell application "Things3"
  -- iterate
  repeat with aProject in projects
    log name of aProject
  end repeat

  -- create
  set newProject to make new project with properties {name:"My Project", notes:"Some notes."}

  -- modify
  set name of newProject to "Renamed!"
  set tag names of newProject to "Home, Mac"

  -- delete
  delete project named "Old Project"
end tell
```

## Area Operations

```applescript
tell application "Things3"
  -- create
  set newArea to make new area with properties {name:"Health"}

  -- modify
  set name of newArea to "Renamed!"

  -- delete (note: the area itself does not go to Trash, but its children do)
  delete area named "Area to Delete"
end tell
```

## Tag Operations

```applescript
tell application "Things3"
  -- create
  make new tag with properties {name:"Home"}

  -- get tags of a to-do
  set tagNames to tag names of first to do of list "Inbox"

  -- set tags
  set tag names of aToDo to "Home, Mac"

  -- rename
  set name of tag "Errands" to "Shopping"

  -- hierarchy
  set parent tag of tag "Home" to tag "Places"

  -- delete
  delete tag "Errands"
end tell
```

## Move Operations

```applescript
tell application "Things3"
  -- move to a built-in list
  move to do named "Task" to list "Today"

  -- mark as complete (move to Logbook)
  set status of to do named "Done task" to completed
  set status of to do named "Skip task" to canceled

  -- schedule for the future (use schedule, not move)
  schedule first to do of list "Inbox" for (current date) + 1 * days

  -- move to a project or area
  set project of to do "Buy milk" to project "Groceries"
  set area of to do "Task" to area "Shopping"

  -- detach from parent
  delete project of first to do of project named "Vacation in Rome"
end tell
```

## UI Interaction

```applescript
tell application "Things3"
  -- get selected to-dos
  set selected to selected to dos

  -- show a specific item
  show to do "Book flights"
  show project "Vacation in Rome"
  show area "Personal"

  -- edit (enter edit mode)
  edit to do named "Task"

  -- Quick Entry
  show quick entry panel
  show quick entry panel with properties {name:"Buy flowers", notes:"She loves tulips."}
  show quick entry panel with autofill yes  -- auto-capture current context

  -- empty Trash
  empty trash

  -- manually log completed items
  log completed now
end tell
```

## AppleScript vs URL Scheme Comparison

| Capability | AppleScript | URL Scheme |
|------------|-------------|------------|
| Platform | Mac only | Mac + iOS |
| Read data | Direct access | Not supported |
| Create To-Do | Supported | Supported |
| Modify To-Do | Supported | Requires auth-token |
| Batch operations | Via loops | JSON commands |
| Delete | Supported | Not supported |
| Heading/Checklist | Not supported | Supported |
| UI interaction | Supported (selection, edit mode) | Limited (show/reveal) |
| Auth Token | Not required | Required for update |
| Cross-app automation | Supported | Supported |

## Limitations

- **Mac only**: iOS/iPadOS is completely unsupported
- **No Heading/Checklist support**: AppleScript cannot operate on heading and checklist item
- **Non-English list names**: Must use the system-localized names
- **No official scripting support**: Features beyond the documentation are not guaranteed to work
