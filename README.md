# OmniCollect

A schema-driven desktop collection manager built with Go, Vue 3, and
Wails. Track any type of collection (coins, books, stamps, etc.) by
dropping a JSON schema file into the modules directory. No code changes
needed to add new collection types.

## Architecture

- **Backend**: Go with SQLite (via `modernc.org/sqlite`, CGO-free)
- **Frontend**: Vue 3 (Composition API, TypeScript) with Pinia stores
- **Desktop Shell**: Wails v2 (embeds frontend in native webview)
- **Image Processing**: `disintegration/imaging` for thumbnail generation

## Core Principles

See [Constitution](.specify/memory/constitution.md) for the full set.
Key rules:

1. **Local-First**: SQLite is the source of truth. 100% offline.
2. **Schema-Driven UI**: Forms are generated at runtime from JSON
   schemas. No hardcoded collection-type templates.
3. **Flat Data Architecture**: Single `items` table with JSON
   `attributes` blob. No JOINs for item data.
4. **Performance Protection**: Grid views use compressed thumbnails
   only. Full-resolution images load on demand.
5. **Type-Safe IPC**: All frontend-backend communication via
   Wails-generated TypeScript bindings.

## Prerequisites

- Go 1.25+
- Node.js 18+
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

## Quick Start

```bash
# Install dependencies
go mod tidy
cd frontend && npm install && cd ..

# Run in development mode (hot reload)
wails dev

# Build production binary
wails build
```

The production binary is at `build/bin/omnicollect.app` (macOS).

## Adding a Collection Type

Create a JSON schema file in `~/.omnicollect/modules/`:

```json
{
  "id": "coins",
  "displayName": "Coins",
  "description": "Coin collection",
  "attributes": [
    {
      "name": "year",
      "type": "number",
      "required": true,
      "display": { "label": "Mint Year", "widget": "text" }
    },
    {
      "name": "country",
      "type": "string",
      "required": true,
      "display": { "label": "Country of Origin" }
    },
    {
      "name": "condition",
      "type": "enum",
      "options": ["Poor", "Fair", "Good", "Fine", "Very Fine", "Uncirculated"],
      "display": { "label": "Condition", "widget": "dropdown" }
    }
  ]
}
```

Restart the app. The new collection type appears in the sidebar.

### Supported Attribute Types

| Type | Input Control | Notes |
|------|---------------|-------|
| `string` | Text input | Default widget |
| `number` | Number input | |
| `boolean` | Checkbox | |
| `date` | Date picker | |
| `enum` | Dropdown | Requires `options` array |

### Display Hints

Each attribute can include a `display` object:
- `label`: Override the display name
- `placeholder`: Input placeholder text
- `widget`: Force a specific control (e.g., `"textarea"` for strings)
- `group`: Group attributes into form sections
- `order`: Sort priority within a group

## Project Structure

```
omnicollect/
  main.go              # Wails entry point, AssetServer config
  app.go               # App struct: SaveItem, GetItems, GetActiveModules,
                       #   ProcessImage, SelectImageFile
  db.go                # SQLite init, schema, FTS5 triggers, CRUD
  imaging.go           # Image validation, thumbnail generation
  modules.go           # Module schema loader
  models.go            # Shared types (Item, ModuleSchema, etc.)
  wails.json           # Wails project config
  frontend/
    src/
      main.ts          # Vue app entry, Pinia setup
      App.vue          # Root layout: sidebar + main content
      stores/
        moduleStore.ts     # Module schema cache
        collectionStore.ts # Item cache with search/filter
      components/
        DynamicForm.vue      # Schema-driven form renderer
        FormField.vue        # Type-dispatched field input
        ModuleSelector.vue   # Collection type picker
        ItemList.vue         # List view with search
        CollectionGrid.vue   # Grid view with lazy thumbnails
        ImageAttach.vue      # Image file picker + attachment
        ImageLightbox.vue    # Full-resolution image overlay
    wailsjs/           # Auto-generated Wails bindings (do not edit)
```

## Data Storage

All data is stored locally:

| Data | Location |
|------|----------|
| Database | `~/Library/Application Support/OmniCollect/collection.db` (macOS) |
| Module schemas | `~/.omnicollect/modules/*.json` |
| Original images | `~/.omnicollect/media/originals/` |
| Thumbnails | `~/.omnicollect/media/thumbnails/` |

## Dependencies

### Go

| Package | Purpose |
|---------|---------|
| `github.com/wailsapp/wails/v2` | Desktop framework + IPC |
| `modernc.org/sqlite` | CGO-free SQLite driver |
| `github.com/google/uuid` | UUID v4 generation |
| `github.com/disintegration/imaging` | Image resize/crop |
| `golang.org/x/image/webp` | WebP format support |

### Frontend

| Package | Purpose |
|---------|---------|
| `vue` | UI framework |
| `pinia` | State management |

## Iteration History

1. **Core Engine** (001): Go backend, SQLite schema with FTS5, Wails
   IPC bindings (SaveItem, GetItems, GetActiveModules)
2. **Dynamic Form Engine** (002): Pinia stores, schema-driven form
   renderer, item list with search/filter, edit support
3. **Image Processing & Grid** (003): Thumbnail generation, Wails
   AssetServer for local media, collection grid with lazy loading,
   full-resolution lightbox
