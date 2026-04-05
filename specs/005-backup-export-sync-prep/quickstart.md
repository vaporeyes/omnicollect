# Quickstart: Backup Export & Sync Preparation

**Feature**: 005-backup-export-sync-prep

## Prerequisites

- Iterations 1-4 must be complete.
- Some items with images should exist in the collection.

## Setup

```bash
git checkout 005-backup-export-sync-prep
```

## Verify Core Functionality

### 1. Export Backup

1. Click "Export Backup" in the sidebar.
2. Choose a save location in the file dialog.
3. Verify progress indicator appears during export.
4. Verify a `.zip` file is created at the chosen location.
5. Extract the archive and verify contents:
   - `collection.db` exists and is a valid SQLite file.
   - `media/originals/` contains all original images.
   - `media/thumbnails/` contains all thumbnails.
   - `modules/` contains all schema JSON files.

### 2. Timestamp Verification

1. Create a new item and save. Note the `created_at` and `updated_at`
   values (visible in the item list as "last modified").
2. Wait a few seconds, then edit the item and save again.
3. Verify `updated_at` changed to the current time.
4. Verify `created_at` did NOT change.
5. Check that timestamps are in UTC (open the database directly if
   needed: `sqlite3 collection.db "SELECT created_at, updated_at FROM items LIMIT 5"`).

### 3. Edge Cases

1. Export with no items or media -- verify a valid (small) ZIP is
   created containing an empty database.
2. Export with a large collection -- verify the UI does not freeze
   during the export process.

## Build

```bash
wails build
```
