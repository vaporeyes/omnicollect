# Feature Specification: Backup Import and Restore

**Feature Branch**: `015-backup-import`  
**Created**: 2026-04-08  
**Status**: Draft  
**Input**: User description: "Import/Restore from backup ZIP -- extract db + media + modules, merge or replace."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Restore Collection from Backup ZIP (Priority: P1)

A user has a backup ZIP created by OmniCollect's export feature and wants to restore it -- either on a fresh installation, a new device, or after accidental data loss. They click "Import Backup" in the sidebar, select the ZIP file, and choose whether to replace their current data or merge the backup's items into their existing collection. The system extracts items, images, and module schemas from the ZIP and populates the database and media storage.

**Why this priority**: Without import, the export/backup feature is a one-way street. This completes the backup cycle and is the core value proposition.

**Independent Test**: Export a backup from an instance with 10 items and 2 modules. Start a fresh instance. Import the backup with "Replace" mode. Verify all 10 items, images, and module schemas are restored.

**Acceptance Scenarios**:

1. **Given** a fresh installation with no data, **When** the user imports a backup ZIP using "Replace" mode, **Then** all items, images, and module schemas from the backup are restored and visible in the application.
2. **Given** an existing collection with data, **When** the user imports a backup ZIP using "Replace" mode, **Then** all existing data is removed and replaced with the backup's content.
3. **Given** an existing collection with data, **When** the user imports a backup ZIP using "Merge" mode, **Then** the backup's items are added to the existing collection. Items with the same ID are updated with the backup's version. New items are inserted. Existing items not in the backup are preserved.
4. **Given** a backup ZIP with images, **When** the import completes, **Then** all original and thumbnail images from the backup are available and display correctly in the application.
5. **Given** a backup ZIP with module schemas, **When** the import completes, **Then** all module schemas from the backup are available in the module selector.
6. **Given** a corrupted or invalid ZIP file, **When** the user tries to import it, **Then** a clear error message is displayed and no data is modified.

---

### User Story 2 - Import Progress and Confirmation (Priority: P2)

A user importing a large backup (hundreds of items, many images) wants to see progress and know when it's done. Before the import starts, a summary shows what will be imported (item count, image count, module count). During import, a progress indicator shows the current status. After completion, a summary toast confirms what was imported.

**Why this priority**: Without progress feedback, the user doesn't know if a large import is working or frozen. The pre-import summary lets them verify they selected the correct file before committing.

**Independent Test**: Select a backup ZIP with 50+ items. Verify the pre-import summary shows correct counts. Start the import. Verify progress updates during import. Verify completion toast shows final counts.

**Acceptance Scenarios**:

1. **Given** the user selects a backup ZIP, **When** the system analyzes it, **Then** a summary dialog shows: item count, image count, module count, and the import mode (Merge/Replace).
2. **Given** the summary dialog is showing, **When** the user confirms, **Then** the import begins with a visible progress indicator.
3. **Given** the import is in progress, **Then** the user sees the current status (e.g., "Importing items: 25/50").
4. **Given** the import completes, **Then** a toast notification shows a summary (e.g., "Imported 50 items, 30 images, 2 modules").
5. **Given** the summary dialog, **When** the user cancels, **Then** no data is modified.

---

### User Story 3 - Import from Cloud Backup (JSON Format) (Priority: P3)

In cloud mode, backups are exported as JSON (items + modules + settings) rather than SQLite database files. The import feature must handle both formats: the traditional ZIP with a SQLite database file (local mode backups) and the JSON-based ZIP (cloud mode backups).

**Why this priority**: Cloud mode produces a different backup format than local mode. Import must handle both to support cross-mode migration (e.g., restoring a local backup into a cloud deployment, or vice versa).

**Independent Test**: Export a cloud-mode backup (JSON format). Import it into a local-mode instance. Verify all items and modules are restored correctly.

**Acceptance Scenarios**:

1. **Given** a ZIP containing `items.json` and `modules.json` (cloud format), **When** imported, **Then** items and modules are restored from the JSON data.
2. **Given** a ZIP containing `collection.db` (local format), **When** imported, **Then** items are read from the SQLite database and modules from the `modules/` directory.
3. **Given** a ZIP with an unrecognized format (no database, no JSON), **Then** the user sees an error: "Unrecognized backup format."

