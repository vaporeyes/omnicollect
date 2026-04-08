# Tasks: Backup Import and Restore

**Input**: Design documents from `/specs/015-backup-import/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/import-contract.md, quickstart.md

**Tests**: Not explicitly requested.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Types and API client functions shared by all stories

- [x] T001 [P] Add `ImportSummary` and `ImportResult` types to `frontend/src/api/types.ts`: ImportSummary (format, itemCount, imageCount, moduleCount, warnings, tempId), ImportResult (itemsImported, imagesRestored, modulesImported, warnings)
- [x] T002 [P] Add `analyzeBackup(file: File)` and `executeImport(tempId: string, mode: string)` functions to `frontend/src/api/client.ts`: analyzeBackup POSTs multipart with file field `backup` to `/api/v1/import/analyze`; executeImport POSTs JSON to `/api/v1/import/execute`

**Checkpoint**: Frontend types and API functions defined

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core import logic in Go

- [x] T003 Create `import.go`: implement `detectBackupFormat(zipReader *zip.Reader) (string, error)` that scans entries for `collection.db` (returns "local") or `items.json` (returns "cloud"), or returns error for unrecognized format
- [x] T004 Implement `analyzeBackup(zipPath string) (ImportSummary, error)` in `import.go`: open ZIP, detect format, count items (by querying embedded SQLite or parsing items.json length), count images (scan `media/originals/` entries), count modules (scan `modules/*.json` or parse modules.json), return summary with tempId = the temp file path
- [x] T005 Implement `readItemsFromLocalBackup(zipReader *zip.Reader) ([]storage.Item, []storage.ModuleSchema, error)` in `import.go`: extract `collection.db` to temp file, open as SQLite, query all items, read module JSON files from `modules/` entries, return items + modules
- [x] T006 Implement `readItemsFromCloudBackup(zipReader *zip.Reader) ([]storage.Item, []storage.ModuleSchema, error)` in `import.go`: parse `items.json` into `[]storage.Item`, parse `modules.json` into `[]storage.ModuleSchema`, return both
- [x] T007 Implement `restoreImages(zipReader *zip.Reader, mediaStore storage.MediaStore) (int, error)` in `import.go`: iterate ZIP entries under `media/originals/` and `media/thumbnails/`, read each file's bytes, call `mediaStore.SaveOriginal` or `SaveThumbnail`, return count of images restored

**Checkpoint**: Core import functions compile; `go vet ./...` passes

---

## Phase 3: User Story 1 - Restore from Backup ZIP (Priority: P1) MVP

**Goal**: Replace and Merge mode import with items, images, and modules restored.

**Independent Test**: Export backup, import with Replace on fresh instance, verify all data restored. Import with Merge, verify existing data preserved + backup data added.

### Implementation for User Story 1

- [x] T008 [US1] Implement `executeReplace(store storage.Store, items []storage.Item, modules []storage.ModuleSchema) error` in `import.go`: delete all existing items (query all IDs, batch delete in transaction), delete all existing modules, insert all backup items, insert all backup modules -- all within a single transaction for atomicity
- [x] T009 [US1] Implement `executeMerge(store storage.Store, items []storage.Item, modules []storage.ModuleSchema) (int, error)` in `import.go`: for each item, try update (if exists) or insert (if new); for each module, save (upsert); return count of items processed
- [x] T010 [US1] Add `handleAnalyzeBackup` handler in `handlers.go`: parse multipart file upload, save to temp file, call `analyzeBackup`, return ImportSummary JSON
- [x] T011 [US1] Add `handleExecuteImport` handler in `handlers.go`: parse ImportRequest JSON (tempId + mode), open ZIP from temp path, detect format, read items/modules, execute replace or merge, restore images, delete temp file, return ImportResult JSON
- [x] T012 [US1] Register import routes in `server.go`: `POST /api/v1/import/analyze` and `POST /api/v1/import/execute`
- [x] T013 [US1] Verify Replace mode: export a backup, import with replace, verify all items/images/modules restored correctly via curl or browser
- [x] T014 [US1] Verify Merge mode: add items manually, import backup with merge, verify manual items preserved + backup items added

**Checkpoint**: Full import/export round-trip works in both modes

---

## Phase 4: User Story 2 - Import Progress and Confirmation (Priority: P2)

**Goal**: Pre-import summary dialog, mode selection, progress spinner, completion toast.

**Independent Test**: Select a backup, verify summary shows correct counts, confirm, verify progress spinner during import, verify completion toast.

### Implementation for User Story 2

- [x] T015 [US2] Create `frontend/src/components/ImportDialog.vue`: multi-step dialog with states -- (1) file picker with drag-drop zone, (2) analyzing spinner, (3) summary display (format, item/image/module counts, warnings) with Replace/Merge radio buttons and Confirm/Cancel, (4) importing spinner, (5) result summary with Close button
- [x] T016 [US2] Style `ImportDialog.vue` with app theme: glassmorphism overlay (matching existing dialogs), Instrument Serif title, Outfit body text, accent-colored confirm button, error-styled warnings
- [x] T017 [US2] Modify `frontend/src/App.vue`: add "Import Backup" button in sidebar (next to "Export Backup"), wire to show ImportDialog; on `@imported` event, refresh modules + items via stores, show success toast

**Checkpoint**: Full UI flow works: file select -> summary -> confirm -> progress -> result toast

---

## Phase 5: User Story 3 - Cross-Format Import (Priority: P3)

**Goal**: Both SQLite and JSON backup formats importable, enabling local-to-cloud and cloud-to-local migration.

**Independent Test**: Export from local mode, import into cloud mode. Export from cloud mode, import into local mode.

### Implementation for User Story 3

- [x] T018 [US3] Verify cloud format (JSON) import works in `import.go`: ensure `readItemsFromCloudBackup` correctly handles items without images (cloud backups may not include media), handles missing settings.json gracefully
- [x] T019 [US3] Verify local format (SQLite) import into PostgreSQL Store: ensure items read from embedded SQLite correctly insert into PostgresStore with JSONB attributes and tags
- [x] T020 [US3] Add validation in `handleAnalyzeBackup` for unrecognized format: return 400 with descriptive error message when ZIP contains neither `collection.db` nor `items.json`

**Checkpoint**: Cross-mode migration works both directions

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, error handling, cleanup

- [x] T021 Add temp file cleanup: in `handleExecuteImport`, ensure temp file is deleted after import (success or failure) via `defer os.Remove`
- [x] T022 Add warning for missing modules: when importing items that reference moduleIds not found in the backup's module schemas, include a warning in ImportResult
- [x] T023 [P] Update `CLAUDE.md`: document import.go, ImportDialog component, REST endpoints, Replace/Merge modes
- [x] T024 [P] Update `README.md`: add Import/Restore section, update iteration history
- [x] T025 Run quickstart.md acceptance test flow (all 17 steps) and fix any issues

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: No dependency on Phase 1 (Go + TS are independent)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (needs import functions)
- **User Story 2 (Phase 4)**: Depends on Phase 1 (needs API client) + Phase 3 (needs working endpoints)
- **User Story 3 (Phase 5)**: Depends on Phase 3 (needs import working for both formats)
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational. MVP -- backend import works.
- **US2 (P2)**: Depends on US1 (needs working endpoints for the UI to call).
- **US3 (P3)**: Depends on US1 (needs import working, then verify cross-format).

### Parallel Opportunities

- T001 and T002 (Phase 1 frontend) and T003-T007 (Phase 2 backend) -- different languages
- T005 (local reader) and T006 (cloud reader) -- different functions
- T023 and T024 (docs) -- different files

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 2 (Foundational) -- import logic in Go
2. Phase 3 (US1) -- REST endpoints + replace/merge
3. **STOP and VALIDATE**: curl test -- upload ZIP, analyze, execute replace, verify data
4. This delivers working import testable from command line

### Incremental Delivery

1. Phase 1 + Phase 2 = types + import logic
2. Phase 3 (US1) = backend endpoints (testable with curl)
3. Phase 4 (US2) = frontend UI (full user experience)
4. Phase 5 (US3) = cross-format verification
5. Phase 6 = polish + docs

---

## Notes

- Two-step flow: upload+analyze -> confirm with tempId -> execute. Avoids re-uploading large ZIPs.
- Replace mode: transactional (atomic). Merge mode: per-item (partial success OK).
- Format detection by file presence: `collection.db` = local, `items.json` = cloud.
- Images are best-effort after DB commit (can't participate in DB transaction).
- Settings NOT imported in v1 (only items, images, modules).
- v1 progress: synchronous with spinner. SSE deferred to v2.
