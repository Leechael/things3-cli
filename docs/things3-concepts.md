# Things3 Feature Design and Usage Concepts

Things3's feature design follows the GTD (Getting Things Done) methodology. Understanding these concepts helps us build the right CLI tool.

## Core Data Model

```
Area
├── Project
│   ├── Heading (grouping header, not a task)
│   │   ├── To-Do
│   │   │   └── Checklist Item
│   │   └── To-Do
│   └── To-Do
└── To-Do (belongs directly to Area, not inside a project)

Tag (can be attached to To-Do, Project, Area)
```

### To-Do

The most basic unit. Attributes:
- **Title** + **Notes** (notes support Markdown formatting, including links, bold, italic, lists, etc.)
- **When** (scheduled date): determines which list it appears in. Supports natural language input such as "tomorrow", "in 3 days", "next tuesday"
- **Deadline**: independent of when
- **Reminder**: reminder time (triggers on the start date). **Can only be set via Shortcuts; URL Scheme does not support this**
- **Tags**: multiple tags (supports inheritance — a project's tags are automatically applied to its sub-tasks)
- **Checklist**: up to 100 sub-items
- **Status**: incomplete / completed / canceled

### Project

A completable collection of tasks with a defined endpoint. Can be grouped internally using Headings. A project itself also has when/deadline/tags/notes.

Completing or canceling a project requires all its to-dos to be completed or canceled first.

### Area

An ongoing area of responsibility (e.g., "Work", "Home") with no completion state. Used to organize projects and standalone to-dos.

### Heading

A grouping header inside a project; it is not a task itself. In the database `type = 2`. Can have an `archived` state.

### Tag

A tag system with **hierarchical structure** (via `TMTag.parent` self-reference). For example, "Places" can have child tags "Home" and "Office". Keyboard shortcuts can be assigned to tags. Tags must be created before they can be used in URL Scheme (URL Scheme cannot create tags; only AppleScript and manual operations can).

### Checklist Item

A sub-item inside a to-do, with its own completion state. Maximum 100 per to-do.

## Built-in Lists (Smart Lists)

Things' lists are not storage locations — they are **views** that are dynamically computed based on rules.

| List | Content | DB Equivalent |
|------|---------|---------------|
| **Inbox** | Uncategorized new tasks | `start = 0` |
| **Today** | Tasks to do today + overdue tasks | `startDate = today` or `deadline <= today`, dynamically computed |
| **Upcoming** | Tasks with a future startDate | `startDate > today` |
| **Anytime** | Tasks that can be done at any time | `start = 1` |
| **Someday** | Tasks to maybe do later | `start = 2` |
| **Logbook** | Completed/canceled tasks | `status IN (2, 3)` |
| **Trash** | Deleted tasks | `trashed = 1` |

### The Special Nature of Today

Today is not a simple database query — it is dynamically composed of **three parts**:

1. **Normal Today tasks**: incomplete tasks with `start=Anytime` and a `startDate` value
2. **Unacknowledged scheduled tasks** (indicated by a yellow dot): tasks with `start=Someday` and a `startDate` in the past — the user previously scheduled them but they were moved to Someday, and Things prompts the user to confirm
3. **Unacknowledged overdue tasks**: tasks without a `startDate` but with an expired `deadline`, that have not been dismissed by the user (`deadlineSuppressionDate` is empty or has expired)

Ordering uses a separate `todayIndex` field.

### Deadline Suppression

When an overdue task appears in Today, the user can dismiss it. Things sets `deadlineSuppressionDate` to the current day, removing it from Today until the date changes and it is re-evaluated. This concept is critical for correctly implementing the Today list.

### Evening

`when=evening` is a sub-state of Today — the task is in the Today list but marked as "evening".

### Tomorrow

`things:///show?id=tomorrow` can be used to view tasks with a startDate of tomorrow.

## When vs Deadline

This is the most important design distinction in Things:

- **When** (scheduled date): "When do I plan to start working on this?"
  - Determines which list the task appears in (Today/Upcoming/Someday)
  - Can be vague (today/evening/someday)
- **Deadline**: "What is the latest this must be completed by?"
  - Independent of when
  - When the deadline is reached, the task automatically appears in Today

A task can have both a when and a deadline. For example:
- When: next Monday (planning to start next Monday)
- Deadline: next Friday (must be done by Friday)

## Recurring Tasks

Things supports recurring tasks, but with special limitations:
- Identified in the database via `rt1_recurrenceRule`
- URL Scheme cannot modify the when/deadline/completed fields of recurring tasks
- Recurring tasks cannot be duplicated via URL Scheme
- Most queries should filter out the template rows of recurring tasks

## UUID Format

Things uses 22-character base32-encoded UUIDs (e.g., `TryhwrjdiHEXfjgNtw81yt`). These can be obtained by:
- Reading the `uuid` field from the database
- In Things, Control-click → Share → Copy Link (the link contains the UUID)
- The `x-things-id` returned by URL Scheme

## Contact (Delegation)

Things internally supports a contact/delegation feature (`TMContact` table, linked via `TMTask.contact`), but this is not a core feature exposed in the user interface. The field exists in the database but has limited practical use.

## Recommended Workflow Patterns

### Capture → Organize → Execute

1. **Capture**: Quickly drop into Inbox (URL Scheme `add` without specifying a list)
2. **Organize**: Assign from Inbox to a Project/Area, set when/deadline/tags
3. **Execute**: Check the Today list and complete tasks

### Area for Classification, Project for Goals

- Area: "Work", "Personal", "Health" — persist indefinitely
- Project: "Finish the report", "Renovate the kitchen" — have a defined endpoint

### Tag for Context

- By location: "Office", "Home", "On the go"
- By tool: "Computer", "Phone"
- By energy: "High energy", "Low energy"
- By priority: "Important"

### Someday Is Not a Trash Can

Someday is a parking lot for "things I might do later" — it should be reviewed regularly, not forgotten.

## CLI Tool Design Reference

Based on the concepts above, the CLI tool should:

1. **Read commands** corresponding to built-in lists: `inbox`, `today`, `upcoming`, `anytime`, `someday`, `logbook`
2. **Browse commands**: `projects`, `areas`, `tags`, filter by project/area/tag
3. **Search**: full-text search of title + notes
4. **Create**: via URL Scheme add/add-project
5. **Modify**: via URL Scheme update (requires auth-token and UUID)
6. **Navigate**: via URL Scheme show to open Things and jump to a specific item

Key principle: **reads go through the database, writes go through URL Scheme**. More complex operations (Heading management, Reminder setting, individual Checklist Item operations) require Shortcuts. Area and Tag management requires AppleScript.
