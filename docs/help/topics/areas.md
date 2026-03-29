# Areas

An area is an ongoing responsibility domain with no completion state; all area write operations use AppleScript.

## Overview

An area in Things represents a long-lived domain of responsibility — Work, Personal, Health, Finance. Unlike projects, areas have no end state: they persist indefinitely and serve as containers for projects and to-dos.

Because the URL Scheme has no area management endpoints, this CLI implements area create, update, and delete entirely through AppleScript. This means these operations require macOS and Things to be running.

## Commands

    things3-cli areas create --name <name> [--tags "t1,t2"]      Create an area
    things3-cli areas list|ls [filters]                          List areas
    things3-cli areas get <id>                                   Get an area by UUID
    things3-cli areas update --id <id>|--name <name> [flags]     Update an area
    things3-cli areas delete --id <id>|--name <name>             Delete an area

## Constraints

- All write operations (create, update, delete) use AppleScript and require macOS + Things running. They will fail on non-macOS systems or when Things is not open.
- --id and --name are interchangeable for update/delete, but --id is preferred in scripts to avoid matching ambiguity when names contain special characters.
- Deleting an area does not delete the projects and to-dos inside it. Things moves them to the top level. Verify before deleting.

## Examples

    # Create an area
    things3-cli areas create --name "Work"

    # List areas as JSON
    things3-cli areas list --json

    # Rename an area
    things3-cli areas update --name "Work" --new-name "Day Job"

    # Delete by UUID
    things3-cli areas delete --id <uuid>

## Related Topics

- things3-cli help project-vs-area
- things3-cli help projects
- things3-cli help applescript
