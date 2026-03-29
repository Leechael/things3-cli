# Things3 SQLite Database (Read-Only)

Things3 stores its data in a local SQLite database. We perform **read-only access** exclusively; all write operations are performed via URL Scheme.

> **Official Warning**: Cultured Code has explicitly stated that direct database access is an "unsafe integration method" that may cause data corruption. We strictly read only.

## Database Path

```
# v3.15.16+ (current version)
~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite

# v3.15.15 and earlier
~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite
```

- Use the environment variable `THINGSDB` to override the default path
- Connection must use read-only mode: `?mode=ro`
- Current database version: **26** (check via `SELECT value FROM Meta WHERE key = 'databaseVersion'`; things.py requires >=24)

## Full Table Structure

The database has **15 tables** in total, categorized by importance:

### Core Business Tables

| Table | Purpose |
|-------|---------|
| `TMTask` | Core task storage (to-do, project, heading) |
| `TMArea` | Areas |
| `TMTag` | Tag definitions |
| `TMTaskTag` | Task-tag associations (many-to-many) |
| `TMAreaTag` | Area-tag associations (many-to-many) |
| `TMChecklistItem` | Checklist items for to-dos |
| `TMContact` | Contacts / delegation targets |

### Configuration and Metadata Tables

| Table | Purpose |
|-------|---------|
| `TMSettings` | Application settings (includes URL Scheme authentication token) |
| `TMSmartList` | Custom smart lists (saved filter criteria) |
| `Meta` | Metadata (database version, etc.) |

### Sync and Internal Tables

| Table | Purpose |
|-------|---------|
| `TMTombstone` | Tombstone records for deleted items (used for sync) |
| `TMMetaItem` | General metadata items |
| `BSSyncronyMetadata` | Things Cloud sync metadata |
| `ThingsTouch_ExtensionCommandStore_Commands` | Extension command queue |
| `ThingsTouch_ExtensionCommandStore_Meta` | Extension command metadata |

---

## TMTask Table Schema

This is the most central table, storing all to-dos, projects, and headings.

### Basic Fields

| Column | Type | Description |
|--------|------|-------------|
| `uuid` | TEXT PK | Unique identifier (22-character base32 encoded) |
| `type` | INTEGER | 0=to-do, 1=project, 2=heading |
| `title` | TEXT | Task title |
| `notes` | TEXT | Notes (max 10,000 characters) |
| `notesSync` | INTEGER | Notes sync status |
| `status` | INTEGER | 0=incomplete, 2=canceled, 3=completed |
| `start` | INTEGER | 0=Inbox, 1=Anytime, 2=Someday |
| `trashed` | INTEGER | 1=deleted, 0/NULL=normal |
| `leavesTombstone` | INTEGER | Whether to leave a tombstone record on deletion (for sync) |

### Relationship Fields

| Column | Type | Description |
|--------|------|-------------|
| `area` | TEXT (indexed) | UUID referencing TMArea |
| `project` | TEXT (indexed) | UUID of the parent project |
| `heading` | TEXT (indexed) | UUID of the parent heading |
| `contact` | TEXT | UUID referencing TMContact (delegation target) |

### Ordering Fields

| Column | Type | Description |
|--------|------|-------------|
| `index` | INTEGER | Display order in the current list |
| `todayIndex` | INTEGER | Display order in the Today list |
| `todayIndexReferenceDate` | INTEGER | Reference date for todayIndex |

### Timestamp Fields

Things3 uses two different time encoding methods:

#### Unix Timestamps (REAL, UTC)

| Column | Description |
|--------|-------------|
| `creationDate` | Creation time |
| `userModificationDate` | Last modification time |
| `stopDate` | Completion/cancellation time |
| `lastReminderInteractionDate` | Last reminder interaction time |

#### Things Custom Binary Date Format (INTEGER)

| Column | Description |
|--------|-------------|
| `startDate` | Start date |
| `startBucket` | Bucketed value for the start date |
| `deadline` | Due date |
| `deadlineSuppressionDate` | Date the deadline was dismissed by the user (see "Deadline Suppression" below) |
| `t2_deadlineOffset` | Deadline offset |
| `reminderTime` | Reminder time |

### Cached Count Fields

