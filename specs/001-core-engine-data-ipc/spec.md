# Feature Specification: Core Engine (Data & IPC)

**Feature Branch**: `001-core-engine-data-ipc`
**Created**: 2026-04-05
**Status**: Draft
**Input**: User description: "Iteration 1: The Core Engine (Data & IPC) - Establish the Go backend, SQLite schema, and the Wails IPC bridge."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Save and Retrieve a Collection Item (Priority: P1)

A collector opens OmniCollect, selects a collection type (e.g., "Coins"),
fills in a form with a title, purchase price, and type-specific attributes,
and saves the item. The item persists locally and appears when the
collector returns to browse their collection.

**Why this priority**: Without the ability to create and retrieve items,
no other feature has value. This is the foundational CRUD path that
everything else depends on.

**Independent Test**: Can be fully tested by saving an item through the
backend API, then querying for it and verifying all fields are returned
correctly. Delivers the core value of persistent local data storage.

**Acceptance Scenarios**:

1. **Given** OmniCollect is running with at least one module schema
   installed, **When** a user saves a new item with title, purchase
   price, and custom attributes, **Then** the item is persisted in the
   local database and can be retrieved by its ID with all fields intact.
2. **Given** multiple items exist across different collection types,
   **When** a user requests items filtered by a specific module ID,
   **Then** only items belonging to that collection type are returned.
3. **Given** an item was previously saved, **When** the user updates
   it with changed attributes and saves again, **Then** the stored
   item reflects the updated values.

---

### User Story 2 - Search Across Collections (Priority: P2)

A collector wants to find a specific item they added weeks ago. They type
a search term into a query field. The system searches across item titles,
base fields, and custom attributes to return matching results.

**Why this priority**: Once a collection grows, browsing becomes
impractical. Full-text search is essential for usability at scale, and
the FTS5 index must be established as part of the core schema.

**Independent Test**: Can be tested by inserting several items with
known attributes, then querying by keyword and verifying the correct
subset is returned. Delivers the value of fast, cross-field search.

**Acceptance Scenarios**:

1. **Given** items exist with various titles and attribute values,
   **When** a user searches with a term that appears in an item's
   title, **Then** that item is included in the results.
2. **Given** items exist with matching terms in custom attributes
   (not title), **When** a user searches for that term, **Then**
   those items are still found via full-text search.
3. **Given** a search query that matches no items, **When** the user
   performs the search, **Then** an empty result set is returned
   without errors.

---

### User Story 3 - Discover Available Collection Types (Priority: P3)

A collector opens OmniCollect for the first time (or after adding new
module schemas to their modules directory). The application scans for
available collection types and presents them so the collector can choose
which type of item to add.

**Why this priority**: Module discovery is necessary for the schema-driven
UI principle, but it is lower priority because a hardcoded test schema
can unblock US1 and US2 during early development.

**Independent Test**: Can be tested by placing JSON schema files in the
modules directory, starting the application, and verifying all schemas
are detected and returned with correct metadata. Delivers the value of
extensible, user-managed collection types.

**Acceptance Scenarios**:

1. **Given** one or more valid JSON schema files exist in the modules
   directory, **When** the application starts, **Then** all valid
   schemas are parsed and available via the active modules query.
2. **Given** a malformed JSON schema file exists alongside valid ones,
   **When** the application scans modules, **Then** valid schemas
   load successfully and the malformed file is reported as an error
   without crashing the application.
3. **Given** no schema files exist in the modules directory, **When**
   the application starts, **Then** an empty list of modules is
   returned and the application remains functional.

---

### Edge Cases

- What happens when two schema files define the same module ID?
  The system MUST reject duplicates and report the conflict at startup.
- What happens when the modules directory does not exist on first run?
  The system MUST create it automatically.
- What happens when an item is saved with attributes not present in
  its module schema? The system MUST store the data without loss
  (schema is advisory for UI generation, not a validation gate for
  storage).
- What happens when the SQLite database file is missing or corrupted?
  The system MUST create a fresh database on first run. Corruption
  recovery is out of scope for this iteration.
- What happens when the `attributes` JSON blob exceeds typical sizes?
  The system MUST handle attribute blobs up to at least 1 MB without
  performance degradation on save/retrieve.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST initialize a local database on first run
  with a single `items` table containing: `id` (UUID primary key),
  `module_id` (text), `title` (text), `purchase_price` (numeric),
  `images` (JSON array), and `attributes` (JSON text).
- **FR-002**: System MUST provide a full-text search index covering
  the `title` field and the contents of the `attributes` JSON blob.
- **FR-003**: System MUST expose a `SaveItem` operation that creates
  or updates an item in the local database.
- **FR-004**: System MUST expose a `GetItems` operation that accepts
  an optional search query and an optional module ID filter, returning
  matching items.
- **FR-005**: System MUST scan a designated modules directory at
  startup and parse each JSON file into a structured module schema
  representation.
- **FR-006**: System MUST expose a `GetActiveModules` operation that
  returns all successfully parsed module schemas.
- **FR-007**: System MUST create the modules directory if it does not
  exist on first run.
- **FR-008**: All three operations (`SaveItem`, `GetItems`,
  `GetActiveModules`) MUST be accessible from the frontend via
  type-safe generated bindings.
- **FR-009**: System MUST operate entirely offline; no network calls
  are permitted during any of these operations.
- **FR-010**: System MUST use a CGO-free SQLite driver to simplify
  cross-platform builds.

### Key Entities

- **Item**: A single collectible record. Has a unique ID, belongs to
  one module (collection type), carries base fields (title, purchase
  price, images) plus a freeform attributes blob driven by the module
  schema.
- **ModuleSchema**: A definition of a collection type. Loaded from a
  JSON file at startup. Describes the custom attributes, their types,
  and display hints that drive the schema-driven UI.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can save a new collection item and retrieve it
  by ID with all fields intact, completing the round-trip in under
  1 second on consumer hardware.
- **SC-002**: Full-text search returns relevant results from both
  base fields and custom attributes within 500 milliseconds for
  collections of up to 10,000 items.
- **SC-003**: The application starts and loads all module schemas
  from the modules directory in under 2 seconds with up to 50
  schema files present.
- **SC-004**: All backend operations are callable from the frontend
  through generated typed bindings with zero raw IPC calls.
- **SC-005**: The system operates with full functionality and zero
  network requests during all CRUD and search operations.

## Assumptions

- This iteration focuses on the backend data layer and IPC bindings
  only. No frontend UI is built in this iteration.
- The modules directory location is `~/.omnicollect/modules/` and is
  not user-configurable in this iteration.
- JSON Schema is used as the module schema format, consistent with
  Constitution Principle II.
- Image storage and thumbnail generation (Constitution Principle IV)
  are out of scope for this iteration; the `images` field stores
  references (paths) only.
- The `purchase_price` field stores values as numeric without currency
  tracking; currency support is deferred to a future iteration.
- `SaveItem` handles both create (new UUID) and update (existing UUID)
  in a single operation (upsert semantics).
