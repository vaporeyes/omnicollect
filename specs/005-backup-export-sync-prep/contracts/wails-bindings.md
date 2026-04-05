# Wails IPC Contracts: Backup Export

**Date**: 2026-04-05
**Feature**: 005-backup-export-sync-prep

## ExportBackup

Generates a ZIP archive containing the database, media, and module
schemas.

**Go signature**:
```go
func (a *App) ExportBackup() (string, error)
```

**Frontend call**:
```typescript
import { ExportBackup } from '../../wailsjs/go/main/App'
const archivePath = await ExportBackup()
```

**Behavior**:
- Opens a native save-file dialog for the collector to choose the
  output location. Default filename includes a UTC timestamp.
- Checkpoints the SQLite WAL to ensure database consistency.
- Creates a ZIP archive containing:
  - `collection.db` (the SQLite database file)
  - `media/originals/*` (all original images)
  - `media/thumbnails/*` (all thumbnails)
  - `modules/*` (all module schema JSON files)
- Writes the archive using streaming compression (not in-memory).
- Returns the path to the created archive on success.
- Rejects with error if the user cancels the dialog, if write fails,
  or if disk space is insufficient.

**Input**: None (save path chosen via dialog).
**Output**: String path to the created archive file.
**Error**: User cancelled, write failure, checkpoint failure.
