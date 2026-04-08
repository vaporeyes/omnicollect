# Import Contract: REST API + UI Component

**Branch**: `015-backup-import` | **Date**: 2026-04-08

## REST API

### Step 1: Upload and Analyze Backup
`POST /api/v1/import/analyze`

**Request**: Multipart form data with file field `backup` (ZIP file).

**Response**: `200 OK` -- `ImportSummary`
```json
{
  "format": "local",
  "itemCount": 50,
  "imageCount": 30,
  "moduleCount": 2,
  "warnings": [],
  "tempId": "abc123"
}
```

**Errors**:
- `400` -- No file uploaded or invalid ZIP
- `400` -- Unrecognized backup format

### Step 2: Confirm and Execute Import
`POST /api/v1/import/execute`

**Request Body**:
```json
{
  "tempId": "abc123",
  "mode": "replace"
}
```

**Response**: `200 OK` -- `ImportResult`
```json
{
  "itemsImported": 50,
  "imagesRestored": 30,
  "modulesImported": 2,
  "warnings": ["3 items reference module 'stamps' which was not found in the backup"]
}
```

**Errors**:
- `400` -- Invalid tempId or mode
- `404` -- Temp file expired or not found
- `500` -- Import failed (Replace mode: rolled back, no data changed)

## UI Component Contract

### ImportDialog

**Props**: none (self-contained dialog)

**Emits**:
| Event | Payload | Description |
|-------|---------|-------------|
| close | none | User cancelled or import completed |
| imported | ImportResult | Import finished successfully |

**Internal States**:
1. **File Selection**: File input for ZIP file
2. **Analyzing**: Spinner while analyzing the backup
3. **Summary**: Shows counts, format, warnings; user selects Replace/Merge
4. **Importing**: Progress spinner during import execution
5. **Complete**: Shows ImportResult summary; close button

## Temp File Lifecycle

1. Upload stores ZIP in system temp directory with a UUID filename
2. Analyze reads from temp file, returns tempId (the UUID)
3. Execute reads from temp file, processes import, deletes temp file
4. If not executed within 30 minutes, temp file is eligible for cleanup (not actively cleaned in v1; OS temp cleanup handles it)
