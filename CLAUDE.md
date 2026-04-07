# OmniCollect Development Guidelines

## Project Overview

Schema-driven desktop collection manager. Go backend + Vue 3 frontend
via Wails v2. See [Constitution](.specify/memory/constitution.md) for
immutable engineering principles.

## Tech Stack

- **Go 1.25+**: Backend (SQLite, image processing, backup, Wails bindings)
- **Vue 3 + TypeScript**: Frontend (Composition API, Pinia stores)
- **Wails v2**: Desktop shell, type-safe IPC, AssetServer
- **SQLite**: Local database via `modernc.org/sqlite` (CGO-free)
- **disintegration/imaging**: Thumbnail generation (CGO-free)
- **vue-codemirror**: CodeMirror 6 JSON editor for schema builder

## Project Structure

```
main.go          # Entry point, Wails config, AssetServer handler
app.go           # Bound methods: SaveItem, GetItems, GetActiveModules,
                 #   ProcessImage, SelectImageFile, SaveCustomModule,
                 #   LoadModuleFile, ExportBackup
db.go            # SQLite init, DDL, FTS5 triggers, CRUD helpers
imaging.go       # Image validation, copy, thumbnail generation
backup.go        # ZIP archive export (database + media + modules)
modules.go       # Module schema loader, save, find helpers
models.go        # Shared Go types (Item, ModuleSchema, ProcessImageResult)
frontend/src/
  stores/        # Pinia: moduleStore, collectionStore, selectionStore,
                 #   toastStore
  components/    # DynamicForm, FormField, ItemList, CollectionGrid,
                 #   ModuleSelector, ImageAttach, ImageLightbox,
                 #   SchemaBuilder, SchemaVisualEditor,
                 #   SchemaCodeEditor, SchemaFormPreview,
                 #   ItemDetail, SettingsPage, ToastProvider,
                 #   ContextMenu, CommandPalette, FilterBar,
                 #   MarkdownEditor, MarkdownRenderer,
                 #   BulkActionBar
```

## Commands

```bash
wails dev        # Development with hot reload
wails build      # Production binary to build/bin/
go vet ./...     # Lint Go code
go mod tidy      # Resolve dependencies
```

## Key Conventions

- All Go files start with two-line ABOUTME comments
- Wails bindings: exported methods on App struct (no context.Context param)
- Frontend calls Wails bindings via generated `wailsjs/go/main/App`
- Media served via AssetServer at /thumbnails/ and /originals/
- Grid views MUST use thumbnails only (Constitution Principle IV)
- No hardcoded collection-type templates (Constitution Principle II)
- README and CLAUDE.md MUST be updated every iteration (Principle VI)
- Module schemas in `~/.omnicollect/modules/*.json`
- Database at user config dir (`os.UserConfigDir()`)
- Global shortcuts: Cmd/Ctrl+K (command palette), Cmd/Ctrl+F (search),
  Cmd/Ctrl+N (new item), Escape (close overlays)
- Command palette (`CommandPalette.vue`): cross-module search via
  `collectionStore.searchAllItems()`, quick actions via keyword matching
- Toast notifications via `useToastStore` (replace all alert() calls)
- Context menus on items in list/grid views via `ContextMenu.vue`
- Faceted filtering via `FilterBar.vue`: collapsible bar generated from
  active module schema; enum pills (multi-select OR), boolean tri-state
  toggles (off/true/false), inline number range min/max inputs
- `GetItems(query, moduleID, filtersJSON)` accepts JSON filter payload;
  backend uses `json_extract()` for attribute filters, direct column for
  `purchasePrice`; `collectionStore.activeFilters` manages filter state
- Markdown support: `widget: "textarea"` schema attributes use
  `MarkdownEditor.vue` (CodeMirror + `@codemirror/lang-markdown`) in forms
  and `MarkdownRenderer.vue` (marked + DOMPurify) in detail views.
  Global `.prose` class in `style.css` styles rendered Markdown.
  Dependencies: `@codemirror/lang-markdown`, `marked`, `dompurify`
- Multi-select via `selectionStore` (Pinia): Set<string> of selected IDs,
  Shift-click range, select-all. `BulkActionBar.vue` floating bar with
  bulk delete, CSV export, module reassignment. Bindings: `DeleteItems`,
  `ExportItemsCSV`, `BulkUpdateModule`

## Wails Bindings (App struct methods)

| Method | Purpose |
|--------|---------|
| `SaveItem(item)` | Create or update a collection item |
| `DeleteItem(id)` | Remove a collection item by ID |
| `GetItems(query, moduleId, filtersJSON)` | Fetch items with search, module, and attribute filters |
| `GetActiveModules()` | Get all loaded module schemas |
| `ProcessImage(path)` | Process image, generate thumbnail |
| `SelectImageFile()` | Open native file dialog for images |
| `SaveCustomModule(json)` | Write module schema to disk, hot reload |
| `LoadModuleFile(moduleId)` | Read module schema JSON for editing |
| `DeleteItems(ids)` | Batch delete items in one transaction |
| `ExportItemsCSV(ids)` | Generate CSV + save dialog for selected items |
| `BulkUpdateModule(ids, newModuleID)` | Reassign items to different module |
| `ExportBackup()` | Create ZIP archive of all data |

## Data Locations

- Database: `~/Library/Application Support/OmniCollect/collection.db`
- Modules: `~/.omnicollect/modules/`
- Media: `~/.omnicollect/media/originals/` and `thumbnails/`

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->

## Active Technologies
- Go 1.25+ (backend), TypeScript + Vue 3 (frontend) + Wails v2 (IPC/bindings), Pinia (state), Vue Composition API (006-command-palette)
- SQLite via modernc.org/sqlite (existing FTS5 full-text search) (006-command-palette)
- SQLite via modernc.org/sqlite (FTS5 full-text search, JSON attributes column) (007-faceted-filtering)
- Go 1.25+ (backend, no changes needed), TypeScript + Vue 3 (frontend) + Wails v2, Pinia, vue-codemirror (existing), CodeMirror markdown extensions (new), marked (new), DOMPurify (new) (008-markdown-textarea)
- SQLite (no changes -- raw Markdown stored as string in existing JSON attributes) (008-markdown-textarea)
- Go 1.25+ (backend -- new bindings), TypeScript + Vue 3 (frontend) + Wails v2 (IPC/bindings), Pinia (state), Vue Composition API (009-bulk-actions)
- SQLite via modernc.org/sqlite (batch delete in transaction, CSV query) (009-bulk-actions)

## Recent Changes
- 006-command-palette: Added Go 1.25+ (backend), TypeScript + Vue 3 (frontend) + Wails v2 (IPC/bindings), Pinia (state), Vue Composition API
