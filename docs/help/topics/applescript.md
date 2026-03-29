# AppleScript

AppleScript is the write backend for delete operations and area/tag management; it requires macOS and Things to be running.

## Overview

AppleScript is a macOS automation language with a direct Things integration dictionary. This CLI uses AppleScript for operations the URL Scheme cannot perform: deleting to-dos and projects, and all area and tag definition management (create, update, delete).

AppleScript operations are synchronous — the CLI waits for Things to complete the operation before returning. This differs from URL Scheme operations which are fire-and-forget.

## Constraints

- Requires macOS. AppleScript operations will fail with a runtime error on Linux or Windows.
- Requires Things to be running and not in a locked/closed state. The CLI will return an error if Things is not available.
- AppleScript requires the user to have granted automation permission to the terminal application. On first run, macOS prompts for this permission. If denied, all AppleScript-backed commands fail silently or with a permission error.
- AppleScript operations in this CLI are not transactional. If a batch operation fails midway, there is no automatic rollback.

## Commands Using AppleScript

    things3-cli delete-todo          Delete a to-do
    things3-cli projects delete      Delete a project
    things3-cli areas create         Create an area
    things3-cli areas update         Update an area
    things3-cli areas delete         Delete an area
    things3-cli tags create          Create a tag
    things3-cli tags update          Update a tag
    things3-cli tags delete          Delete a tag

## Examples

    # Grant automation permission (required once per terminal app)
    # Run any AppleScript-backed command; macOS will prompt for permission.
    things3-cli tags list

    # Verify Things is running before a delete script
    things3-cli status && things3-cli delete-todo --id <uuid>

## Related Topics

- things3-cli help url-scheme
- things3-cli help areas
- things3-cli help tags