| Column | Type | Description |
|--------|------|-------------|
| `untrashedLeafActionsCount` | INTEGER | Total number of non-trashed leaf tasks (used for projects) |
| `openUntrashedLeafActionsCount` | INTEGER | Number of incomplete non-trashed leaf tasks (used for projects) |
| `checklistItemsCount` | INTEGER | Total number of checklist items |
| `openChecklistItemsCount` | INTEGER | Number of incomplete checklist items |

### Cache and Internal Fields

| Column | Type | Description |
|--------|------|-------------|
| `cachedTags` | BLOB | Tag cache (internal optimization) |
| `experimental` | BLOB | Experimental feature data |

### Recurring Task Fields

| Column | Type | Description |
|--------|------|-------------|
| `rt1_recurrenceRule` | BLOB | Recurrence rule (NULL means non-repeating) |
| `rt1_repeatingTemplate` | TEXT (indexed) | UUID of the repeating task template |
| `rt1_instanceCreationStartDate` | INTEGER | Start date for instance creation |
| `rt1_instanceCreationPaused` | INTEGER | Whether instance creation is paused |
| `rt1_instanceCreationCount` | INTEGER | Number of instances created |
| `rt1_afterCompletionReferenceDate` | INTEGER | Reference date after completion |
| `rt1_nextInstanceStartDate` | INTEGER | Start date of the next instance |
| `repeater` | BLOB | New recurrence rule (being migrated) |
| `repeaterMigrationDate` | REAL | Recurrence rule migration date |

> Most queries filter with `rt1_recurrenceRule IS NULL`; recurring tasks require separate handling.

---

## TMArea Table

| Column | Type | Description |
|--------|------|-------------|
| `uuid` | TEXT PK | Unique identifier |
| `title` | TEXT | Area name |
| `visible` | INTEGER | Whether visible |
| `index` | INTEGER | Display order |
| `cachedTags` | BLOB | Tag cache |
| `experimental` | BLOB | Experimental feature data |

Tags are associated via `TMAreaTag`.

## TMTag Table

| Column | Type | Description |
|--------|------|-------------|
| `uuid` | TEXT PK | Unique identifier |
| `title` | TEXT | Tag name |
| `shortcut` | TEXT | Keyboard shortcut |
| `usedDate` | REAL | Last used time |
| `parent` | TEXT | Parent tag UUID (self-referencing, supports nested hierarchy) |
| `index` | INTEGER | Display order |
| `experimental` | BLOB | Experimental feature data |

> **Tags support hierarchical structure**: The `parent` field implements a tag tree. For example, "Places" can have child tags "Home" and "Office".

## TMChecklistItem Table

| Column | Type | Description |
|--------|------|-------------|
| `uuid` | TEXT PK | Unique identifier |
| `title` | TEXT | Item title |
| `status` | INTEGER | Same as TMTask.status |
| `task` | TEXT (indexed) | UUID of the parent TMTask |
| `index` | INTEGER | Display order |
| `stopDate` | REAL | Completion time (Unix timestamp) |
| `creationDate` | REAL | Creation time |
| `userModificationDate` | REAL | Modification time |
| `leavesTombstone` | INTEGER | Leave tombstone on deletion |
| `experimental` | BLOB | Experimental feature data |

## TMContact Table

| Column | Type | Description |
|--------|------|-------------|
| `uuid` | TEXT PK | Unique identifier |
| `displayName` | TEXT | Display name |
| `firstName` | TEXT | First name |
| `lastName` | TEXT | Last name |
| `emails` | TEXT | Email addresses |
| `appleAddressBookId` | TEXT | Apple Address Book ID |
| `index` | INTEGER | Display order |

> TMTask.contact references this table for the task delegation/associated contact feature.

## TMSmartList Table

| Column | Type | Description |
|--------|------|-------------|
| `uuid` | TEXT PK | Unique identifier |
| `title` | TEXT | Smart list name |
| `index` | INTEGER | Display order |
| `definition` | BLOB | Filter criteria definition |
| `experimental` | BLOB | Experimental feature data |

> User-defined smart lists; the definition field contains the filter rules.

## TMSettings Table

| Column | Type | Description |
|--------|------|-------------|
| `uuid` | TEXT PK | Fixed value `RhAzEf6qDxCD5PmnZVtBZR` |
| `logInterval` | INTEGER | Logbook auto-log interval |
| `manualLogDate` | REAL | Last manual log date |
| `groupTodayByParent` | INTEGER | Whether to group the Today list by parent |
| `uriSchemeAuthenticationToken` | TEXT | URL Scheme authentication token |
| `experimental` | BLOB | Experimental feature data |

