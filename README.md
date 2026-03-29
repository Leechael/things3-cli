# things3-cli

`things3-cli` 是一个基于 Go 的 Things3 命令行工具，遵循以下技术边界：

- 读取：本地 SQLite（只读）
- 写入：Things URL Scheme（`add` / `update` / `show` / `search` / `json` 等）

实现依据：`docs/*_zh.md` 中的能力矩阵与 URL 语义约束。

## Install

### Option A: Download from GitHub Releases

```bash
gh release list -R Leechael/things3--cli
TAG="vX.Y.Z"
./scripts/print-release-download.sh "$TAG"
gh release download "$TAG" -R Leechael/things3--cli --pattern "things3-cli-*.tar.gz"
```

解压后将 `things3-cli` 放入你的 `PATH`。

### Option B: Build from source

```bash
git clone git@github.com:Leechael/things3--cli.git
cd things3--cli
make build
```

## Required configuration

```bash
# update / update-project 等命令需要
export THINGS_API_TOKEN="<token>"

# 可选：覆盖默认数据库路径
export THINGSDB="/absolute/path/to/main.sqlite"
```

先检查环境：

```bash
things3-cli status
things3-cli status --json
```

## Commands

### To-do CRUD（核心顶层命令）

- `add-todo`（create）
- `ls`（read list）
- `get-todo <id>`（read one）
- `update-todo --id <id> ...`（update）
- `delete-todo --id <id>|--name <title>`（delete，AppleScript）

`add-todo` 与 `ls` 支持直接按名称输入 `project` / `area`，并支持 `--tags` 多标签（逗号分隔；`ls` 为 AND 匹配）。

### Project operations

- `projects create`
- `projects list` (`ls`) / `projects get <id>`
- `projects update --id <id> ...`
- `projects delete --id <id>` 或 `projects delete --name <title>`（AppleScript）

### Area CRUD

- `areas create --name <name> [--tags "Tag1,Tag2"]`（AppleScript）
- `areas list` (`ls`) / `areas get <id>`
- `areas update --id <id>|--name <name> [--new-name <name>] [--tags "Tag1,Tag2"]`（AppleScript）
- `areas delete --id <id>|--name <name>`（AppleScript）

### Tag CRUD

- `tags create --name <name> [--parent-name <name>|--parent-id <id>]`（AppleScript）
- `tags list` (`ls`) / `tags get <id>`（`tags list` 默认按 parent 分组展示）
- `tags update --id <id>|--name <name> [--new-name <name>] [--parent-name <name>|--parent-id <id>]`（AppleScript）
- `tags delete --id <id>|--name <name>`（AppleScript）

### Other commands

- `show`
- `search`
- `version`
- `json`
- `help todos|projects|areas|tags`（主题文档与 best practices）
- `add` / `update`（to-do URL Scheme 原生命令，兼容保留）

## Output modes

- `--json`: 机器可解析 JSON
- `--plain`: 稳定纯文本（tabwriter 输出，无表头）
- `--jq`: 仅在 `--json` 下可用，使用 `itchyny/gojq`

## Usage examples

```bash
# 列出任务（核心 ls 命令）
things3-cli ls --search "today"
things3-cli ls --status incomplete --project "Home" --tags "Errand,Important" --json

# 创建/更新/删除 to-do（顶层命令）
things3-cli add-todo --title "Buy milk" --when today --project "Shopping" --tags "Errand,Important"
things3-cli update-todo --id "todo-uuid" --append-notes "\nextra details" --reveal
things3-cli delete-todo --id "todo-uuid"

# 区域 CRUD（AppleScript）
things3-cli areas create --name "Health" --tags "Personal"
things3-cli areas update --name "Health" --new-name "Wellness"
things3-cli areas delete --name "Wellness"

# 项目删除（AppleScript）
things3-cli projects delete --id "project-uuid"

# 项目创建与读取
things3-cli add-project --title "Plan trip" --area "Personal"
things3-cli projects list --json --jq '.results[].title'

# 标签 CRUD（AppleScript）
things3-cli tags create --name "Errand"
things3-cli tags update --name "Errand" --new-name "Shopping"
things3-cli tags delete --name "Shopping"

# 跳转和搜索
things3-cli show --id today
things3-cli search "vacation"

# JSON 批量
things3-cli json --data-file ./payload.json

# 主题帮助
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

- URL Scheme 不支持删除 to-do/project；本 CLI 使用 AppleScript 提供删除能力。
- Area CRUD 在本 CLI 中通过 AppleScript 实现（仅 macOS 可用）。
- Heading 独立创建/编辑、Checklist 单项编辑仍需 Shortcuts 额外支持，不在本 CLI 当前范围内。
