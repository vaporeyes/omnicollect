# IPC Contract: Bulk Action Bindings

**Branch**: `009-bulk-actions` | **Date**: 2026-04-07

## New Binding: DeleteItems

**Signature**: `DeleteItems(ids: string[]) -> { deleted: number }`

Deletes all items with the given IDs in a single atomic transaction. Returns the count of items actually deleted (some IDs may not exist).

| Parameter | Type | Description |
|-----------|------|-------------|
| ids | string[] | Array of item IDs to delete |

| Return Field | Type | Description |
|-------------|------|-------------|
| deleted | number | Count of items successfully deleted |

## New Binding: ExportItemsCSV

**Signature**: `ExportItemsCSV(ids: string[]) -> string`

Queries the database for items with the given IDs and returns a CSV string. Columns: id, title, module (display name), purchasePrice, createdAt, updatedAt, then all unique attribute keys alphabetically.

| Parameter | Type | Description |
|-----------|------|-------------|
| ids | string[] | Array of item IDs to export |

| Return | Type | Description |
|--------|------|-------------|
| csv | string | Complete CSV content including header row |

## New Binding: BulkUpdateModule

**Signature**: `BulkUpdateModule(ids: string[], newModuleID: string) -> { updated: number }`

Updates the module_id of all specified items in a single atomic transaction. Preserves all other fields including attributes.

| Parameter | Type | Description |
|-----------|------|-------------|
| ids | string[] | Array of item IDs to update |
| newModuleID | string | Target module ID |

| Return Field | Type | Description |
|-------------|------|-------------|
| updated | number | Count of items successfully updated |

## New Binding: WriteFile

**Signature**: `WriteFile(path: string, content: string) -> void`

Writes a string to a file at the given path. Used by CSV export after the save dialog provides the path.

| Parameter | Type | Description |
|-----------|------|-------------|
| path | string | Absolute file path from save dialog |
| content | string | File content to write |

## UI Contract: BulkActionBar Component

### Props

| Prop | Type | Required | Description |
|------|------|----------|-------------|
| count | number | yes | Number of selected items |

### Emits

| Event | Payload | Description |
|-------|---------|-------------|
| delete | none | User clicked "Delete Selected" |
| export | none | User clicked "Export Selected as CSV" |
| editModule | none | User clicked "Bulk Edit Module" |
| deselectAll | none | User clicked "Deselect All" |

## UI Contract: SelectionStore

### State

| Field | Type | Description |
|-------|------|-------------|
| selectedIds | Set<string> | Currently selected item IDs |
| lastClickedIndex | number / null | Shift-click range anchor |

### Actions

| Action | Parameters | Description |
|--------|-----------|-------------|
| toggle(id, index) | id: string, index: number | Toggle item selection; set anchor |
| shiftSelect(index, items) | index: number, items: Item[] | Range select from anchor to index |
| selectAll(items) | items: Item[] | Select all visible items |
| clear() | none | Clear all selections |
| isSelected(id) | id: string | Check if item is selected |
