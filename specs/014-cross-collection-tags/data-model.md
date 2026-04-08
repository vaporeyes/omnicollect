# Data Model: Cross-Collection Tags

**Branch**: `014-cross-collection-tags` | **Date**: 2026-04-08

## Modified Entity: Item

A new `tags` field is added to the Item struct and both database schemas.

### Item (updated)

| Field | Type (Go) | SQLite | PostgreSQL | Notes |
|-------|-----------|--------|------------|-------|
| ... | (existing fields unchanged) | | | |
| Tags | []string | TEXT (JSON array) | JSONB (array) | New. Default `[]`. Stored lowercase. |

### Storage.Item Struct Addition

```
Tags []string `json:"tags"`
```

### DDL Changes

**SQLite**:
```sql
ALTER TABLE items ADD COLUMN tags TEXT NOT NULL DEFAULT '[]';
```

**PostgreSQL** (in tenant schema DDL):
```sql
ALTER TABLE items ADD COLUMN tags JSONB NOT NULL DEFAULT '[]';
CREATE INDEX idx_items_tags ON items USING GIN(tags);
```

### FTS/Search Index Updates

**SQLite FTS5 triggers**: The insert/update triggers must include tag text in the indexed content.

**PostgreSQL tsvector trigger**: Include tag values in the search vector computation.

## New Store Interface Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| GetAllTags | `() ([]TagCount, error)` | Returns all distinct tags with item counts |
| RenameTag | `(oldName, newName string) (int64, error)` | Renames a tag across all items; returns count |
| DeleteTag | `(name string) (int64, error)` | Removes a tag from all items; returns count |

### TagCount Type

```
type TagCount struct {
    Name  string `json:"name"`
    Count int    `json:"count"`
}
```

## Query Modifications

### GetItems -- Tag Filter

The `GetItems` query gains an optional `tags` parameter (JSON array of tag names). When present, only items whose `tags` array contains at least one of the specified values are returned (OR logic).

**SQLite**: `WHERE EXISTS (SELECT 1 FROM json_each(items.tags) WHERE value IN (?, ...))`

**PostgreSQL**: `WHERE tags ?| array[?, ...]`

This filter combines with existing filters (text search, module, attributes) using AND logic.

## REST API Changes

### Modified Endpoints

| Endpoint | Change |
|----------|--------|
| GET /api/v1/items | New `tags` query param (JSON array of tag names) |
| POST /api/v1/items | Item body includes `tags` field |

### New Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| /api/v1/tags | GET | Returns all tags with item counts |
| /api/v1/tags/rename | POST | Body: `{oldName, newName}` -- renames across all items |
| /api/v1/tags/{name} | DELETE | Removes tag from all items |

## TypeScript Type Changes

```typescript
interface Item {
  // ... existing fields ...
  tags: string[]  // New
}

interface TagCount {
  name: string
  count: number
}
```
