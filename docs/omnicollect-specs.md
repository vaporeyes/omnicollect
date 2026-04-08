# OmniCollect: Spec-Kit

## Project Constitution

These are the immutable engineering principles that govern the OmniCollect codebase. Any PR or feature addition that violates these rules must be refactored.

1. **Local-First Mandate:** The primary source of truth is the local SQLite database. The application must have 100% functionality without an active internet connection. No required centralized cloud accounts.
2. **Schema-Driven UI:** UI forms for collection items are strictly generated at runtime from JSON schemas. There will be no hardcoded Vue templates for specific item types (e.g., `BookForm.vue` or `CoinForm.vue` are forbidden).
3. **Flat Data Architecture:** All item metadata will be stored as flat JSON documents within a single SQLite table using an EAV (Entity-Attribute-Value) pattern. Avoid complex SQL `JOIN` operations for item attributes.
4. **Performance & Memory Protection:** The frontend must never load original, high-resolution media files into list/grid views. All list renders must use the generated, highly compressed local thumbnails.
5. **Type-Safe IPC:** All communication between the Vue 3 frontend and Go backend must utilize Wails-generated TypeScript bindings. Raw untyped IPC calls are prohibited.

---

## Iterative Specifications

### Iteration 1: The Core Engine (Data & IPC)

**Goal:** Establish the Go backend, SQLite schema, and the Wails IPC bridge.

* **1.1 Database Initialization:**
  * Implement CGO-free `modernc.org/sqlite`.
  * Create the `items` table with `id` (UUID), `module_id` (TEXT), base fields (`title`, `purchase_price`), `images` (JSON array), and `attributes` (JSON text).
  * Implement SQLite FTS5 (Full-Text Search) indexing on base fields and the `attributes` JSON blob.
* **1.2 Module Schema Manager:**
  * Create Go routines to scan `~/.omnicollect/modules/` on boot.
  * Parse `.json` files into a strongly typed Go struct (`ModuleSchema`).
* **1.3 Wails Bindings (CRUD):**
  * Expose `SaveItem`, `GetItems(query string, moduleId string)`, and `GetActiveModules()` to the frontend.

### Iteration 2: Dynamic Rendering (Vue 3 Bridge)

**Goal:** Build the Vue 3 frontend that dynamically renders forms based on the payloads delivered by Iteration 1.

* **2.1 State Management:**
  * Initialize Pinia stores: `useModuleStore` (caches schemas) and `useCollectionStore` (caches loaded items).
* **2.2 Dynamic Form Engine:**
  * Create `DynamicForm.vue`.
  * Implement a loop parsing the `fields` array from a module schema.
  * Map JSON types to HTML inputs: `string` -> `<input type="text">`, `enum` -> `<select>`, `integer` -> `<input type="number">`.
* **2.3 Data Consolidation:**
  * Ensure the form submission constructs a strict JSON payload where custom fields are nested within an `attributes` object, ready for the Go backend.

### Iteration 3: The Media Pipeline

**Goal:** Handle high-resolution image processing and local secure rendering.

* **3.1 Go Image Processor:**
  * Implement `github.com/disintegration/imaging`.
  * Create a Go method `ProcessImage(path string)` that saves a JPEG copy to `~/.omnicollect/media/originals/` and a compressed 300x300 crop to `~/.omnicollect/media/thumbnails/`.
* **3.2 Wails Asset Server:**
  * Configure Wails `AssetServer` with a `NewLocalFileHandler` bound to the media directory to bypass browser security protocols.
* **3.3 Vue Lazy Rendering:**
  * Implement a standard `<CollectionGrid>` component.
  * Use `http://wails.localhost/thumbnails/{filename}` for the image `src`.
  * Ensure native browser `loading="lazy"` is applied to all grid images.

### Iteration 4: The Module Builder GUI

**Goal:** Implement the split-pane visual builder to allow users to create and edit their own schema files.

* **4.1 Split-Pane Layout:**
  * Build the primary view separating the Visual Canvas (Left) and the Code Editor (Right).
* **4.2 Reactive State Engine:**
  * Establish a deep-watched Vue `ref` holding the drafted schema.
* **4.3 JSON Text Editor:**
  * Integrate a lightweight code editor component (e.g., `vue-codemirror`).
  * Bind the editor string to the reactive Vue state, wrapping the parser in a `try/catch` to gracefully handle mid-typing syntax errors without crashing the visual canvas.
