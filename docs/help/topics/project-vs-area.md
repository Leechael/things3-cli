# Project vs Area

A project has a defined end state you can complete; an area is an ongoing responsibility that never finishes.

## Overview

This is the most common structural decision in Things. The distinction is intentional in Things' data model and affects how commands behave:

**Project** — use when there is a clear outcome you can check off. "Launch new website", "Complete tax return", "Plan team offsite" are projects. A project can be marked completed or canceled. Its to-dos work toward that single outcome.

**Area** — use when you are describing an ongoing domain of responsibility. "Work", "Health", "Finance", "Side Projects" are areas. An area has no completion state. It is a long-lived container for multiple projects and standalone to-dos.

A simple test: can you imagine finishing it? If yes, it is a project. If it just continues indefinitely, it is an area.

## Constraints

- Projects can be completed or canceled via `projects update --completed` or `--canceled`. Areas have no such state.
- A to-do can belong to a project OR an area, not both. Assigning to a project takes precedence.
- Projects can be nested inside areas, but projects cannot be nested inside other projects (Things does not support sub-projects).
- Area create/update/delete require macOS and Things to be running.

## Examples

    # Create an area for an ongoing domain
    things3-cli areas create --name "Health"

    # Create a project inside that area
    things3-cli projects create --title "Run a 5K" --area "Health"

    # Add a to-do to the project (not the area directly)
    things3-cli add-todo --title "Buy running shoes" --project "Run a 5K"

    # Add a standalone to-do to an area (not a specific project)
    things3-cli add-todo --title "Schedule annual checkup" --area "Health"

## Related Topics

- things3-cli help projects
- things3-cli help areas
- things3-cli help todos
