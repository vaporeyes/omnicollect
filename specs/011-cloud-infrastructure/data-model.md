# Data Model: Cloud Infrastructure Migration

**Branch**: `011-cloud-infrastructure` | **Date**: 2026-04-08

## PostgreSQL Schema (per tenant)

Each tenant gets schema `tenant_{id}` containing these tables:

### items

| Column | Type | Notes |
|--------|------|-------|
| id | TEXT PRIMARY KEY | UUID, same as SQLite |
| module_id | TEXT NOT NULL | References modules(id) |
| title | TEXT NOT NULL | Searchable |
| purchase_price | REAL | Nullable |
| images | JSONB NOT NULL DEFAULT '[]' | Array of filenames |
| attributes | JSONB NOT NULL DEFAULT '{}' | Flat key-value attributes |
| search_vector | tsvector | Generated from title + attribute text values |
| created_at | TIMESTAMPTZ NOT NULL | |
| updated_at | TIMESTAMPTZ NOT NULL | |

**Indexes**:
- `idx_items_module_id ON items(module_id)`
- `idx_items_search_vector ON items USING GIN(search_vector)`

**Trigger**: `items_search_update` -- BEFORE INSERT OR UPDATE -- recomputes `search_vector` from `title` and text values extracted from `attributes` JSONB.

### modules

| Column | Type | Notes |
|--------|------|-------|
| id | TEXT PRIMARY KEY | Module identifier |
| display_name | TEXT NOT NULL | Human-readable name |
| description | TEXT | Optional |
| schema_json | JSONB NOT NULL | Full module schema including attributes array |
| created_at | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ NOT NULL DEFAULT NOW() | |

### settings

| Column | Type | Notes |
|--------|------|-------|
| key | TEXT PRIMARY KEY | Settings key (e.g., "theme") |
| value | JSONB NOT NULL | Settings value |

## New Go Interfaces

### Store (database abstraction)

```
type Store interface {
    QueryItems(query, moduleID, filtersJSON string) ([]Item, error)
    InsertItem(item Item) (Item, error)
    UpdateItem(item Item) (Item, error)
    DeleteItem(id string) error
    DeleteItems(ids []string) (int64, error)
    BulkUpdateModule(ids []string, newModuleID string) (int64, error)
    ExportItemsCSV(ids []string, modules []ModuleSchema) (string, error)
    GetModules() ([]ModuleSchema, error)
    SaveModule(schema ModuleSchema) error
    LoadModuleFile(id string) (string, error)
    GetSettings() (string, error)
    SaveSettings(json string) error
    Close() error
}
```

### MediaStore (object storage abstraction)

```
type MediaStore interface {
    SaveOriginal(filename string, data []byte) error
    SaveThumbnail(filename string, data []byte) error
    OriginalURL(filename string) string
    ThumbnailURL(filename string) string
}
```

## Query Translation Reference

| Operation | SQLite | PostgreSQL |
|-----------|--------|------------|
| Text search | `items_fts MATCH ?` | `search_vector @@ plainto_tsquery(?)` |
| Search ranking | `ORDER BY rank` | `ORDER BY ts_rank(search_vector, q)` |
| JSON extract text | `json_extract(attributes, '$.field')` | `attributes->>'field'` |
| JSON extract number | `json_extract(attributes, '$.field') >= ?` | `(attributes->>'field')::numeric >= ?` |
| JSON null check | `json_extract(attributes, '$.field') IS NOT NULL` | `attributes ? 'field'` |

## Image URL Strategy

| Mode | Thumbnail URL | Original URL |
|------|---------------|--------------|
| Local | `/thumbnails/{filename}` (filesystem) | `/originals/{filename}` (filesystem) |
| Cloud | `/thumbnails/{filename}` (backend proxies to S3) | `/originals/{filename}` (backend proxies to S3) |

Frontend URLs are identical in both modes. The backend transparently serves from filesystem or proxies from S3.