* **4.4 Visual Drag-and-Drop:**
  * Implement a form builder that mutates the reactive state when users add fields, toggle "required" flags, or add enum options.
* **4.5 Disk Sync:**
  * Expose a `SaveCustomModule(schema JSON)` binding in Go that marshals the schema and writes it to the local `~/.omnicollect/modules/` directory.

### Iteration 5: "Bring Your Own Sync" (Future Proofing)

**Goal:** Prepare the architecture for user-hosted network synchronization.

* **5.1 Timestamp Tracking:** Ensure the Go backend strictly manages an `updated_at` UTC timestamp for every item modification.
* **5.2 Export Utilities:** Build Go methods to generate a `.zip` archive containing the SQLite DB and the media folder for manual backups.
* *(Spec-Kit Deferred)*: The standalone containerized sync server (Docker/k3s) will be detailed in a future spec-kit definition once the local client reaches v1.0.

---

### **Critical Bugs & Edge Cases**

**1. SQLite FTS5 Syntax Panics**
In `db.go`, the `queryItems` function passes the raw user search text directly into the FTS5 `MATCH` clause: 
`WHERE items_fts MATCH ?`
FTS5 has a strict query syntax. If a user types an unclosed quote (e.g., `12" vinyl`) or a reserved keyword, SQLite will throw a syntax error, causing the search to fail entirely. 
* **The Fix:** Sanitize the input before querying. Wrap the user's string in double quotes and append a wildcard for robust partial matching: 
`safeQuery := "\"" + strings.ReplaceAll(query, "\"", "\"\"") + "\"*"`

**2. Filename URL Encoding**
In `CollectionGrid.vue` and `ImageLightbox.vue`, images are bound directly to the src:
`:src="'/thumbnails/' + item.images[0]"`
While the Go backend generates UUID-based filenames for standard uploads (which are URL-safe), if you ever introduce a feature to import existing folders where original filenames are preserved, spaces or special characters in the filename will break the image rendering.
* **The Fix:** Wrap the filename in `encodeURIComponent(item.images[0])`.

**3. Unbounded Image Processing**
In `imaging.go`, `validateImage` decodes the config to check dimensions, but `generateThumbnail` loads the entire image into memory using `imaging.Open`. If a user uploads a massive 50MB TIFF scan, this could cause a significant memory spike.
* **The Fix:** Implement a file size check before attempting to process the image, or utilize an imaging library that supports stream-based downsampling.

---

### **UI & UX Enhancements**

**Advanced Media Inspection (The Loupe)**
When examining the patina and mint marks on ancient Roman currency, or verifying the condition and grading notes on comic books, a standard lightbox is often insufficient.
* **Enhancement:** Implement a "Hover Loupe" or deep-zoom feature in the `ImageLightbox.vue` component. Libraries like OpenSeadragon or a custom Vue directive tracking mouse coordinates can allow users to pan around the original high-resolution image natively without repeatedly clicking to zoom.

**True Data Table for List View**
The current `ItemList.vue` uses a stacked layout (`<ul class="items">`). Managing a vast collection requires high-density data visualization.
* **Enhancement:** Upgrade the list view to a dynamic data table. Since the active `ModuleSchema` defines the exact data types (currency, dates, strings), you can automatically generate sortable column headers for the specific module being viewed. This allows users to sort their inventory by `purchase_price`, `mint_year`, or `condition`.

**Drag-and-Drop Schema Builder**
The `SchemaVisualEditor.vue` currently relies on `^` and `v` buttons to reorder fields.
* **Enhancement:** Integrate a lightweight library like `vuedraggable` (built on Sortable.js). Allowing users to physically drag fields up and down the canvas creates a much more tactile and satisfying module-building experience.

**First-Class Theming & Dark Mode**
Desktop applications are heavily expected to respect system themes. The current `style.css` uses hardcoded hex values (e.g., `#3182ce`, `#f7fafc`). 
* **Enhancement:** Extract all colors into CSS variables (`--bg-primary`, `--text-main`, `--accent-blue`). You can hook into the Wails runtime via `WindowSetSystemDefaultTheme()` to automatically toggle a `.dark-theme` class on the Vue root, immediately modernizing the application's feel.

**Actionable Empty States**
The "No items found" and "No collection types available" messages are currently passive text blocks.
* **Enhancement:** Replace these with SVGs and primary call-to-action buttons. When the collection is empty, render a large button that automatically opens the "New Schema" builder to eliminate friction for first-time onboarding.
