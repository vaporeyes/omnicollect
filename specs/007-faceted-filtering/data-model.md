# Data Model: Schema-Driven Faceted Filtering

**Branch**: `007-faceted-filtering` | **Date**: 2026-04-07

## Existing Entities (modified)

### Item (no schema change)

The `attributes` JSON column is now queried with `json_extract()` for filtering. No new columns, indexes, or tables.

| Field | Type | Filter Role |
|-------|------|-------------|
| id | string (UUID) | Not filterable |
| moduleId | string | Existing module filter |
| title | string | Existing FTS5 search |
| purchasePrice | number (nullable) | Filterable as number range (direct column, not json_extract) |
| images | string[] | Not filterable |
| attributes | JSON object | Filterable via json_extract per schema definition |
| createdAt | string (RFC3339) | Not filterable |
| updatedAt | string (RFC3339) | Not filterable |

### ModuleSchema (no change)

Schema attributes drive filter bar generation. Relevant attribute properties:

| Attribute Property | Used By Filter |
|--------------------|---------------|
| name | Filter field key (maps to json_extract path) |
| type | Determines filter control: "enum" -> pills, "boolean" -> tri-state, "number" -> range |
| options | Enum pill values (array of strings) |
| display.label | Filter group label in UI |

## New Frontend-Only Types

### AttributeFilter

Represents a single active filter constraint. Lives in the collection store, not persisted.

| Field | Type | Description |
|-------|------|-------------|
| field | string | Attribute name (or "purchasePrice" for the base field) |
| op | "in" / "eq" / "gte" / "lte" | Filter operator |
| values | string[] | For "in" operator: selected enum values |
| value | any | For "eq"/"gte"/"lte": single comparison value |

### FilterState

Reactive map in the collection store.

| Key | Value | Description |
|-----|-------|-------------|
| attribute name | AttributeFilter[] | Active filters for that attribute (e.g., enum has one "in" filter, number may have "gte" and "lte") |

## Backend Filter Payload

The `GetItems` binding receives filters as a JSON string. The backend parses it into a list of filter objects and builds dynamic WHERE clauses.

### Filter Object Schema

```
{ "field": string, "op": "in"|"eq"|"gte"|"lte", "value"?: any, "values"?: string[] }
```

### Query Generation Rules

- `"in"` operator: `json_extract(attributes, '$.field') IN (?, ?, ...)`
- `"eq"` operator: `json_extract(attributes, '$.field') = ?`
- `"gte"` operator: `json_extract(attributes, '$.field') >= ?` (or `purchase_price >= ?` for base field)
- `"lte"` operator: `json_extract(attributes, '$.field') <= ?` (or `purchase_price <= ?` for base field)
- Null handling: `json_extract(attributes, '$.field') IS NOT NULL` added when any filter targets that field

## No Database Migrations

This feature queries existing data using SQLite JSON functions. No schema changes, no new tables, no new indexes.
