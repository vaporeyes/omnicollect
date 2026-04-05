# Tasks: Image Processing & Grid Display

**Input**: Design documents from `/specs/003-image-processing-grid/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Not explicitly requested in spec. Test tasks omitted.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go backend**: project root (`imaging.go`, `main.go`, `models.go`)
- **Vue frontend**: `frontend/src/components/`
- **Media storage**: `~/.omnicollect/media/originals/` and `thumbnails/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add imaging dependencies and the ProcessImageResult type.

- [x] T001 Add Go dependencies: `github.com/disintegration/imaging` and `golang.org/x/image` in `go.mod` via `go get`
- [x] T002 Add `ProcessImageResult` struct to `models.go` with fields: Filename, OriginalPath, ThumbnailPath, Width, Height, Format (all with JSON struct tags per contracts/wails-bindings.md)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Image processing backend and Wails AssetServer configuration. MUST be complete before any user story.

**CRITICAL**: No user story work can begin until this phase is complete.

- [x] T003 Create `imaging.go`: implement `processImage(sourcePath string) (ProcessImageResult, error)` that (a) validates the file is a supported image via `image.DecodeConfig` with blank imports for JPEG/PNG/GIF/WebP decoders, (b) generates a UUID filename, (c) copies the original to `~/.omnicollect/media/originals/{uuid}.{ext}`, (d) generates a 300x300 JPEG thumbnail via `imaging.Fill` with `imaging.Center` anchor, `imaging.Lanczos` filter, quality 80, `AutoOrientation(true)`, saves to `~/.omnicollect/media/thumbnails/{uuid}.jpg`, (e) returns ProcessImageResult with dimensions and format. Create media directories automatically if missing.
- [x] T004 Add `ProcessImage` method on App struct in `app.go`: calls `processImage()` from T003, exposed to frontend via Wails binding. Validate sourcePath is non-empty, return error for invalid/unsupported files.
- [x] T005 Configure Wails AssetServer custom handler in `main.go`: create `newLocalFileHandler()` that returns an `http.ServeMux` mapping `/thumbnails/` to `http.FileServer(http.Dir(thumbnailsDir))` with `http.StripPrefix`, and `/originals/` to `http.FileServer(http.Dir(originalsDir))` with `http.StripPrefix`. Set as `AssetServer.Handler` in `wails.Run` options. Resolve media paths from `os.UserHomeDir()`.
- [x] T006 Add ABOUTME comments to `imaging.go`

**Checkpoint**: `ProcessImage("/path/to/photo.jpg")` generates original + thumbnail. Thumbnails served at `/thumbnails/{filename}` in the webview. Originals served at `/originals/{filename}`.

---

## Phase 3: User Story 1 - Attach Images to a Collection Item (Priority: P1)

**Goal**: Collector attaches images to an item via file picker. Images are processed and persisted with the item.

**Independent Test**: Create item, attach two images, save. Verify originals and thumbnails exist on disk. Re-open item and confirm image filenames persisted.

### Implementation for User Story 1

- [x] T007 [US1] Create `frontend/src/components/ImageAttach.vue`: accepts `images: string[]` prop (v-model pattern via `update:images` emit). Renders "Add Image" button that opens Wails native file dialog (via `wailsjs/runtime` OpenFileDialog). For each selected file, calls `ProcessImage(path)` binding, appends returned filename to images array. Displays thumbnail previews of attached images using `/thumbnails/{filename}` src. Allows removing images (remove from array). Shows error messages for rejected files.
- [x] T008 [US1] Integrate `ImageAttach.vue` into `frontend/src/components/DynamicForm.vue`: add an image attachment section below the attribute fields. Bind to the item's `images` array. Include attached image filenames in the saved Item payload.
- [x] T009 [US1] Run `wails build` and verify ProcessImage binding is generated in `frontend/wailsjs/go/main/App.d.ts`. Test attach flow: create item, attach image, save, verify files exist in media directories.

**Checkpoint**: Collector can attach images via file picker. Originals and thumbnails generated. Image filenames saved with item.

---

## Phase 4: User Story 2 - Collection Grid with Thumbnails (Priority: P2)

**Goal**: Collector browses items in a visual grid showing lazy-loaded thumbnails.

**Independent Test**: Save items with images. Switch to grid view. Verify thumbnails display, lazy loading works, no originals loaded.

### Implementation for User Story 2

