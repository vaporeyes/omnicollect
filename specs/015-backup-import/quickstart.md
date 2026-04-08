# Quickstart: Backup Import and Restore

**Branch**: `015-backup-import` | **Date**: 2026-04-08

## Prerequisites

- Existing codebase on the `015-backup-import` branch
- A backup ZIP file (create one via "Export Backup")
- For cloud format testing: a cloud-mode backup (run export in cloud mode)

## No New Dependencies

Uses existing `archive/zip`, `database/sql`, Store/MediaStore interfaces.

## Files to Create

1. **`import.go`** -- All import logic: ZIP format detection, local format reader (SQLite DB), cloud format reader (JSON), Replace mode (transactional clear + insert), Merge mode (per-item upsert), image restoration via MediaStore, module schema import
2. **`frontend/src/components/ImportDialog.vue`** -- Multi-step import modal: file picker -> analyze -> summary with mode selection -> execute -> result display

## Files to Modify

1. **`handlers.go`** -- Add `handleAnalyzeBackup` (POST /api/v1/import/analyze) and `handleExecuteImport` (POST /api/v1/import/execute)
2. **`server.go`** -- Register import routes
3. **`frontend/src/api/types.ts`** -- Add ImportSummary and ImportResult types
4. **`frontend/src/api/client.ts`** -- Add analyzeBackup and executeImport functions
5. **`frontend/src/App.vue`** -- Add "Import Backup" button in sidebar, wire ImportDialog

## Implementation Order

1. Backend: import.go format detection + local format reader
2. Backend: cloud/JSON format reader
3. Backend: Replace mode (transactional)
4. Backend: Merge mode (per-item)
5. Backend: image restoration via MediaStore
6. Backend: REST endpoints (analyze + execute)
7. Frontend: ImportDialog component
8. Frontend: wire into App.vue sidebar
9. Update CLAUDE.md and README

## Acceptance Test Flow

### Replace Mode
1. Export a backup from an instance with items
2. Click "Import Backup" in sidebar
3. Select the backup ZIP file
4. Verify summary shows correct item/image/module counts
5. Select "Replace" mode, click Confirm
6. Verify all existing data replaced with backup content
7. Verify images display correctly

### Merge Mode
8. Add a new item manually
9. Import the same backup in "Merge" mode
10. Verify the manually-added item is preserved
11. Verify backup items are restored alongside it

### Cross-Format
12. Export a cloud-mode backup (JSON format)
13. Import it into a local-mode instance -- verify items + modules restored
14. Export a local-mode backup
15. Import it into a cloud-mode instance -- verify items + modules restored

### Error Cases
16. Try importing a non-ZIP file -- verify error message
17. Try importing a ZIP without collection.db or items.json -- verify format error
