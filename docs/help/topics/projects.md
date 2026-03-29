# Projects

A project groups related to-dos toward a defined outcome; create/update use URL Scheme, delete uses AppleScript.

## Overview

A project in Things has a defined end state — it is something you finish. It can contain to-dos organized under optional headings, and belongs to an area or sits at the top level. When all to-dos are done, the project itself can be marked completed or canceled.

Use a project when the work has a clear completion criteria. Use an area when the responsibility is ongoing with no single endpoint. See `things3-cli help project-vs-area` for the full comparison.

## Commands

    things3-cli ls-projects [filters]                        List incomplete projects (default)
    things3-cli ls-projects --all [filters]                  List all projects
    things3-cli projects create --title <title> [flags]      Create a project
    things3-cli projects list|ls [filters]                   List projects
    things3-cli projects get <id>                            Get a project by UUID
    things3-cli projects update --id <id> [flags]            Update a project
    things3-cli projects delete --id <id>|--name <title>     Delete a project

## Constraints

- URL Scheme cannot delete projects. Delete uses AppleScript and requires macOS + Things running.
- update-project cannot append child to-dos directly. Add to-dos to a project via add-todo --project.
- Heading management (create/reorder headings) is not supported by URL Scheme. Use Shortcuts Actions for heading operations.
- Completing or canceling a project via update-project does not automatically complete its open to-dos. Handle to-do state separately.

## Examples

    # Create a project in an area
    things3-cli projects create --title "Website Relaunch" --area "Work"

    # List incomplete projects as JSON
    things3-cli ls-projects --json

    # List all projects as JSON
    things3-cli ls-projects --all --json

    # Get a specific project
    things3-cli projects get <uuid>

    # Move a project to a different area
    things3-cli projects update --id <uuid> --area "Personal"

    # Mark a project completed
    things3-cli projects update --id <uuid> --completed

## Related Topics

- things3-cli help project-vs-area
- things3-cli help areas
- things3-cli help todos
- things3-cli help url-scheme
- things3-cli help applescript
