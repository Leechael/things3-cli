# Tags

Tags are cross-cutting context labels that support a shallow parent-child hierarchy.

## Overview

A tag in Things is a context label that can be applied to any to-do or project regardless of area or project membership. Tags support one level of parent-child hierarchy — a tag can have a parent tag, enabling grouping like Work > Deep Work or Personal > Health.

Tag create, update, and delete require macOS and Things to be running.

## Commands

    things3-cli tags create --name <name> [--parent-name <n>|--parent-id <id>]   Create a tag
    things3-cli tags list|ls [filters]                                            List tags
    things3-cli tags get <id>                                                     Get a tag by UUID
    things3-cli tags update --id <id>|--name <name> [--new-name <n>] [flags]     Update a tag
    things3-cli tags delete --id <id>|--name <name>                              Delete a tag

## Constraints

- Create, update, and delete require macOS and Things to be running.
- Tag hierarchy is limited to one level. A child tag cannot have its own children.
- Deleting a parent tag does not delete its child tags. Children become top-level tags.

## Examples

    # Create a top-level tag
    things3-cli tags create --name "work"

    # Create a child tag
    things3-cli tags create --name "deep-work" --parent-name "work"

    # List all tags
    things3-cli tags list

    # Rename a tag
    things3-cli tags update --name "deep-work" --new-name "focused"

    # Apply tags when creating a to-do
    things3-cli add-todo --title "Write report" --tags "work,focused"

## Related Topics

- things3-cli help todos
