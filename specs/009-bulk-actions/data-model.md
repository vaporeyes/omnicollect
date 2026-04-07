# Data Model: Multi-Select and Bulk Actions

**Branch**: `009-bulk-actions` | **Date**: 2026-04-07

## Existing Entities (no schema changes)

### Item (unchanged)

All bulk operations work on existing items. No new columns, tables, or indexes.

| Operation | SQL Pattern |
|-----------|------------|
| Batch delete | `DELETE FROM items WHERE id IN (?, ?, ...)` within a transaction |
| Bulk module update | `UPDATE items SET module_id = ?, updated_at = ? WHERE id IN (?, ?, ...)` within a transaction |
| CSV export query | `SELECT * FROM items WHERE id IN (?, ?, ...)` |

FTS5 triggers (`items_ad`, `items_au`) handle index maintenance automatically within each transaction.

## New Frontend-Only Types

### SelectionState (selectionStore)

| Field | Type | Description |
|-------|------|-------------|
| selectedIds | Set<string> | Set of currently selected item IDs |
| lastClickedIndex | number / null | Index of last clicked item for Shift-range anchor |

### BulkActionBar Props

| Prop | Type | Description |
|------|------|-------------|
| count | number | Number of selected items |
| visible | boolean | Whether the bar is shown (count > 0) |

## CSV Export Format

| Column Order | Source |
|-------------|--------|
| id | item.id |
| title | item.title |
| module | item.moduleId (resolved to displayName) |
| purchasePrice | item.purchasePrice |
| createdAt | item.createdAt |
| updatedAt | item.updatedAt |
| (dynamic) | All unique attribute keys from selected items, sorted alphabetically |

Items missing an attribute column get an empty cell. Header row always present.

## No Database Migrations

All operations use existing tables and columns. Batch operations use SQL `WHERE id IN (...)` within transactions.
