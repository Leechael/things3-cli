# Todos

To-dos are the core actionable items in Things.

## Overview

A to-do represents a single task with an optional title, notes, deadline, tags, checklist items, and scheduling (when). To-dos belong to a project, an area, or sit in the Inbox if unassigned.

## Commands

    things3-cli add-todo --title <title> [flags]          Create a to-do
    things3-cli ls-todo [filters]                         List to-dos (full filters)
    things3-cli inbox [filters]                           List Inbox to-dos
    things3-cli today [filters]                           List Today to-dos
    things3-cli upcoming [filters]                        List Upcoming to-dos
    things3-cli anytime [filters]                         List Anytime to-dos
    things3-cli someday [filters]                         List Someday to-dos
    things3-cli get-todo <id>                             Get a to-do by UUID
    things3-cli update-todo --id <id> [flags]             Update a to-do
    things3-cli delete-todo --id <id>|--name <title>      Delete a to-do

## Constraints

- Delete requires macOS and Things to be running.
- Repeating to-dos have limited update support: when, deadline, and completed fields may not behave as expected. Prefer editing repeating to-dos in the Things UI.
- Checklist items cannot be edited individually. Use --checklist-items to replace the full list.
- After a create or update, there is a brief delay before the change appears in list and get results.

## Examples

    # Add a to-do assigned to a project
    things3-cli add-todo --title "Draft proposal" --project "Q2 Planning"

    # List today's to-dos as JSON
    things3-cli today --json

    # Filter by tag and extract titles with jq
    things3-cli ls-todo --tags "work,focused" --json --jq '.[].title'

    # Update scheduling
    things3-cli update-todo --id <uuid> --when tomorrow

    # Delete by UUID
    things3-cli delete-todo --id <uuid>

## Related Topics

- things3-cli help projects
- things3-cli help areas
- things3-cli help tags
