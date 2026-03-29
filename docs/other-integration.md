# Things3 Other Integration Methods and Technical Details

In addition to SQLite read-only access and URL Scheme, there are some additional integration channels and technical details worth noting.

## Apple Shortcuts Actions

Things 3.17+ provides a set of **Shortcuts Actions**, which is the most comprehensive programmatic interface, capable of things that neither URL Scheme nor AppleScript can do.

**System requirements**: Things 3.17+, macOS 14 / iOS 17 / visionOS 26+

### Actions List

| Action | Capability | Can other methods substitute? |
|--------|------------|-------------------------------|
| **Find Items** | Query by conditions (type, title, parent, status, tags, dates), returns up to 500 items | SQLite can substitute |
| **Get Selected Items** | Get currently selected to-do | AppleScript can substitute |
| **Edit Items** | Modify attributes, supports Set/Append/Prepend/Add/Remove operations | URL Scheme partial substitute |
| **Delete Items** | **Delete** to-do/project | URL Scheme ❌, AppleScript can substitute |
| **Duplicate Items** | **Standalone duplication** | URL Scheme only has duplicate via update |
| **Create Heading** | **Standalone heading creation** (no need to create together with a project) | URL Scheme ❌, AppleScript ❌ |
| **Run Things URL** | Execute URL Scheme commands in the background (without switching to the Things window) | `open -g` can substitute |
| **Open List** | Navigate to a list | URL Scheme show can substitute |
| **Show Items** | Display specific items | URL Scheme show can substitute |

### Shortcuts-Exclusive Capabilities

These operations **can only be done with Shortcuts**, no other method supports them:

1. **Modify Reminder Time**: URL Scheme cannot set/modify reminder time, Shortcuts can
2. **Standalone Heading Creation**: Add a heading to an existing project without rebuilding the entire project
3. **Edit Heading**: Modify the title and archive state of an existing heading
4. **Edit Checklist Item**: Modify the title and completion state of an individual checklist item
5. **Remove operations**: Remove tags, remove area, etc. (URL Scheme can only replace, not remove individual items)

### Using Shortcuts from the CLI

Shortcuts can be invoked from the command line:

```bash
# Run a shortcut
shortcuts run "My Things Shortcut"

# Run with input
echo "input data" | shortcuts run "My Things Shortcut"

# List all shortcuts
shortcuts list
```

> **Limitation**: Shortcuts must be created in Shortcuts.app in advance and cannot be dynamically generated. However, you can create a generic shortcut that accepts parameters.

---

## Other Input Channels

### Mail to Things

Supports email-to-to-do via Things Cloud: send or forward an email to a dedicated `@things.email` address.

**Technical limitations**:
- Email subject → to-do title (max 4,000 characters)
- Email body → to-do notes (plain text, Markdown supported)
- Attachments not supported (images, PDFs, etc. are ignored)
- Metadata not supported (cannot set tags, deadline, project, etc. via email)
- Rate limit: **max 100 emails per 24 hours**
- Requires Things Cloud to be enabled
- Can integrate with Zapier, IFTTT, TaskClone for workflow automation

### Share Sheet

System-level sharing, create a to-do directly from Safari, Mail, and other apps. Only plain text is accepted. Sharing a file from Finder creates a file path link, not an attachment.

### Drag & Drop

Drag from Mail, Safari, Finder, etc. into Things. Dragging a file creates a file link. Only plain text is supported on iOS.

### Quick Entry (Mac only)

`Ctrl+Space` (default shortcut) brings up a quick entry panel from any app. Enabling Autofill automatically captures the link of the webpage, email, or file currently being viewed.

### Import from Other Apps

Things can import from the following apps:
- Apple Reminders
- Microsoft To Do

Items from these apps appear in Inbox, and the user selects which ones to import.

### Copy & Paste

Pasting multi-line text:
- Paste into a list → creates one to-do per line
- Paste into the Title field → first line becomes the title, the rest goes into notes
- Paste into the Checklist field → creates one checklist item per line
- Maximum 100 lines

### Live Text (iOS)

In the Notes field of a to-do, scan text input via the camera. Title and Checklist fields are not supported.

---

## Key Technical Limitations

### No Public API

> "Things does not provide a public API at this time."

The official documentation explicitly states that no REST API or any remote API is provided. Available integration methods are limited to:
- URL Scheme (local Mac/iOS)
- AppleScript (local Mac)
- Apple Shortcuts (local Mac/iOS/visionOS)
- Mail to Things (via Things Cloud)

