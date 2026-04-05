# Data Model: Backup Export & Sync Preparation

**Date**: 2026-04-05
**Feature**: 005-backup-export-sync-prep

## Entities

### Backup Archive (filesystem output)

A self-contained ZIP file containing all data needed to restore or
transfer a collection.

**Archive structure**:
```
omnicollect-backup-20260405-120000.zip
  collection.db              # SQLite database
  media/
    originals/               # Full-resolution images
      {uuid}.jpg
      {uuid}.png
    thumbnails/              # 300x300 JPEG thumbnails
      {uuid}.jpg
  modules/                   # Module schema definitions
    coins.json
    vinyl-records.json
```

### Timestamp Fields (existing, verified)

The `items` table already has:

| Field | Format | Set on create | Set on update |
|-------|--------|--------------|---------------|
| `created_at` | ISO 8601 UTC (RFC 3339) | Yes | No (preserved) |
| `updated_at` | ISO 8601 UTC (RFC 3339) | Yes | Yes (always current) |

This iteration verifies these are correctly managed across all code
paths.