---

### Edge Cases

- What happens when the backup contains items referencing a module that is not included in the backup? The items are imported with their existing moduleId; the user is warned that some items reference missing module schemas.
- What happens when a "Merge" import has ID conflicts? The backup's version overwrites the existing item (backup wins). This is the expected behavior for restoring from a known-good backup.
- What happens when the backup contains images that already exist in storage? Images with the same filename are overwritten (backup version wins). This is safe because filenames are UUIDs.
- What happens when the import is interrupted (e.g., browser closes)? "Replace" mode should be wrapped in a transaction: either all items are replaced or none are. "Merge" mode processes items individually, so partial imports leave the database in a usable (if incomplete) state.
- What happens when the backup is very large (thousands of items, gigabytes of images)? The system should process items in batches and stream images rather than loading everything into memory.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The application MUST provide an "Import Backup" action accessible from the sidebar (alongside "Export Backup").
- **FR-002**: The user MUST be able to select a backup ZIP file via a file picker (file input in web mode, native dialog in desktop mode).
- **FR-003**: The system MUST analyze the ZIP file and display a pre-import summary (item count, image count, module count) before starting the import.
- **FR-004**: The user MUST choose between "Replace" mode (clear existing data, then import) and "Merge" mode (add/update from backup, preserve existing data not in backup).
- **FR-005**: In "Replace" mode, the system MUST remove all existing items, images, and module schemas before importing the backup's content.
- **FR-006**: In "Merge" mode, items with matching IDs MUST be updated with the backup's version. New items MUST be inserted. Existing items not in the backup MUST be preserved.
- **FR-007**: The system MUST restore all images (originals and thumbnails) from the backup into the configured storage backend (local filesystem or S3).
- **FR-008**: The system MUST restore all module schemas from the backup.
- **FR-009**: The system MUST support both backup formats: SQLite database file (local mode) and JSON files (cloud mode).
- **FR-010**: The system MUST show progress during import for large backups.
- **FR-011**: The system MUST display a completion summary (items imported, images restored, modules loaded, any warnings).
- **FR-012**: Import errors MUST NOT leave the database in a corrupted state. "Replace" mode MUST be atomic (all-or-nothing).
- **FR-013**: The system MUST validate the ZIP file structure before importing and reject unrecognized formats with a clear error.

### Key Entities

- **Backup Archive**: A ZIP file containing either a SQLite database + media + module files (local format) or JSON files + media (cloud format).
- **Import Summary**: A preview of what the backup contains (counts) and the selected import mode, shown before import begins.
- **Import Result**: A post-import report of what was actually imported, including any warnings (e.g., missing module schemas).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A backup exported from any OmniCollect instance can be imported into any other instance (local or cloud) with 100% data fidelity -- all items, images, tags, and modules restored.
- **SC-002**: "Replace" mode import of a 100-item backup completes in under 10 seconds (excluding image transfer time).
- **SC-003**: The user sees a pre-import summary within 2 seconds of selecting the backup file.
- **SC-004**: Zero data corruption in any import failure scenario -- the database remains usable after a failed or interrupted import.
- **SC-005**: Both backup formats (SQLite-based and JSON-based) are importable, enabling cross-mode migration (local to cloud, cloud to local).

## Assumptions

- The import feature is backend-driven: the frontend uploads the ZIP file to a REST endpoint, and the backend handles extraction, validation, and data insertion.
- In web/cloud mode, the ZIP file is uploaded via multipart form data (same pattern as image upload). In desktop/local mode, a file input element is used.
- "Replace" mode uses a database transaction for atomicity (insert all items in one transaction). If any step fails, the transaction rolls back and existing data is preserved.
- "Merge" mode processes items individually (not transactional across all items). This allows partial imports to succeed -- some items may be imported even if others fail.
- Image restoration copies files from the ZIP to the configured media storage (local filesystem or S3). Images are keyed by filename (UUID-based), so conflicts indicate the same image.
- The backup ZIP format is the same format produced by the existing Export Backup feature. No custom format negotiation is needed.
- Settings (theme, etc.) from the backup are NOT imported in v1. Only items, images, and module schemas are restored. Settings import could be added later if needed.
- The "Import Backup" button appears in the sidebar next to "Export Backup". No separate page or route is needed.
