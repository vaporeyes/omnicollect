# Data Model: Backup Import and Restore

**Branch**: `015-backup-import` | **Date**: 2026-04-08

## No Database Schema Changes

Import uses existing tables and Store interface methods. No new columns, tables, or indexes.

## New Types

### ImportSummary (returned by analyze step)

| Field | Type | Description |
|-------|------|-------------|
| format | string | "local" or "cloud" |
| itemCount | int | Number of items in backup |
| imageCount | int | Number of original images |
| moduleCount | int | Number of module schemas |
| warnings | string[] | Non-fatal issues (e.g., "3 items reference missing module 'stamps'") |
| tempId | string | Temporary identifier for the uploaded file (used in confirm step) |

### ImportRequest (sent to confirm step)

| Field | Type | Description |
|-------|------|-------------|
| tempId | string | References the previously uploaded/analyzed backup |
| mode | string | "replace" or "merge" |

### ImportResult (returned by import step)

| Field | Type | Description |
|-------|------|-------------|
| itemsImported | int | Items inserted or updated |
| imagesRestored | int | Image files copied to storage |
| modulesImported | int | Module schemas saved |
| warnings | string[] | Non-fatal issues encountered during import |

## Backup ZIP Formats

### Local Format (SQLite-based)
```
backup.zip
  collection.db        # SQLite database file
  media/
    originals/         # Original image files
      {uuid}.jpg
    thumbnails/        # Thumbnail image files
      {uuid}.jpg
  modules/
    coins.json         # Module schema files
    books.json
```

### Cloud Format (JSON-based)
```
backup.zip
  items.json           # JSON array of Item objects
  modules.json         # JSON array of ModuleSchema objects
  settings.json        # Settings (ignored during import v1)
```

Note: Cloud format may not include media files (images stored in S3). Media-less imports are valid -- items import without images.

## Store Interface (no changes)

Import uses existing Store methods:
- `InsertItem` / `UpdateItem` for items
- `SaveModule` for module schemas
- `DeleteItem` (in Replace mode, for clearing existing items)
- `QueryItems` (to list existing items for Replace mode deletion)

MediaStore methods:
- `SaveOriginal` / `SaveThumbnail` for image restoration
