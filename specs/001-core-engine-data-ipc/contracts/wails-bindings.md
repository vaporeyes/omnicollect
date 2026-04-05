# Wails IPC Contracts: Core Engine

**Date**: 2026-04-05
**Feature**: 001-core-engine-data-ipc

These contracts define the Go methods exposed to the Vue 3 frontend
via Wails-generated TypeScript bindings. Each method on the bound
`App` struct becomes a callable function in
`frontend/wailsjs/go/main/App.ts`.

## SaveItem

Creates a new item or updates an existing one (upsert).

**Go signature**:
```go
func (a *App) SaveItem(item Item) (Item, error)
```

**Frontend call**:
```typescript
import { SaveItem } from '../wailsjs/go/main/App';
import { main } from '../wailsjs/go/models';

const saved: main.Item = await SaveItem(item);
```

**Behavior**:
- If `item.id` is empty, generates a new UUID v4 and sets `created_at`.
- If `item.id` exists in the database, updates all fields.
- Always sets `updated_at` to the current timestamp.
- Returns the saved item (with generated ID and timestamps).
- Rejects with error if `module_id` or `title` is empty.

**Input**: `Item` object (id may be empty for create).
**Output**: `Item` object with all fields populated.
**Error**: Validation failure or database write error.

## GetItems

Retrieves items with optional full-text search and module filtering.

**Go signature**:
```go
func (a *App) GetItems(query string, moduleID string) ([]Item, error)
```

**Frontend call**:
```typescript
import { GetItems } from '../wailsjs/go/main/App';

const items = await GetItems("silver dollar", "coins");
const allItems = await GetItems("", "");
```

**Behavior**:
- If `query` is non-empty, performs FTS5 search across title and
  flattened attributes.
- If `moduleID` is non-empty, filters results to that collection type.
- Both filters may be combined.
- If both are empty, returns all items.
- Results are ordered by relevance (when searching) or by
  `updated_at` descending (when browsing).
- Returns an empty array (not null) when no items match.

**Input**: `query` (search text, may be empty), `moduleID` (filter, may be empty).
**Output**: Array of `Item` objects.
**Error**: Database read error.

## GetActiveModules

Returns all module schemas loaded at startup.

**Go signature**:
```go
func (a *App) GetActiveModules() ([]ModuleSchema, error)
```

**Frontend call**:
```typescript
import { GetActiveModules } from '../wailsjs/go/main/App';

const modules = await GetActiveModules();
```

**Behavior**:
- Returns all ModuleSchema objects that were successfully parsed from
  `~/.omnicollect/modules/` at application startup.
- Schemas that failed to parse are excluded (errors logged at startup).
- Returns an empty array if no valid schemas exist.

**Input**: None.
**Output**: Array of `ModuleSchema` objects.
**Error**: Unlikely (schemas are pre-loaded), but possible on internal error.

## Shared Types (Go structs mapped to TypeScript)

### Item

```go
type Item struct {
    ID            string `json:"id"`
    ModuleID      string `json:"moduleId"`
    Title         string `json:"title"`
    PurchasePrice *float64 `json:"purchasePrice"`
    Images        []string `json:"images"`
    Attributes    map[string]interface{} `json:"attributes"`
    CreatedAt     string `json:"createdAt"`
    UpdatedAt     string `json:"updatedAt"`
}
```

### ModuleSchema

```go
type ModuleSchema struct {
    ID          string            `json:"id"`
    DisplayName string            `json:"displayName"`
    Description string            `json:"description,omitempty"`
    Attributes  []AttributeSchema `json:"attributes"`
}
```

### AttributeSchema

```go
type AttributeSchema struct {
    Name     string        `json:"name"`
    Type     string        `json:"type"`
    Required bool          `json:"required,omitempty"`
    Options  []string      `json:"options,omitempty"`
    Display  *DisplayHints `json:"display,omitempty"`
}
```

### DisplayHints

```go
type DisplayHints struct {
    Label       string `json:"label,omitempty"`
    Placeholder string `json:"placeholder,omitempty"`
    Widget      string `json:"widget,omitempty"`
    Group       string `json:"group,omitempty"`
    Order       int    `json:"order,omitempty"`
}
```
