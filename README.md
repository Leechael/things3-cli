# things3-cli

`things3-cli` is a Go-based command-line tool for Things 3, built with the following technical boundaries:

- Read: Local SQLite (read-only)
- Write: Things URL Scheme (`add` / `update` / `show` / `search` / `json`, etc.)

Implementation basis: capability matrix and URL semantic constraints in `docs/*.md`.

## Install

### Option A: Download from GitHub Releases

```bash
gh release list -R Leechael/things3--cli
TAG="vX.Y.Z"
./scripts/print-release-download.sh "$TAG"
gh release download "$TAG" -R Leechael/things3--cli --pattern "things3-cli-*.tar.gz"
```

Extract the archive and move `things3-cli` into your `PATH`.

### Option B: Build from source

```bash
git clone git@github.com:Leechael/things3--cli.git
cd things3--cli
make build
```

## Required configuration

```bash
# Required for commands like update / update-project
export THINGS_API_TOKEN="<token>"

# Optional: Override default database path
export THINGSDB="/absolute/path/to/main.sqlite"
```

Check the environment first:

```bash
things3-cli status
things3-cli status --json
```

## Commands

### To-do CRUD (Core Top-Level Commands)

- `add-todo` (create)
- `ls-todo` (read list with full filters)
- `get-todo <id>` (read one)
- `update-todo --id <id> ...` (update)
- `delete-todo --id <id>|--name <title>` (delete via AppleScript)

`add-todo` and `ls-todo` support specifying `project` / `area` directly by name, and support `--tags` with multiple tags (comma-separated; `ls-todo` uses AND matching).
It also provides `inbox` / `today` / `upcoming` / `anytime` / `someday` as common view commands, supporting combined filtering with `project` / `area` / `tags`.

### Project operations

- `ls-projects` (default: incomplete projects only)
- `ls-projects --all` (show all projects across statuses)
- `projects create`
- `projects list` (`ls`) / `projects get <id>`
- `projects update --id <id> ...`
- `projects delete --id <id>` or `projects delete --name <title>` (via AppleScript)

### Area CRUD

- `ls-areas`
- `areas create --name <name> [--tags "Tag1,Tag2"]` (via AppleScript)
- `areas list` (`ls`) / `areas get <id>`
- `areas update --id <id>|--name <name> [--new-name <name>] [--tags "Tag1,Tag2"]` (via AppleScript)
- `areas delete --id <id>|--name <name>` (via AppleScript)

### Tag CRUD

- `tags create --name <name> [--parent-name <name>|--parent-id <id>]` (via AppleScript)
- `tags list` (`ls`) / `tags get <id>` (`tags list` groups by parent by default)
- `tags update --id <id>|--name <name> [--new-name <name>] [--parent-name <name>|--parent-id <id>]` (via AppleScript)
- `tags delete --id <id>|--name <name>` (via AppleScript)

### Other commands

- `show`
- `search`
- `version`
- `json`
- `help todos|projects|areas|tags` (topic documentation and best practices)
- `add` / `update` (native to-do URL Scheme commands, preserved for compatibility)

## Output modes

- `--json`: Machine-parsable JSON
- `--plain`: Stable plain text (tabwriter output, no headers)
- `--jq`: Available only under `--json`, using `itchyny/gojq`

## Usage examples

```bash
# List tasks (core ls-todo command)
things3-cli ls-todo --search "today"
things3-cli ls-todo --status incomplete --project "Home" --tags "Errand,Important" --json

# Common view commands (support combined filtering)
things3-cli inbox --project "Dashboard" --area "Phala Cloud"
things3-cli today --tags "Errand,Important"

# Create/Update/Delete to-dos (top-level commands)
things3-cli add-todo --title "Buy milk" --when today --project "Shopping" --tags "Errand,Important"
things3-cli update-todo --id "todo-uuid" --append-notes "\nextra details" --reveal
things3-cli delete-todo --id "todo-uuid"

# Area CRUD (via AppleScript)
things3-cli areas create --name "Health" --tags "Personal"
things3-cli areas update --name "Health" --new-name "Wellness"
things3-cli areas delete --name "Wellness"

# Project deletion (via AppleScript)
things3-cli projects delete --id "project-uuid"

# Project creation and reading
things3-cli add-project --title "Plan trip" --area "Personal"
things3-cli ls-projects --json --jq '.results[].title'
things3-cli ls-projects --all --json

# Tag CRUD (via AppleScript)
things3-cli tags create --name "Errand"
things3-cli tags update --name "Errand" --new-name "Shopping"
things3-cli tags delete --name "Shopping"

# Show and search
things3-cli show --id today
things3-cli search "vacation"

# JSON batch processing
things3-cli json --data-file ./payload.json

# Topic help
things3-cli help todos
things3-cli help projects
things3-cli help areas
things3-cli help tags
```

## Development

```bash
make tidy
make fmt
make test
make bdd-test
make ci
make build
make cross-build
```

## Notes

- The URL Scheme does not support deleting to-dos/projects; this CLI uses AppleScript to provide delete capabilities.
- Area CRUD is implemented via AppleScript in this CLI (macOS only).
- Creating/editing independent headings and checklist items requires additional Shortcuts support and is not currently within the scope of this CLI.