- [x] T010 [US2] Create `frontend/src/components/CollectionGrid.vue`: accepts `items: Item[]` and `modules: ModuleSchema[]` props. Renders items as cards in a CSS grid layout. Each card shows the first image thumbnail via `<img :src="'/thumbnails/' + item.images[0]" loading="lazy" />` or a placeholder SVG/icon if no images. Shows item title below thumbnail. Emits `select(item)` on card click. Emits `viewImage(item, filename)` on thumbnail click.
- [x] T011 [US2] Add grid/list view toggle to `frontend/src/App.vue`: add a toggle button that switches between ItemList (existing) and CollectionGrid. Persist the view preference in a reactive ref. Wire CollectionGrid events: `select` for editing, `viewImage` for lightbox (wired in US3), `filterChange`/`search` shared with ItemList via the collection store.
- [x] T012 [US2] Add placeholder image for items without images: create a simple SVG placeholder inline in `CollectionGrid.vue` (no external file needed). Also handle missing thumbnail files gracefully with an `@error` handler on the `<img>` tag that swaps to the placeholder.

**Checkpoint**: Grid view displays items with thumbnails. Lazy loading confirmed by scrolling. View toggle switches between list and grid. No originals loaded in grid (verify via dev tools network tab).

---

## Phase 5: User Story 3 - Full-Resolution Image Viewing (Priority: P3)

**Goal**: Collector clicks a thumbnail to view the full-resolution original in a lightbox.

**Independent Test**: Save item with large image. View in grid. Click thumbnail. Verify full-res loads in overlay.

### Implementation for User Story 3

- [x] T013 [US3] Create `frontend/src/components/ImageLightbox.vue`: accepts `filename: string` and `visible: boolean` props. When visible, renders a full-screen overlay with the original image at `/originals/{filename}`. Close button and click-outside-to-close behavior. Emits `close()`. Only loads the `<img>` when `visible` is true (v-if, not v-show) to ensure on-demand loading.
- [x] T014 [US3] Wire lightbox into `frontend/src/App.vue`: when CollectionGrid emits `viewImage`, set lightbox filename and show it. On lightbox close, clear state. Also wire lightbox from DynamicForm when clicking attached image thumbnails.

**Checkpoint**: Clicking a thumbnail opens full-res in lightbox. Closing unloads the image. Grid remains responsive after closing.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Styling, cleanup, and end-to-end validation.

- [x] T015 [P] Add grid layout CSS to `CollectionGrid.vue`: responsive grid (auto-fill, minmax 200px), card styling, thumbnail aspect-ratio container, title truncation, hover effects
- [x] T016 [P] Add lightbox CSS to `ImageLightbox.vue`: overlay background, centered image, max-width/max-height constraints, close button positioning, fade transition
- [x] T017 Run `quickstart.md` validation: follow all verification steps (attach, grid, full-res, error handling, multiple images)
- [x] T018 Run `wails build` and verify production binary serves thumbnails and originals correctly from local media directory

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Setup -- BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (T003-T006)
- **User Story 2 (Phase 4)**: Depends on US1 (needs items with images to display)
- **User Story 3 (Phase 5)**: Depends on US2 (lightbox triggered from grid)
- **Polish (Phase 6)**: Depends on all user stories complete

### User Story Dependencies

- **User Story 1 (P1)**: Requires Phase 2. Independent of other stories.
- **User Story 2 (P2)**: Requires US1 (items with images to display in grid).
- **User Story 3 (P3)**: Requires US2 (grid emits viewImage event).

### Within Each User Story

- Backend before frontend
- Components before App.vue integration
- Story complete before moving to next priority

### Parallel Opportunities

- T015 and T016 can run in parallel (different files)
- T001 and T002 can run in parallel (different files)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T006)
3. Complete Phase 3: User Story 1 (T007-T009)
4. **STOP and VALIDATE**: Attach image, verify thumbnail generated
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational -> Processing pipeline ready
2. Add User Story 1 -> Image attachment works -> Demo (MVP!)
3. Add User Story 2 -> Grid with thumbnails -> Demo
4. Add User Story 3 -> Full-res lightbox -> Demo
5. Polish -> Production-ready

### Recommended Execution Order

```
Phase 1 (Setup: deps + types)
  |
Phase 2 (Foundational: imaging.go + AssetServer)
  |
Phase 3 (US1: ImageAttach + DynamicForm integration)
  |
Phase 4 (US2: CollectionGrid + view toggle)
  |
Phase 5 (US3: ImageLightbox)
  |
Phase 6 (Polish: CSS + validation)
```

All three user stories are sequential: US2 needs images from US1,
US3 needs the grid from US2.

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently testable at its checkpoint
- Commit after each task or logical group
- Constitution Principle IV is enforced by architecture: grid uses
  `/thumbnails/` paths, full-res uses `/originals/` only on demand
