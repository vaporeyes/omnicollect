# OmniCollect Frontend

Vue 3 + TypeScript frontend for OmniCollect, embedded in the Wails
desktop shell.

## Structure

```
src/
  main.ts              # App entry, Pinia registration
  App.vue              # Root layout: sidebar + main content area
  style.css            # Global styles
  stores/
    moduleStore.ts     # Fetches/caches module schemas from backend
    collectionStore.ts # Fetches/caches items, handles save/filter/search
  components/
    DynamicForm.vue    # Schema-driven form (create + edit modes)
    FormField.vue      # Single field renderer (type dispatch)
    ModuleSelector.vue # Collection type picker sidebar
    ItemList.vue       # List view with filter + search
    CollectionGrid.vue # Grid view with lazy-loaded thumbnails
    ImageAttach.vue    # Image file picker + thumbnail preview
    ImageLightbox.vue  # Full-resolution image overlay
wailsjs/               # Auto-generated Wails bindings (do not edit)
```

## Development

The frontend runs within `wails dev` (Vite dev server with hot reload).
Do not run `npm run dev` standalone -- Wails manages the dev server.

## Wails Bindings

Backend methods are called via generated TypeScript in `wailsjs/go/main/App`:

- `SaveItem(item)` -- Create or update an item
- `GetItems(query, moduleId)` -- Fetch items with optional search/filter
- `GetActiveModules()` -- Get all loaded module schemas
- `ProcessImage(path)` -- Process image, generate thumbnail
- `SelectImageFile()` -- Open native file dialog

Types are in `wailsjs/go/models.ts` (Item, ModuleSchema, etc.).

## Media URLs

Local media files are served by the Wails AssetServer:

- `/thumbnails/{filename}` -- 300x300 JPEG thumbnails (grid/list views)
- `/originals/{filename}` -- Full-resolution originals (lightbox only)
