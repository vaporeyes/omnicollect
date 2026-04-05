# Component Contracts: Image Processing & Grid

**Date**: 2026-04-05
**Feature**: 003-image-processing-grid

## CollectionGrid.vue

Displays items in a visual grid with lazy-loaded thumbnails.

**Props**:
- `items: Item[]` (required) -- Items to display in the grid.
- `modules: ModuleSchema[]` (required) -- For resolving module names.

**Emits**:
- `select(item: Item)` -- When user clicks an item card to edit.
- `viewImage(item: Item, filename: string)` -- When user clicks a
  thumbnail to view full-resolution.

**Behavior**:
- Renders items as cards in a responsive grid layout.
- Each card shows the first thumbnail (or placeholder if no images).
- All `<img>` tags use `loading="lazy"` for deferred loading.
- Image `src` is `/thumbnails/{filename}` (served by Wails AssetServer).
- Shows item title below the thumbnail.
- Shows placeholder icon for items without images or missing files.

## ImageAttach.vue

File picker for attaching images to an item.

**Props**:
- `images: string[]` (required) -- Current image filenames on the item.

**Emits**:
- `update:images(filenames: string[])` -- Updated image list after
  attach or remove.

**Behavior**:
- Provides an "Add Image" button that opens a native file dialog.
- Calls `ProcessImage(path)` for each selected file.
- Appends the returned filename to the images array.
- Displays thumbnails of currently attached images.
- Allows removing an image (removes from array, does not delete files).
- Shows error messages for rejected files.

## ImageLightbox.vue

Full-resolution image viewer overlay.

**Props**:
- `filename: string` (required) -- Image filename to display.
- `visible: boolean` (required) -- Whether the lightbox is shown.

**Emits**:
- `close()` -- When user closes the lightbox.

**Behavior**:
- Displays the full-resolution original at `/originals/{filename}`.
- Shows a close button and supports clicking outside to close.
- Loaded on demand (only when visible is true).
