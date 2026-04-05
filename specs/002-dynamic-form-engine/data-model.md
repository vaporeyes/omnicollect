# Data Model: Dynamic Form Engine

**Date**: 2026-04-05
**Feature**: 002-dynamic-form-engine

## Frontend State Entities

This feature introduces no new backend entities. All data types are
defined by the Wails-generated TypeScript bindings from Iteration 1.
The frontend manages two cached state domains.

### Module Store State

| Field | Type | Description |
|-------|------|-------------|
| modules | ModuleSchema[] | All loaded module schemas |
| loading | boolean | True while fetching from backend |
| error | string or null | Error message from last fetch |

**Source**: `GetActiveModules()` Wails binding, called once at startup.

### Collection Store State

| Field | Type | Description |
|-------|------|-------------|
| items | Item[] | Cached collection items |
| loading | boolean | True while fetching from backend |
| error | string or null | Error message from last operation |
| activeModuleId | string | Currently selected module filter (empty = all) |
| searchQuery | string | Current search text (empty = no search) |

**Source**: `GetItems(query, moduleId)` Wails binding, called on
startup and after each save operation.

### Form State (component-local, not in store)

| Field | Type | Description |
|-------|------|-------------|
| baseFields.title | string | Item title (required) |
| baseFields.purchasePrice | number or null | Optional price |
| attributes | Record<string, any> | Custom field values keyed by attribute name |
| editingItem | Item or null | Null for create mode, populated for edit |
| validationErrors | Record<string, string> | Field name to error message |

## Type Mappings (Schema Attribute Type -> Form Input)

| Schema Type | Input Control | Value Type | Default Value |
|-------------|---------------|------------|---------------|
| "string" | text input | string | "" |
| "number" | number input | number or null | null |
| "boolean" | checkbox | boolean | false |
| "date" | date input | string (ISO) | "" |
| "enum" | select dropdown | string | "" (first option) |

## Validation Rules

- `title` MUST be non-empty (required base field).
- `purchasePrice` MAY be null (optional base field).
- Custom attributes marked `required: true` in the schema MUST be
  non-empty/non-null before submission.
- String fields: non-empty means trimmed length > 0.
- Number fields: non-null and valid number.
- Enum fields: must be one of the schema's `options` values.
- Boolean fields: no validation needed (always has a value).
- Date fields marked required: non-empty string.
