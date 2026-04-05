# Data Model: Core Engine (Data & IPC)

**Date**: 2026-04-05
**Feature**: 001-core-engine-data-ipc

## Entities

### Item

A single collectible record stored in the local SQLite database.

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| id | TEXT (UUID v4) | PRIMARY KEY | Unique identifier |
| module_id | TEXT | NOT NULL | References the collection type |
| title | TEXT | NOT NULL | User-visible item name |
| purchase_price | REAL | | Numeric price (no currency) |
| images | TEXT | DEFAULT '[]' | JSON array of image paths |
| attributes | TEXT | DEFAULT '{}' | JSON object of module-specific fields |
| created_at | TEXT | NOT NULL | ISO 8601 timestamp |
| updated_at | TEXT | NOT NULL | ISO 8601 timestamp |

**Notes**:
- `images` stores file path references only. Actual media files live
  on disk outside the database.
- `attributes` is a flat JSON object whose keys are defined by the
  corresponding ModuleSchema. The database does not enforce attribute
  structure; the schema is advisory for UI generation.
- `created_at` and `updated_at` are managed by the backend, not
  user-editable.

### Item FTS Index (items_fts)

FTS5 virtual table for full-text search across items.

| Column | Source | Description |
|--------|--------|-------------|
| title | items.title | Searchable item name |
| attrs_text | Flattened items.attributes | All attribute values concatenated |

**Sync**: Maintained by INSERT/UPDATE/DELETE triggers on the `items`
table. Uses external content mode to avoid data duplication.

### ModuleSchema (in-memory only)

Loaded from JSON files in `~/.omnicollect/modules/` at startup. Not
stored in SQLite.

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique module identifier (e.g., "coins") |
| displayName | string | Human-readable name (e.g., "Coins") |
| description | string | Optional description of the collection type |
| attributes | []AttributeSchema | Ordered list of custom fields |

### AttributeSchema (nested in ModuleSchema)

| Field | Type | Description |
|-------|------|-------------|
| name | string | Attribute key in the JSON blob |
| type | string | Data type: "string", "number", "boolean", "date", "enum" |
| required | boolean | Whether the UI should require this field |
| options | []string | Allowed values (for "enum" type only) |
| display | DisplayHints | Optional rendering hints |

### DisplayHints (nested in AttributeSchema)

| Field | Type | Description |
|-------|------|-------------|
| label | string | Override display label |
| placeholder | string | Input placeholder text |
| widget | string | UI widget hint: "text", "textarea", "dropdown", "rating" |
| group | string | Group name for sectioning attributes |
| order | int | Sort priority within a group |

## Relationships

```
ModuleSchema (in-memory)
    |
    | 1:many (via module_id)
    v
Item (SQLite)
    |
    | 1:1 (triggers)
    v
items_fts (FTS5 virtual table)
```

- One ModuleSchema describes many Items sharing the same `module_id`.
- Each Item has exactly one FTS entry maintained by triggers.
- No foreign key constraint between Item.module_id and ModuleSchema
  (schemas are in-memory, items persist across restarts even if a
  schema file is temporarily removed).

## Schema DDL

```sql
CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    module_id TEXT NOT NULL,
    title TEXT NOT NULL,
    purchase_price REAL,
    images TEXT NOT NULL DEFAULT '[]',
    attributes TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_items_module_id ON items(module_id);

CREATE VIRTUAL TABLE IF NOT EXISTS items_fts USING fts5(
    title,
    attrs_text,
    content='',
    contentless_delete=1
);

-- Triggers omitted here; see research.md R2 for full definitions.
```

## Validation Rules

- `id` MUST be a valid UUID v4 string.
- `module_id` MUST be a non-empty string.
- `title` MUST be a non-empty string.
- `purchase_price` MAY be null (not all items have a price).
- `images` MUST be valid JSON array (may be empty).
- `attributes` MUST be valid JSON object (may be empty).
- ModuleSchema `id` MUST be unique across all loaded schemas.
- ModuleSchema `displayName` MUST be non-empty.
- AttributeSchema `name` MUST be unique within a ModuleSchema.
- AttributeSchema `type` MUST be one of the recognized type strings.
