# IPC Contract: GetItems with Filters

**Branch**: `007-faceted-filtering` | **Date**: 2026-04-07

## Modified Binding: GetItems

### Current Signature

```
GetItems(query: string, moduleID: string) -> Item[]
```

### Updated Signature

```
GetItems(query: string, moduleID: string, filtersJSON: string) -> Item[]
```

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| query | string | yes (empty = no text search) | FTS5 search query |
| moduleID | string | yes (empty = all modules) | Module filter |
| filtersJSON | string | yes (empty = no attribute filters) | JSON array of filter objects |

### Filter Object Schema

```json
[
  {"field": "condition", "op": "in", "values": ["Mint", "Fine"]},
  {"field": "isGraded", "op": "eq", "value": true},
  {"field": "year", "op": "gte", "value": 1900},
  {"field": "year", "op": "lte", "value": 1950},
  {"field": "purchasePrice", "op": "gte", "value": 50.00}
]
```

### Operators

| Operator | Applies To | Behavior |
|----------|-----------|----------|
| `in` | enum | Match if attribute value is in the provided `values` array (OR) |
| `eq` | boolean | Match if attribute value equals `value` |
| `gte` | number | Match if attribute value >= `value` |
| `lte` | number | Match if attribute value <= `value` |

### Special Fields

- `purchasePrice`: Filtered on the top-level `purchase_price` column directly, not via json_extract
- All other fields: Filtered via `json_extract(attributes, '$.fieldName')`

### Backward Compatibility

Passing an empty string for `filtersJSON` produces the same behavior as the current two-parameter version. Existing callers (command palette `searchAllItems`) must be updated to pass a third empty-string argument.

## UI Contract: FilterBar Component

### Props

| Prop | Type | Required | Description |
|------|------|----------|-------------|
| schema | ModuleSchema / null | yes | Active module schema (null hides the bar) |
| filters | Record<string, AttributeFilter[]> | yes | Current active filters |

### Emits

| Event | Payload | Description |
|-------|---------|-------------|
| update | Record<string, AttributeFilter[]> | User changed a filter; new complete filter state |
| clear | none | User clicked "Clear all filters" |
