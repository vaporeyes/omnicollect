# OmniCollect

A schema-driven desktop collection manager built with Go, Vue 3, and
Wails. Track any type of collection (coins, books, stamps, etc.) by
dropping a JSON schema file into the modules directory. No code changes
needed to add new collection types.

## Architecture

- **Backend**: Go with SQLite (local) or PostgreSQL (cloud)
- **Frontend**: Vue 3 (Composition API, TypeScript) with Pinia stores
- **Desktop Shell**: Wails v2 (embeds frontend in native webview)
- **Image Processing**: `disintegration/imaging` for thumbnail generation
- **Cloud Storage**: S3-compatible object store (AWS S3, MinIO, R2)
- **Containerization**: Docker multi-stage build, docker-compose for dev

## Core Principles

See [Constitution](.specify/memory/constitution.md) for the full set.
Key rules:

1. **Local-First**: SQLite is the source of truth in local mode. Cloud mode uses PostgreSQL + S3.
2. **Schema-Driven UI**: Forms are generated at runtime from JSON
   schemas. No hardcoded collection-type templates.
3. **Flat Data Architecture**: Single `items` table with JSON
   `attributes` blob. No JOINs for item data.
4. **Performance Protection**: Grid views use compressed thumbnails
   only. Full-resolution images load on demand.
5. **Type-Safe IPC**: All frontend-backend communication via
   Wails-generated TypeScript bindings.
6. **Documentation is Paramount**: README, CLAUDE.md, and spec
   artifacts MUST be updated with every iteration.

## Prerequisites

- Go 1.25+
- Node.js 18+
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

## Quick Start

```bash
# Install dependencies
go mod tidy
cd frontend && npm install && cd ..

# Desktop mode (Wails hot reload)
wails dev

# Standalone HTTP server mode
go run . --serve                # localhost:8080
go run . --serve --port 3001    # custom port

# Build production desktop binary
wails build
```

The production binary is at `build/bin/omnicollect.app` (macOS).

### Standalone Server Mode

Run `go run . --serve` to start OmniCollect as a standalone HTTP server.
The REST API is available at `http://localhost:8080/api/v1/` and the
frontend is served at the root. No Wails desktop shell required.

### Cloud Deployment

Set environment variables to enable cloud backends:

```bash
# PostgreSQL + S3 (cloud mode)
export DATABASE_URL=postgres://user:pass@host:5432/omnicollect
export S3_ENDPOINT=https://s3.amazonaws.com
export S3_BUCKET=omnicollect-media
export S3_ACCESS_KEY=AKIA...
export S3_SECRET_KEY=secret...
go run . --serve
```

Or use Docker:

```bash
docker build -t omnicollect .
docker-compose up   # Runs app + PostgreSQL + MinIO
```

### Data Migration (SQLite to PostgreSQL)

```bash
export DATABASE_URL=postgres://user:pass@host:5432/omnicollect
go run . --migrate --source ~/.omnicollect/collection.db --tenant default
```

## Adding a Collection Type

Use the built-in **Schema Builder** (click "+ New Schema" in the
sidebar) to create collection types visually with a live form preview.
Or create a JSON schema file manually in `~/.omnicollect/modules/`:

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

Restart the app (or save via the Schema Builder for instant reload).
The new collection type appears in the sidebar.

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
- `widget`: Force a specific control (e.g., `"textarea"` for Markdown editor)
- `group`: Group attributes into form sections
- `order`: Sort priority within a group

## Project Structure

```
omnicollect/
  main.go              # Entry point: --serve, --migrate, or Wails desktop
  config.go            # Environment-based config (DATABASE_URL, S3_*, etc.)
  app.go               # Core business logic with Store/MediaStore interfaces
  db.go                # Legacy helpers (dbFilePath for backup)
  imaging.go           # Image validation, thumbnail generation (returns bytes)
  backup.go            # ZIP archive export (database + media + modules)
  modules.go           # Legacy helpers (modulesDir for backup)
  settings.go          # Settings methods delegating to Store
  models.go            # Shared types (Item, ModuleSchema, etc.)
  Dockerfile           # Multi-stage build (Go + Node -> alpine)
  docker-compose.yml   # Dev stack (app + postgres + minio)
  storage/
    db.go              # Store interface (database abstraction)
    media.go           # MediaStore interface (storage abstraction)
    sqlite.go          # SQLiteStore: local SQLite with FTS5
    postgres.go        # PostgresStore: PostgreSQL schema-per-tenant
    local.go           # LocalMediaStore: local filesystem
    s3.go              # S3MediaStore: S3-compatible object store
    migrate.go         # SQLite-to-PostgreSQL migration tool
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
        ItemList.vue         # List view with search + context menu
        CollectionGrid.vue   # Grid view with lazy thumbnails + context menu
        ItemDetail.vue       # Premium split-layout item detail view
        ImageAttach.vue      # Image file picker + attachment
        ImageLightbox.vue    # Full-resolution image overlay
        SchemaBuilder.vue    # Split-pane schema editor
        SchemaVisualEditor.vue # Visual field builder
        SchemaCodeEditor.vue # CodeMirror JSON editor
        SchemaFormPreview.vue # Live form preview
        SettingsPage.vue     # Theme configuration
        CommandPalette.vue   # Spotlight-style search overlay (Cmd/Ctrl+K)
        ContextMenu.vue      # Right-click context menu
        ToastProvider.vue    # Global toast notifications
      stores/
        toastStore.ts        # Toast notification queue
    wailsjs/           # Auto-generated Wails bindings (do not edit)
```

## Data Storage

### Local Mode (default, no env vars)

