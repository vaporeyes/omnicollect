# Quickstart: Image Processing & Grid Display

**Feature**: 003-image-processing-grid

## Prerequisites

- Iterations 1 and 2 must be complete and functional.
- At least one module schema in `~/.omnicollect/modules/`.
- Some test images (JPEG, PNG) available on your filesystem.

## Setup

```bash
git checkout 003-image-processing-grid
cd frontend && npm install && cd ..
```

## Development

```bash
wails dev
```

## Verify Core Functionality

### 1. Image Attachment

1. Click a module type to create a new item.
2. Fill in the title and other fields.
3. Click "Add Image" in the image attachment section.
4. Select a JPEG or PNG photo from your filesystem.
5. Verify a thumbnail preview appears in the form.
6. Save the item.
7. Verify files exist:
   - `~/.omnicollect/media/originals/` contains the original.
   - `~/.omnicollect/media/thumbnails/` contains a 300x300 JPEG.

### 2. Grid View

1. Switch to grid view (toggle button).
2. Verify items with images show their thumbnails.
3. Verify items without images show a placeholder.
4. Scroll through the grid -- off-screen images should load lazily.
5. Open browser dev tools (if available) and verify no requests
   to `/originals/` are made during grid browsing.

### 3. Full-Resolution Viewing

1. In the grid view, click a thumbnail.
2. Verify the full-resolution original loads in a lightbox overlay.
3. Close the lightbox and verify the grid remains responsive.

### 4. Error Handling

1. Try attaching a non-image file (e.g., a .txt file).
2. Verify an error message is shown and the file is rejected.

### 5. Multiple Images

1. Edit an item and attach 3-5 images.
2. Save and verify all thumbnails appear.
3. View in grid -- the first thumbnail is shown on the card.

## Build

```bash
wails build
```

Verify the production binary serves thumbnails correctly from
the local media directory.
