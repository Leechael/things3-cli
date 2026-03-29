# Tags

Tags are cross-cutting context labels that support a shallow parent-child hierarchy; tag definition management uses AppleScript.

## Overview

A tag in Things is a context label that can be applied to any to-do or project regardless of area or project membership. Tags support one level of parent-child hierarchy — a tag can have a parent tag, enabling grouping like Work > Deep Work or Personal > Health.

The URL Scheme can reference existing tags on to-dos and projects but cannot create, rename, or delete tag definitions. This CLI implements tag definition management via AppleScript.

## Commands

    things3-cli tags create --name <name> [--parent-name <n>|--parent-id <id>]   Create a tag
    things3-cli tags list|ls [filters]                                            List tags
    things3-cli tags get <id>                                                     Get a tag by UUID
    things3-cli tags update --id <id>|--name <name> [--new-name <n>] [flags]     Update a tag
    things3-cli tags delete --id <id>|--name <name>                              Delete a tag

## Constraints

- Tag definition operations (create, update, delete) use AppleScript and require macOS + Things running.
- URL Scheme can only apply existing tags to to-dos/projects. It cannot create new tag definitions.
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

    # Apply existing tags when creating a to-do (URL Scheme)
    things3-cli add-todo --title "Write report" --tags "work,focused"

## Related Topics

- things3-cli help todos
- things3-cli help applescript
