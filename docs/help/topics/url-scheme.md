# URL Scheme

The Things URL Scheme is the write backend for to-dos and projects; it requires Things to be running and handles only create and update operations.

## Overview

The Things URL Scheme (`things:///`) is a macOS inter-app communication protocol that opens Things and performs actions. This CLI uses it for all to-do and project create/update operations. The scheme is processed by Things asynchronously — the CLI dispatches the URL and returns immediately; Things processes it in the background.

Because the URL Scheme is processed by the running Things application, write operations require Things to be open. Reads (via SQLite) work independently of Things' running state.

## Constraints

- Cannot delete to-dos or projects. Delete uses AppleScript instead.
- Cannot manage areas, tag definitions, or headings. These require AppleScript or Shortcuts Actions.
- Cannot query data. All reads go through SQLite directly.
- Operations are fire-and-forget: the CLI cannot confirm that a write succeeded. If Things is not running, the URL is silently dropped on some macOS versions.
- Repeating to-do fields (when, deadline) behave differently from regular to-dos. The URL Scheme may update only the current instance or all instances depending on Things' internal logic.
- Authentication token is required for write operations. Set via --token flag or THINGS_API_TOKEN environment variable.

## Examples

    # The status command verifies URL Scheme auth
    things3-cli status

    # These commands use URL Scheme internally
    things3-cli add-todo --title "Example"
    things3-cli update-todo --id <uuid> --notes "updated"
    things3-cli projects create --title "New project"

## Related Topics

- things3-cli help applescript
- things3-cli help todos
- things3-cli help projects
- things3-cli help exit-codes