```sql
SELECT uriSchemeAuthenticationToken FROM TMSettings WHERE uuid = 'RhAzEf6qDxCD5PmnZVtBZR'
```

## TMTombstone Table

| Column | Type | Description |
|--------|------|-------------|
| `uuid` | TEXT PK | Unique identifier |
| `deletionDate` | REAL | Deletion time |
| `deletedObjectUUID` | TEXT (indexed) | UUID of the deleted object |

> Used to track deleted objects during Things Cloud sync.

## Meta Table

| Column | Type | Description |
|--------|------|-------------|
| `key` | TEXT PK | Key name |
| `value` | TEXT | Value |

Currently known key-value pairs:
- `databaseVersion` = `26`
- `didRemoveOrphanHeadings` = `true`
- `didCreateDefaultTags` = `true`

---

## Things Custom Date Encoding

This is the most distinctive technical detail of Things3. Dates and times do not use standard formats; instead they use bit-shift encoding.

### Date Encoding (startDate, deadline, deadlineSuppressionDate)

```
Bit layout: YYYYYYYYYYYMMMMDDDDD0000000
- Y (bit 16-25): year value, left-shifted 16 bits
- M (bit 12-15): month value, left-shifted 12 bits
- D (bit 7-11):  day value, left-shifted 7 bits
- 7 trailing zero bits
```

**Example**: 2021-03-28

```
year = 2021, month = 3, day = 28
encoded = (2021 << 16) | (3 << 12) | (28 << 7)
        = 132382720 + 12288 + 3584
        = 132464128
binary: 0000 0111 1110 0101 0011 1000 0000 0000
```

### Time Encoding (reminderTime)

```
Bit layout: hhhhhmmmmmm00000000000000000000
- h (bit 26-30): hour value, left-shifted 26 bits
- m (bit 20-25): minute value, left-shifted 20 bits
- 20 trailing zero bits
```

**Example**: 12:34

```
encoded = (12 << 26) | (34 << 20) = 840957952
```

### Conversion in SQL

```sql
-- Things date -> ISO date
CASE WHEN startDate IS NOT NULL THEN
  printf('%04d-%02d-%02d',
    (startDate >> 16) & 2047,    -- year
    (startDate >> 12) & 15,      -- month
    (startDate >> 7) & 31)       -- day
END

-- Things time -> ISO time
CASE WHEN reminderTime IS NOT NULL THEN
  printf('%02d:%02d',
    (reminderTime >> 26) & 31,   -- hour
    (reminderTime >> 20) & 63)   -- minute
END
```

---

## Deadline Suppression

When a to-do's deadline has passed and it appears in the Today list, the user can choose to "dismiss" it (not wanting to deal with it today). At that point, Things sets `deadlineSuppressionDate` to the current date, removing the task from the Today list until Things is reopened or the date changes and it is re-evaluated.

This field is critical for correctly computing the Today list (see below).

---

## Full Logic for Computing the Today List

The Today list is not a single database field; it is the dynamic result of **three parts** combined:

### 1. Normal Today Tasks

```sql
-- Incomplete tasks with start = Anytime and a startDate
SELECT * FROM TMTask
WHERE status = 0 AND trashed = 0
  AND start = 1              -- Anytime
  AND startDate IS NOT NULL  -- has a start date
  AND rt1_recurrenceRule IS NULL
ORDER BY todayIndex
```

### 2. Unacknowledged Scheduled Tasks (Yellow Dot Indicator)

Tasks that were scheduled in the past but moved to Someday — Things shows a yellow dot to prompt the user to confirm:

```sql
-- start = Someday and startDate is in the past
SELECT * FROM TMTask
WHERE status = 0 AND trashed = 0
  AND start = 2                                    -- Someday
  AND startDate IS NOT NULL
  AND startDate < <today_as_things_date>           -- past date
  AND rt1_recurrenceRule IS NULL
ORDER BY todayIndex
```

### 3. Unacknowledged Overdue Tasks

Tasks with a past deadline but no startDate, unless the user has already dismissed them:

```sql
-- no startDate + deadline has passed + not suppressed
SELECT * FROM TMTask
WHERE status = 0 AND trashed = 0
  AND startDate IS NULL
  AND deadline IS NOT NULL
  AND deadline < <today_as_things_date>            -- overdue
  AND (deadlineSuppressionDate IS NULL
       OR deadlineSuppressionDate < <today_as_things_date>)  -- not suppressed
  AND rt1_recurrenceRule IS NULL
```