| Data | Location |
|------|----------|
| Database | `~/Library/Application Support/OmniCollect/collection.db` (macOS) |
| Module schemas | `~/.omnicollect/modules/*.json` |
| Original images | `~/.omnicollect/media/originals/` |
| Thumbnails | `~/.omnicollect/media/thumbnails/` |

### Cloud Mode (env vars set)

| Data | Location |
|------|----------|
| Database | PostgreSQL (schema-per-tenant: `tenant_{id}`) |
| Module schemas | PostgreSQL `modules` table (JSONB) |
| Original images | S3 `originals/` prefix |
| Thumbnails | S3 `thumbnails/` prefix |
| Settings | PostgreSQL `settings` table |

## Dependencies

### Go

| Package | Purpose |
|---------|---------|
| `github.com/wailsapp/wails/v2` | Desktop framework + IPC |
| `modernc.org/sqlite` | CGO-free SQLite driver |
| `github.com/lib/pq` | PostgreSQL driver |
| `github.com/aws/aws-sdk-go-v2` | S3-compatible object storage |
| `github.com/google/uuid` | UUID v4 generation |
| `github.com/disintegration/imaging` | Image resize/crop |
| `golang.org/x/image/webp` | WebP format support |

### Frontend

| Package | Purpose |
|---------|---------|
| `vue` | UI framework |
| `pinia` | State management |
| `vue-codemirror` | CodeMirror 6 editor wrapper |
| `@codemirror/lang-json` | JSON syntax highlighting |

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Cmd/Ctrl+K | Toggle command palette (search items + quick actions) |
| Cmd/Ctrl+F | Focus search bar (switches to list view) |
| Cmd/Ctrl+N | New item for active collection module |
| Escape | Close topmost overlay (palette, lightbox, form, detail, builder, settings) |

The **Command Palette** provides instant access to any item across all
modules. Type keywords to surface quick actions: "new" (create item/schema),
"settings", "backup"/"export". Navigate results with arrow keys and Enter.

Right-click any item in list or grid view for a context menu with
View, Edit, and Delete actions.

## Multi-Select & Bulk Actions

Click checkboxes in list view or selection badges in grid view to select
multiple items. Shift-click to select a contiguous range. When items are
selected, a floating action bar appears at the bottom with:

- **Delete Selected**: Removes all selected items in one atomic operation
- **Export CSV**: Generates a CSV file with all selected items' data
- **Bulk Edit Module**: Reassigns selected items to a different collection type

Selection persists across list/grid view switches and clears on navigation.

## Rich Text / Markdown

Schema attributes with `widget: "textarea"` render a Markdown editor
with a formatting toolbar (bold, italic, heading, lists, links). Content
is stored as raw Markdown in the database and rendered as formatted HTML
in the item detail view. All rendered HTML is sanitized to prevent
script injection.

## Faceted Filtering

When a specific collection type is selected, a collapsible **Filter Bar**
appears above the item list/grid. It dynamically generates filter controls
from the module's JSON schema:

- **Enum attributes**: Multi-select pills (OR logic within the same field)
- **Boolean attributes**: Tri-state toggle (off -> Yes -> No -> off)
- **Number attributes**: Inline min/max range inputs (400ms debounce)
- **Purchase Price**: Always available as a number range filter

Filters combine with AND logic across attributes and with text search.
A "Clear all" button removes all active filters at once.

## Backup & Export

Click "Export Backup" in the sidebar to create a complete ZIP archive
containing the database, all media files, and module schemas. The
archive is self-contained and can be used for manual recovery or
transfer to another machine.

## Iteration History

1. **Core Engine** (001): Go backend, SQLite schema with FTS5, Wails
   IPC bindings (SaveItem, GetItems, GetActiveModules)
2. **Dynamic Form Engine** (002): Pinia stores, schema-driven form
   renderer, item list with search/filter, edit support
3. **Image Processing & Grid** (003): Thumbnail generation, Wails
   AssetServer for local media, collection grid with lazy loading,
   full-resolution lightbox
4. **Schema Visual Builder** (004): Split-pane schema editor with
   visual drag-and-drop field builder, CodeMirror JSON editor with
   bidirectional sync, live form preview, save-to-disk with hot reload
5. **Backup Export & Sync Prep** (005): ZIP archive export of
   database + media + modules, UTC timestamp hardening for future sync
6. **UX & Power User Features** (006): Command palette (Cmd/Ctrl+K)
   for cross-module item search and quick actions, global keyboard
   shortcuts (Cmd+F/N/Esc), right-click context menus, toast
   notifications, item delete with confirmation, premium split-layout
   item detail view with Instrument Serif/Outfit typography
7. **Faceted Filtering** (007): Schema-driven filter bar with enum
   multi-select pills, boolean tri-state toggles, number range inputs,
   purchasePrice filtering, collapsible UI, backend json_extract queries
8. **Markdown Textarea** (008): CodeMirror Markdown editor with
   formatting toolbar for textarea widgets, safe rendered HTML in detail
   views via marked + DOMPurify, global .prose typography class
9. **Multi-Select & Bulk Actions** (009): Checkboxes in list view,
   selection badges in grid view, Shift-click range select, floating
   glassmorphism action bar, atomic batch delete, CSV export with save
   dialog, bulk module reassignment
10. **REST API Migration** (010): Decoupled Wails IPC to standard HTTP
    REST endpoints under /api/v1/, fetch-based frontend client, standalone
    server mode (--serve), multipart image upload, Content-Disposition
    downloads for export, TypeScript API types replacing Wails codegen
11. **Cloud Infrastructure** (011): Storage abstraction layer (Store +
    MediaStore interfaces), PostgreSQL backend with schema-per-tenant
    isolation and tsvector FTS, S3-compatible media storage, SQLite-to-PG
    migration tool (--migrate), Docker multi-stage build, docker-compose
    dev stack, health check endpoint, config via environment variables
