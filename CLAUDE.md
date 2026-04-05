# OmniCollect Development Guidelines

## Project Overview

Schema-driven desktop collection manager. Go backend + Vue 3 frontend
via Wails v2. See [Constitution](.specify/memory/constitution.md) for
immutable engineering principles.

## Tech Stack

- **Go 1.25+**: Backend (SQLite, image processing, Wails bindings)
- **Vue 3 + TypeScript**: Frontend (Composition API, Pinia stores)
- **Wails v2**: Desktop shell, type-safe IPC, AssetServer
- **SQLite**: Local database via `modernc.org/sqlite` (CGO-free)
- **disintegration/imaging**: Thumbnail generation (CGO-free)

## Project Structure

```
main.go          # Entry point, Wails config, AssetServer handler
app.go           # Bound methods: SaveItem, GetItems, GetActiveModules,
                 #   ProcessImage, SelectImageFile
db.go            # SQLite init, DDL, FTS5 triggers, CRUD helpers
imaging.go       # Image validation, copy, thumbnail generation
modules.go       # Module schema loader from ~/.omnicollect/modules/
models.go        # Shared Go types (Item, ModuleSchema, ProcessImageResult)
frontend/src/
  stores/        # Pinia: moduleStore, collectionStore
  components/    # DynamicForm, FormField, ItemList, CollectionGrid,
                 #   ModuleSelector, ImageAttach, ImageLightbox
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
- Module schemas in `~/.omnicollect/modules/*.json`
- Database at user config dir (`os.UserConfigDir()`)

## Data Locations

- Database: `~/Library/Application Support/OmniCollect/collection.db`
- Modules: `~/.omnicollect/modules/`
- Media: `~/.omnicollect/media/originals/` and `thumbnails/`

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