### No File Attachment Support

Things does not support attaching files, images, PDFs, etc. to to-dos. Files dragged in from Finder or shared from other apps only create file path links. For cross-device file access, files need to be stored in cloud storage (Dropbox, etc.) with the URL saved in notes.

### No Alphabetical Sorting

Things has no automatic sorting functionality, only manual drag-and-drop ordering is supported.

### Delete Operations

- Mac: Deleted items go to Trash and can be recovered
- iOS: **Deletion is permanent**, can only be undone immediately by shaking the device or pressing Cmd+Z
- Items cannot be deleted via URL Scheme
- Items can be deleted via AppleScript
- Items can be deleted via Shortcuts

### Special Handling of Recurring Tasks

- URL Scheme update cannot modify when/deadline/completed/completion-date of recurring tasks
- Recurring tasks cannot be duplicated via URL Scheme
- SQLite queries typically filter out recurring tasks (`rt1_recurrenceRule IS NULL`)
- Recurring tasks are special entities in the database and require additional logic to handle

### Project Visibility Rules

- Projects with a future start date will not appear in the sidebar
- Projects in Someday will not appear in the sidebar
- The sidebar only shows "active" projects — this is intended behavior, not a bug

---

## Data Safety Warning

The official documentation explicitly warns:

> "Some developers have built unsafe integrations with Things by accessing either your database or your Things Cloud data directly... risks data corruption."

Specifically, Amie is called out for using an unsafe integration method that resulted in user data loss. Our approach (read-only SQLite + URL Scheme for writes) is safe.

---

## Integration Methods Comparison Summary

| Method | Read | Write | Delete | Platform | Requires Token |
|--------|------|-------|--------|----------|----------------|
| SQLite (read-only) | Full data | ❌ | ❌ | Mac | No |
| URL Scheme | ❌ | Create/Modify | ❌ | Mac + iOS | Required for update |
| AppleScript | Supported | Full | Supported | Mac only | No |
| Apple Shortcuts | Supported (≤500 items) | Full | Supported | Mac + iOS + visionOS | No |
| Mail to Things | ❌ | Create only | ❌ | All platforms | No |

> **Important**: The table above is an overview from the integration method dimension. For the entity dimension (specific operation capabilities and limitations for Area/Tag/Heading/Checklist), see [operations-matrix.md](operations-matrix.md).

## Our Technology Choices

For the CLI tool:
- **Read**: SQLite read-only access (most complete, fastest, most programmable)
- **Write**: URL Scheme (the officially supported safe write method)
- **AppleScript**: Creating/deleting Areas and Tags must rely on AppleScript (URL Scheme lacks this capability)
- **Shortcuts**: The only programmatic solution for modifying/deleting Headings and individual Checklist item operations (requires pre-created shortcuts). For a complete Actions reference, see [shortcuts-actions.md](shortcuts-actions.md)
- **Capability boundary**: Reminder Time can only be set via Shortcuts

## Reference Documentation Index

| Document | URL | Content |
|----------|-----|---------|
| URL Scheme | https://culturedcode.com/things/support/articles/2803573/ | Complete URL command reference |
| AppleScript Introduction | https://culturedcode.com/things/support/articles/2803572/ | AppleScript overview and examples |
| AppleScript Commands | https://culturedcode.com/things/support/articles/4562654/ | AppleScript technical reference |
| Shortcuts Actions | https://culturedcode.com/things/support/articles/9596775/ | Shortcuts technical reference |
| Shortcuts Introduction | https://culturedcode.com/things/support/articles/2955145/ | Shortcuts overview and example library |
| Shortcuts Automation | https://culturedcode.com/things/support/articles/8085279/ | Location/time triggers |
| Add from Other Apps | https://culturedcode.com/things/support/articles/2803569/ | Share Sheet / Drag & Drop, etc. |
| Mail to Things | https://culturedcode.com/things/support/articles/2908262/ | Create to-dos via email |
| Data Import | https://culturedcode.com/things/support/articles/2803555/ | Migrate data from other apps |
| Mac Shortcuts | https://culturedcode.com/things/support/articles/2785159/ | Complete keyboard shortcuts list |
| Markdown | https://culturedcode.com/things/support/articles/4651820/ | Markdown support in Notes |
| Natural Language Input | https://culturedcode.com/things/support/articles/9780167/ | Natural language parsing for dates |
| FAQ | https://culturedcode.com/things/support/articles/2967034/ | Frequently asked questions |
