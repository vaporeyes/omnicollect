# Data Model: Smart Folders (Saved Views)

**Branch**: `019-smart-folders` | **Date**: 2026-04-10

## Overview

No new database tables or backend changes. Smart Folders are stored as a JSON array within the existing settings blob under the `smartFolders` key.

## Entities

### SmartFolder

A named, persisted view configuration.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string | Yes | Unique identifier (8-character hex, generated at creation) |
| name | string | Yes | User-given display name (non-empty, no max length enforced) |
| moduleId | string | No | Saved module ID (empty string = "All Types") |
| searchQuery | string | No | Saved search text (empty string = no search) |
| filters | AttributeFilter[] | No | Saved attribute filters (empty array = no filters) |
| tags | string[] | No | Saved tag filter list (empty array = no tag filters) |
| createdAt | string | Yes | ISO 8601 timestamp of creation (for ordering) |

### AttributeFilter (existing type, reused)

Already defined in `collectionStore.ts`.

| Field | Type | Description |
|-------|------|-------------|
| field | string | Attribute name to filter on |
| op | string | Operator: 'in', 'eq', 'gte', 'lte' |
| value | any | Single value (for eq/gte/lte) |
| values | string[] | Multiple values (for 'in' operator) |

## Storage Shape

Within the settings JSON blob:

```json
{
  "theme": { ... },
  "smartFolders": [
    {
      "id": "a1b2c3d4",
      "name": "Mint Liberty Coins",
      "moduleId": "coins",
      "searchQuery": "Liberty",
      "filters": [{"field": "condition", "op": "in", "values": ["Mint State (MS)"]}],
      "tags": [],
      "createdAt": "2026-04-10T12:00:00Z"
    }
  ]
}
```

## Validation Rules

- `name` must be non-empty (whitespace-only rejected)
- `id` must be unique within the array (enforced at creation)
- Duplicate names are allowed (per spec)
- `moduleId` referencing a deleted module is handled gracefully at apply time (falls back to "All Types")
- `filters` referencing removed enum options are silently ignored at apply time

## Lifecycle

1. **Create**: User clicks "Save Current View", types name, presses Enter. New SmartFolder appended to array, settings saved.
2. **Apply**: User clicks folder in sidebar. Store sets module/search/filters/tags from saved state. `activeSmartFolderId` set to this folder's ID.
3. **Rename**: User right-clicks, selects Rename, edits inline, presses Enter. Name updated in array, settings saved.
4. **Delete**: User right-clicks, selects Delete, confirms. Folder removed from array, settings saved. If deleted folder was active, `activeSmartFolderId` cleared.
