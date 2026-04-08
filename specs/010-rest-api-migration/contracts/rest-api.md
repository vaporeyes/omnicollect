# REST API Contract: OmniCollect v1

**Branch**: `010-rest-api-migration` | **Date**: 2026-04-07
**Base Path**: `/api/v1`

## Items

### List/Search Items
`GET /api/v1/items`

| Query Param | Type | Required | Description |
|-------------|------|----------|-------------|
| query | string | no | FTS5 search query |
| moduleId | string | no | Filter by module |
| filters | string | no | JSON-encoded attribute filter array |

**Response**: `200 OK` -- `Item[]`

### Create/Update Item
`POST /api/v1/items`

**Request Body**: `Item` (JSON). If `id` is empty, creates; if `id` exists, updates.

**Response**: `200 OK` -- `Item` (saved item with generated ID/timestamps)

**Errors**: `400` if moduleId or title missing.

### Delete Item
`DELETE /api/v1/items/{id}`

**Response**: `204 No Content`

**Errors**: `404` if item not found.

### Batch Delete Items
`POST /api/v1/items/batch-delete`

**Request Body**: `{"ids": string[]}`

**Response**: `200 OK` -- `{"deleted": number}`

### Batch Update Module
`POST /api/v1/items/batch-update-module`

**Request Body**: `{"ids": string[], "newModuleId": string}`

**Response**: `200 OK` -- `{"updated": number}`

## Modules

### List Active Modules
`GET /api/v1/modules`

**Response**: `200 OK` -- `ModuleSchema[]`

### Save Custom Module
`POST /api/v1/modules`

**Request Body**: `ModuleSchema` (JSON string of the full schema)

**Response**: `200 OK` -- `ModuleSchema` (saved schema)

**Errors**: `400` if JSON invalid or validation fails.

### Load Module File
`GET /api/v1/modules/{id}/file`

**Response**: `200 OK` -- raw JSON string (module schema file content)

**Errors**: `404` if module not found.

## Images

### Upload Image
`POST /api/v1/images/upload`

**Request**: Multipart form data with file field `image`.

**Response**: `200 OK` -- `ProcessImageResult`

**Errors**: `400` if no file or invalid image format.

## Export

### Export Backup (ZIP)
`GET /api/v1/export/backup`

**Response**: `200 OK` with `Content-Type: application/zip` and `Content-Disposition: attachment; filename="omnicollect-backup-{timestamp}.zip"`. Body is the ZIP file bytes.

### Export Items CSV
`POST /api/v1/export/csv`

**Request Body**: `{"ids": string[]}`

**Response**: `200 OK` with `Content-Type: text/csv` and `Content-Disposition: attachment; filename="omnicollect-export-{count}-items.csv"`. Body is the CSV string.

## Settings

### Load Settings
`GET /api/v1/settings`

**Response**: `200 OK` -- JSON settings object (theme config, etc.)

### Save Settings
`PUT /api/v1/settings`

**Request Body**: JSON settings object.

**Response**: `200 OK`

## Static Media

### Thumbnails
`GET /thumbnails/{filename}` -- serves from media/thumbnails/

### Originals
`GET /originals/{filename}` -- serves from media/originals/

## Error Response Format

All errors return JSON:

```json
{
  "error": "Human-readable error message"
}
```

| Status | Meaning |
|--------|---------|
| 200 | Success |
| 204 | Success (no content, e.g., delete) |
| 400 | Validation error (bad input) |
| 404 | Resource not found |
| 500 | Internal server error |

## CORS Headers (Development Mode)

All responses include:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type`

OPTIONS preflight requests return `204` with these headers.
