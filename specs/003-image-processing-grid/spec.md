# Feature Specification: Image Processing & Grid Display

**Feature Branch**: `003-image-processing-grid`
**Created**: 2026-04-05
**Status**: Draft
**Input**: User description: "Iteration 3: Handle high-resolution image processing, thumbnail generation, local secure rendering, and lazy-loaded collection grid."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Attach Images to a Collection Item (Priority: P1)

A collector is adding or editing a collection item and wants to attach
one or more photographs. They select image files from their local
filesystem. The system processes each image: storing the original for
archival and generating a compressed thumbnail for fast display. The
item's image references are updated and persisted.

**Why this priority**: Image attachment is the foundational capability.
Without it, thumbnails and the grid view have nothing to display.

**Independent Test**: Add an item, attach two images via file picker.
Verify both originals are stored in the media directory and both
thumbnails are generated. Re-open the item and confirm image references
are persisted.

**Acceptance Scenarios**:

1. **Given** a collector is editing or creating an item, **When** they
   select one or more image files, **Then** each image is processed:
   the original is copied to the media storage and a 300x300 compressed
   thumbnail is generated.
2. **Given** an image has been processed, **When** the item is saved,
   **Then** the item's image list references the processed filenames.
3. **Given** a collector attaches a very large image (e.g., 20 MB
   DSLR photo), **When** processing completes, **Then** the thumbnail
   is under 100 KB and the original is preserved at full resolution.
4. **Given** the media directories do not yet exist, **When** the
   first image is processed, **Then** the directories are created
   automatically.

---

### User Story 2 - View Collection as a Grid with Thumbnails (Priority: P2)

A collector browses their collection and sees items displayed in a
visual grid. Each item shows its thumbnail image (if available) along
with the title. The grid loads quickly because it uses compressed
thumbnails, not original images. Images below the visible area load
only when scrolled into view.

**Why this priority**: The grid view is the primary visual interface
for browsing collections. It directly enforces Constitution Principle IV
(Performance & Memory Protection) by using thumbnails exclusively.

**Independent Test**: Save several items with images. Switch to grid
view. Verify thumbnails display correctly, load lazily on scroll, and
that no original high-resolution images are loaded in the grid.

**Acceptance Scenarios**:

1. **Given** items with attached images exist, **When** the collector
   views the collection grid, **Then** each item displays its
   thumbnail image and title.
2. **Given** the grid contains more items than fit on screen, **When**
   the collector scrolls, **Then** off-screen thumbnails load only
   when they enter the viewport (lazy loading).
3. **Given** an item has no attached images, **When** it appears in
   the grid, **Then** a placeholder icon is shown instead of a broken
   image.
4. **Given** the grid is displaying thumbnails, **When** inspecting
   network/memory, **Then** no original full-resolution images are
   loaded into the grid view.

---

### User Story 3 - View Full-Resolution Image (Priority: P3)

A collector clicks on a thumbnail in the grid or item detail view to
see the full-resolution original image. The original loads on demand
in a detail/lightbox view.

**Why this priority**: Full-resolution viewing is important for
inspection but is secondary to the grid browsing experience. It
completes the image lifecycle: attach, browse thumbnails, inspect
originals.

**Independent Test**: Save an item with a large image. View it in
the grid (thumbnail displayed). Click the thumbnail. Verify the
full-resolution original loads in a detail view.

**Acceptance Scenarios**:

1. **Given** an item has attached images, **When** the collector
   clicks a thumbnail, **Then** the full-resolution original image
   is displayed in a detail view.
2. **Given** the collector is viewing a full-resolution image,
   **When** they close the detail view, **Then** the image is
   unloaded from memory and the grid remains responsive.

---

### Edge Cases

- What happens when the collector attaches a non-image file (e.g., PDF)?
  The system MUST reject non-image files with a clear error message.
- What happens when the collector attaches a corrupted image file?
  The system MUST report the error and skip the corrupted file without
  crashing.
- What happens when disk space is insufficient for media storage?
  The system MUST report a clear error message rather than silently
  failing.
- What happens when an item references an image file that has been
  manually deleted from the media directory? The grid MUST show a
  placeholder instead of a broken image.
- What happens when many images are attached to a single item?
  The system MUST handle at least 20 images per item without
  degradation.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide an image attachment capability
  that accepts files from the local filesystem via a file picker.
- **FR-002**: The system MUST accept common image formats (JPEG, PNG,
  WebP, GIF) and reject non-image files with a clear error.
- **FR-003**: For each attached image, the system MUST store the
  original file at full resolution in a dedicated originals directory.
- **FR-004**: For each attached image, the system MUST generate a
  compressed 300x300 thumbnail in a dedicated thumbnails directory.
- **FR-005**: Thumbnails MUST be under 100 KB regardless of original
  image size.
- **FR-006**: The system MUST create media storage directories
  automatically if they do not exist.
- **FR-007**: The system MUST update the item's image references
  after processing and persist them with the item.
- **FR-008**: The collection grid MUST display items with their
  thumbnail images, never original full-resolution images.
- **FR-009**: All images in the grid MUST use lazy loading so that
  off-screen images are not loaded until scrolled into view.
- **FR-010**: The system MUST serve local media files securely to
  the frontend without requiring external HTTP servers.
- **FR-011**: Clicking a thumbnail MUST open the full-resolution
  original in a detail view, loaded on demand.
- **FR-012**: The grid MUST show a placeholder for items without
  images or with missing image files.

### Key Entities

- **Media File**: A processed image stored locally. Has an original
  (full resolution) and a thumbnail (300x300 compressed). Referenced
  by filename in the item's images array.
- **Collection Grid**: A visual layout of collection items showing
  thumbnail images and titles. Uses lazy loading for performance.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A collector can attach an image to an item and see
  the thumbnail in the grid within 3 seconds of selecting the file.
- **SC-002**: The grid view scrolls smoothly (no visible jank) with
  100+ items, each having at least one thumbnail.
- **SC-003**: No original full-resolution image is ever loaded into
  the grid view (verified by inspecting loaded resources).
- **SC-004**: Thumbnails are generated at 300x300 pixels and under
  100 KB for images up to 30 MB in original size.
- **SC-005**: The system handles at least 20 images per item without
  performance degradation in the grid.
- **SC-006**: Full-resolution originals load on demand in under
  2 seconds when a collector clicks a thumbnail.

## Assumptions

- Iterations 1 (Core Engine) and 2 (Dynamic Form Engine) are complete.
- The item's `images` field (JSON array of strings) stores filenames,
  not full paths. The backend resolves filenames to the media directory.
- The media directory is `~/.omnicollect/media/` with `originals/` and
  `thumbnails/` subdirectories.
- Thumbnail dimensions are 300x300 pixels using a fit/crop strategy
  (not stretch). JPEG compression at quality 80 is the default.
- Only single-user local access is assumed. No network serving of
  media files.
- Video files are out of scope for this iteration.
- The grid view is a new display mode alongside the existing list view,
  not a replacement. The collector can switch between grid and list.
