# Feature Specification: Backup Export & Sync Preparation

**Feature Branch**: `005-backup-export-sync-prep`
**Created**: 2026-04-05
**Status**: Draft
**Input**: User description: "Iteration 5: Prepare architecture for future sync -- timestamp tracking, zip export of database and media for manual backups."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Export a Full Backup (Priority: P1)

A collector wants to create a complete backup of their collection data
and media. They click an "Export Backup" button and choose a save
location. The system generates a single archive file containing the
database and all media files (originals and thumbnails). The collector
can store this archive externally for safekeeping or transfer to
another machine.

**Why this priority**: Manual backup is the most immediate user need.
Without network sync (deferred to a future iteration), a local export
is the only way to protect against data loss.

**Independent Test**: Add several items with images. Click Export
Backup. Verify a single archive is created containing the database
file and all media directories. Verify the archive can be extracted
and the files are intact.

**Acceptance Scenarios**:

1. **Given** the collector has items and media in their collection,
   **When** they trigger an export, **Then** a single archive file is
   generated containing the database and all media files.
2. **Given** the export is in progress, **When** the collector waits,
   **Then** progress feedback is shown (not a frozen UI).
3. **Given** the export completes, **When** the collector inspects the
   archive, **Then** it contains the database file and the complete
   media directory structure (originals/ and thumbnails/).
4. **Given** the collector chooses a save location via file dialog,
   **When** the export completes, **Then** the archive is saved at
   the chosen location with a timestamped filename.

---

### User Story 2 - Verify Timestamps on All Modifications (Priority: P2)

A collector modifies items over time (creating, updating, attaching
images). Every modification MUST record an accurate UTC timestamp so
that future sync tools can determine which changes are newest. The
collector can see when an item was last modified in the item list and
detail views.

**Why this priority**: Accurate timestamps are a prerequisite for any
future conflict resolution in sync. They must be correct before the
sync server is built. This story also improves the user experience by
showing meaningful "last modified" information.

**Independent Test**: Create an item, note the timestamp. Wait, then
update the item. Verify the `updated_at` timestamp changed to the
current time. Verify the timestamp is in UTC format.

**Acceptance Scenarios**:

1. **Given** a new item is created, **When** the item is saved,
   **Then** both `created_at` and `updated_at` are set to the current
   UTC time.
2. **Given** an existing item is modified, **When** the item is saved,
   **Then** `updated_at` is set to the current UTC time while
   `created_at` remains unchanged.
3. **Given** items exist in the collection, **When** the collector
   views the item list or detail view, **Then** the last-modified
   timestamp is displayed in a human-readable local format.

---

### Edge Cases

- What happens when the database file is locked during export (e.g.,
  a write in progress)? The system MUST wait for the lock to release
  or use a snapshot mechanism to avoid corrupted exports.
- What happens when disk space is insufficient to create the archive?
  The system MUST report a clear error and clean up any partial files.
- What happens when the media directory is very large (many GB)?
  The export MUST handle large archives without running out of memory
  (streaming compression, not in-memory buffering).
- What happens when the collector exports with no items or media?
  The system MUST still produce a valid archive containing the empty
  database.
- What happens when the collector's clock is significantly wrong?
  Timestamps are UTC-based; the system MUST use UTC regardless of
  local timezone settings.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide an "Export Backup" action
  accessible from the application UI.
- **FR-002**: The export MUST generate a single archive file
  containing the SQLite database file and the complete media directory
  (originals and thumbnails).
- **FR-003**: The archive filename MUST include a timestamp for
  identification (e.g., `omnicollect-backup-2026-04-05T120000.zip`).
- **FR-004**: The collector MUST be able to choose the save location
  via a native file dialog.
- **FR-005**: The export MUST handle large media directories using
  streaming compression (not loading entire archive into memory).
- **FR-006**: The system MUST show progress feedback during export
  and not freeze the UI.
- **FR-007**: All item modifications (create, update) MUST record an
  `updated_at` timestamp in UTC (ISO 8601 format).
- **FR-008**: Item creation MUST set both `created_at` and
  `updated_at` to the current UTC time.
- **FR-009**: Item updates MUST set `updated_at` to the current UTC
  time without modifying `created_at`.
- **FR-010**: The item list and detail views MUST display the
  last-modified timestamp in the collector's local timezone.

### Key Entities

- **Backup Archive**: A compressed archive file containing a snapshot
  of the database and all media. Identified by a timestamp in the
  filename. Self-contained: extracting it produces a complete,
  restorable dataset.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A collector can export a complete backup in under 30
  seconds for collections of up to 1,000 items with 500 images.
- **SC-002**: The exported archive contains 100% of database records
  and media files, verifiable by item count comparison.
- **SC-003**: All item timestamps are in UTC ISO 8601 format, verified
  by querying the database directly.
- **SC-004**: Modifying an item always updates `updated_at` to the
  current UTC time, with zero exceptions across all modification paths.
- **SC-005**: The export process does not freeze the UI -- the
  collector can see progress or a loading indicator throughout.

## Assumptions

- Iterations 1-4 are complete and functional.
- The timestamp tracking (FR-007 through FR-009) is already partially
  implemented in the existing `insertItem` and `updateItem` functions
  from Iteration 1. This iteration verifies and hardens the behavior.
- The export archive format is ZIP (standard, cross-platform).
- Import/restore from a backup archive is out of scope for this
  iteration -- the archive is intended for manual recovery or future
  sync server consumption.
- The deferred sync server (Docker/k3s) is explicitly out of scope
  and will be specified in a future iteration after the local client
  reaches v1.0.
- Module schema files from `~/.omnicollect/modules/` are included in
  the export alongside the database and media, ensuring a complete
  portable backup.
