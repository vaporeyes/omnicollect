# Tasks: Backup Export & Sync Preparation

**Input**: Design documents from `/specs/005-backup-export-sync-prep/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Not explicitly requested in spec. Test tasks omitted.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go backend**: project root (`backup.go`, `app.go`, `db.go`)
- **Vue frontend**: `frontend/src/App.vue`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No new dependencies needed. Setup is minimal.

- [x] T001 Verify Go standard library `archive/zip` is available (no install needed -- confirm with `go vet ./...`)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core backup functionality in Go.

- [x] T002 Create `backup.go` with ABOUTME comments: implement `createBackupArchive(outputPath string, db *sql.DB) error` that (a) runs `PRAGMA wal_checkpoint(TRUNCATE)` on the database to ensure WAL is flushed, (b) creates a ZIP file at outputPath using `archive/zip`, (c) adds the SQLite database file (`collection.db` from `dbFilePath()`), (d) walks `~/.omnicollect/media/` and adds all files under `media/originals/` and `media/thumbnails/` preserving directory structure, (e) walks `~/.omnicollect/modules/` and adds all `.json` files under `modules/`. Use streaming writes (copy file data directly into zip writer, not buffer in memory).
- [x] T003 Add `ExportBackup` method on App struct in `app.go`: opens a Wails native save-file dialog with default filename `omnicollect-backup-{YYYYMMDD-HHMMSS}.zip`, calls `createBackupArchive(path, a.db)`, returns the output path string. Return empty string + nil if user cancels the dialog.

**Checkpoint**: `ExportBackup()` creates a valid ZIP containing database, media, and modules.

---

## Phase 3: User Story 1 - Export a Full Backup (Priority: P1)

**Goal**: Collector clicks Export Backup button, chooses save location, gets a complete ZIP archive.

**Independent Test**: Add items with images. Click Export Backup. Verify ZIP contains collection.db, media/originals/*, media/thumbnails/*, modules/*.json.

### Implementation for User Story 1

- [x] T004 [US1] Add "Export Backup" button to sidebar in `frontend/src/App.vue`: place below the "New Schema" button. On click, call `ExportBackup()` binding. Show "Exporting..." state while in progress. Show success message with filename on completion. Show error message on failure.
- [x] T005 [US1] Verify export end-to-end: run `wails dev`, create items with images, click Export Backup, extract the ZIP, confirm all expected files are present.

**Checkpoint**: Collector can export a complete backup via the UI. Archive contains all data.

---

## Phase 4: User Story 2 - Verify Timestamps (Priority: P2)

**Goal**: All item modifications record accurate UTC timestamps. Timestamps visible in the UI.

**Independent Test**: Create item, verify timestamps. Update item, verify `updated_at` changed and `created_at` preserved.

### Implementation for User Story 2

- [x] T006 [US2] Audit `insertItem` in `db.go`: verify it sets both `created_at` and `updated_at` to `time.Now().UTC().Format(time.RFC3339)`. Confirm the existing implementation is correct (expected: already correct from Iteration 1).
- [x] T007 [US2] Audit `updateItem` in `db.go`: verify it sets `updated_at` to `time.Now().UTC().Format(time.RFC3339)` and does NOT modify `created_at`. Confirm the existing implementation is correct (expected: already correct -- `updateItem` only sets `updated_at`).
- [x] T008 [US2] Verify the frontend `ItemList.vue` and `CollectionGrid.vue` display `updatedAt` timestamps. Confirm the existing `formatDate` function in `ItemList.vue` converts UTC to local display. Verify `CollectionGrid.vue` does not currently show timestamps -- if missing, add the formatted date below the title in each grid card.

**Checkpoint**: All modification paths produce correct UTC timestamps. Timestamps visible in both list and grid views.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final validation.

- [x] T009 Add ABOUTME comments to `backup.go` (already included in T002 -- verify present)
- [x] T010 Run `quickstart.md` validation: follow all verification steps (export, timestamp verification, edge cases)
- [x] T011 Run `wails build` and verify production binary includes export functionality

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup
- **User Story 1 (Phase 3)**: Depends on Foundational (T002-T003)
- **User Story 2 (Phase 4)**: Independent of US1 (audit only)
- **Polish (Phase 5)**: Depends on all stories complete

### User Story Dependencies

- **User Story 1 (P1)**: Requires Phase 2. Independent.
- **User Story 2 (P2)**: Independent of US1 (audit/verification only).

### Parallel Opportunities

- US1 and US2 can run in parallel after Phase 2 (independent concerns)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1-2: Setup + Foundational (T001-T003)
2. Complete Phase 3: User Story 1 (T004-T005)
3. **STOP and VALIDATE**: Export backup, verify ZIP contents
4. Deploy/demo if ready

### Recommended Execution Order

```
Phase 1 (Setup: verify deps)
  |
Phase 2 (Foundational: backup.go + ExportBackup binding)
  |
  +-- Phase 3 (US1: UI button + verify) [can parallel]
  |
  +-- Phase 4 (US2: timestamp audit) [can parallel]
  |
Phase 5 (Polish: validation + build)
```

---

## Notes

- This is a compact iteration (11 tasks). Most work is in backup.go.
- US2 is primarily an audit/verification exercise, not new code.
- No new dependencies -- Go standard library covers everything.