### Combined Ordering

The results from all three parts are merged and sorted by `todayIndex` and `startDate`.

---

## Context Trashing

When a project or heading is trashed, the `trashed` field on its child to-dos may still be 0. Be aware of this when querying:

```sql
-- Exclude tasks whose parent has been trashed (context_trashed = False)
SELECT t.* FROM TMTask t
LEFT JOIN TMTask p ON t.project = p.uuid
LEFT JOIN TMTask h ON t.heading = h.uuid
WHERE t.trashed = 0
  AND (p.uuid IS NULL OR p.trashed = 0)
  AND (h.uuid IS NULL OR h.trashed = 0)
```

---

## Common Query Patterns

### Get All Incomplete To-Dos

```sql
SELECT uuid, title, notes, status, start
FROM TMTask
WHERE type = 0              -- to-do
  AND status = 0            -- incomplete
  AND trashed = 0
  AND rt1_recurrenceRule IS NULL
ORDER BY "index"
```

### Get Tasks Under a Specific Project

```sql
SELECT t.uuid, t.title, t.status
FROM TMTask t
WHERE t.project = '<project-uuid>'
  AND t.trashed = 0
ORDER BY t."index"
```

### Get Tasks with Tags

```sql
SELECT t.uuid, t.title, GROUP_CONCAT(tag.title, ', ') as tags
FROM TMTask t
LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
LEFT JOIN TMTag tag ON tt.tags = tag.uuid
WHERE t.type = 0 AND t.status = 0 AND t.trashed = 0
GROUP BY t.uuid
```

### Get Tag Hierarchy Tree

```sql
-- Get all tags and their parents
SELECT t.uuid, t.title, t.shortcut, p.title as parent_title
FROM TMTag t
LEFT JOIN TMTag p ON t.parent = p.uuid
ORDER BY t."index"
```

### Get Areas with Their Tags

```sql
SELECT a.uuid, a.title, GROUP_CONCAT(tag.title, ', ') as tags
FROM TMArea a
LEFT JOIN TMAreaTag at ON a.uuid = at.areas
LEFT JOIN TMTag tag ON at.tags = tag.uuid
WHERE a.visible = 1
GROUP BY a.uuid
ORDER BY a."index"
```

### Get Progress Statistics for Projects

```sql
SELECT uuid, title,
  untrashedLeafActionsCount as total_tasks,
  openUntrashedLeafActionsCount as open_tasks,
  (untrashedLeafActionsCount - openUntrashedLeafActionsCount) as done_tasks
FROM TMTask
WHERE type = 1 AND status = 0 AND trashed = 0
```

### Get Checklist Progress for a To-Do

```sql
SELECT uuid, title,
  checklistItemsCount as total_items,
  openChecklistItemsCount as open_items
FROM TMTask
WHERE type = 0 AND checklistItemsCount > 0 AND trashed = 0
```

---

## Status Value Constants

```python
# TMTask.type
TYPE_TO_DO    = 0
TYPE_PROJECT  = 1
TYPE_HEADING  = 2

# TMTask.status
STATUS_INCOMPLETE = 0
STATUS_CANCELED   = 2
STATUS_COMPLETED  = 3

# TMTask.start
START_INBOX    = 0
START_ANYTIME  = 1
START_SOMEDAY  = 2
```

## Notes

1. **Read-only access**: Never write to the database; all modifications are performed via URL Scheme
2. **Date encoding**: Things dates use a custom binary format, not standard Unix timestamps — they must be decoded correctly
3. **NULL handling**: Many fields use NULL to represent "not set"; be mindful of this when querying
4. **Database lock**: Things.app holds a database lock while running; use read-only mode to avoid conflicts
5. **Path wildcard**: The path for v3.15.16+ includes `ThingsData-*` and requires glob matching
6. **Recurring tasks**: Most scenarios should filter out recurring tasks (`rt1_recurrenceRule IS NULL`)
7. **Context trashing**: When a parent is trashed, child tasks may still have `trashed = 0`; use a JOIN to check
8. **Tag hierarchy**: TMTag.parent supports nested tags; consider the hierarchy when querying
9. **Cached counts**: The `*Count` fields are caches maintained by Things and can be used directly for progress display
10. **Sync tables**: Sync-related tables such as TMTombstone and BSSyncronyMetadata are for Things Cloud use only and do not need to be read by us
