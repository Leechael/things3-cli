# Areas

An area is an ongoing responsibility domain with no completion state.

## Overview

An area in Things represents a long-lived domain of responsibility — Work, Personal, Health, Finance. Unlike projects, areas have no end state: they persist indefinitely and serve as containers for projects and to-dos.

Area create, update, and delete require macOS and Things to be running.

## Commands

    things3-cli ls-areas [filters]                               List areas
    things3-cli areas create --name <name> [--tags "t1,t2"]      Create an area
    things3-cli areas list|ls [filters]                          List areas
    things3-cli areas get <id>                                   Get an area by UUID
    things3-cli areas update --id <id>|--name <name> [flags]     Update an area
    things3-cli areas delete --id <id>|--name <name>             Delete an area

## Constraints

- Create, update, and delete require macOS and Things to be running.
- --id is preferred over --name for update/delete in scripts to avoid ambiguity with names containing special characters.
- Deleting an area does not delete its contents. Projects and to-dos inside it are moved to the top level. Verify before deleting.

## Examples

    # Create an area
    things3-cli areas create --name "Work"

    # List areas as JSON
    things3-cli ls-areas --json

    # Rename an area
    things3-cli areas update --name "Work" --new-name "Day Job"

    # Delete by UUID
    things3-cli areas delete --id <uuid>

## Related Topics

- things3-cli help project-vs-area
- things3-cli help projects
