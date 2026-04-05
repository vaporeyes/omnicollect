# Feature Specification: Dynamic Form Engine

**Feature Branch**: `002-dynamic-form-engine`
**Created**: 2026-04-05
**Status**: Draft
**Input**: User description: "Iteration 2: Build the Vue 3 frontend with dynamic form rendering based on module schemas from Iteration 1."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Add a New Collection Item (Priority: P1)

A collector selects a collection type (e.g., "Coins") from a list of
available modules. The application renders a form with the standard
fields (title, purchase price) plus all custom attribute fields defined
by that module's schema. The collector fills in the form and saves. The
item is persisted and appears in the collection list.

**Why this priority**: This is the primary user interaction -- without
the ability to add items through a dynamically generated form, the
schema-driven UI principle has no value.

**Independent Test**: Select a module, fill in all fields, save. Verify
the item appears in the collection list with correct title and that
re-opening it shows all saved attribute values.

**Acceptance Scenarios**:

1. **Given** the application has loaded module schemas from the backend,
   **When** the collector selects a module type, **Then** a form renders
   with the base fields (title, purchase price) and all custom attribute
   fields defined in that module's schema.
2. **Given** a module schema defines attributes with types "string",
   "number", "boolean", "date", and "enum", **When** the form renders,
   **Then** each attribute is rendered as the appropriate input control
   (text input, number input, checkbox, date picker, dropdown).
3. **Given** a module schema defines an attribute with display hints
   (label, placeholder, widget), **When** the form renders, **Then**
   the hints are applied to the corresponding input control.
4. **Given** a module schema marks certain attributes as required,
   **When** the collector attempts to save without filling required
   fields, **Then** validation errors are shown and the save is
   blocked.
5. **Given** the collector has filled in all required fields, **When**
   they save the form, **Then** the item is persisted with custom
   attributes nested inside an `attributes` object in the payload.

---

### User Story 2 - Browse and Search Collection Items (Priority: P2)

A collector views a list of their saved items, optionally filtered by
collection type or searched by keyword. The list shows item titles and
key metadata. Switching between collection types updates the list.

**Why this priority**: Viewing saved items is essential to validate
that the form engine works end-to-end and that data round-trips
correctly through the backend.

**Independent Test**: Save several items across two module types.
Browse all items, filter by one module, search by keyword. Verify
correct items appear in each case.

**Acceptance Scenarios**:

1. **Given** items exist in the collection, **When** the collector
   opens the application, **Then** a list of items is displayed
   showing titles and the collection type they belong to.
2. **Given** items exist across multiple collection types, **When**
   the collector selects a specific module filter, **Then** only
   items of that type are shown.
3. **Given** items exist with various titles and attribute values,
   **When** the collector enters a search term, **Then** matching
   items are returned (searching across titles and attribute values).
4. **Given** no items match the current filter or search, **When**
   the list renders, **Then** an empty state message is displayed
   (not a blank screen).

---

### User Story 3 - Edit an Existing Item (Priority: P3)

A collector selects an item from the list to edit it. The form
pre-populates with the item's current values (including all custom
attributes). The collector makes changes and saves. The updated values
are persisted.

**Why this priority**: Editing is important but builds directly on
the form engine from US1 -- the same dynamic form is reused with
pre-populated data.

**Independent Test**: Save an item, open it for editing, change a
field, save. Verify the change persists and appears correctly in
the list and when re-opened.

**Acceptance Scenarios**:

1. **Given** an item exists in the collection, **When** the collector
   selects it for editing, **Then** the dynamic form renders with all
   fields pre-populated with the item's current values.
2. **Given** the collector changes one or more fields on an existing
   item, **When** they save, **Then** the updated values are persisted
   and the list reflects the changes.
3. **Given** the collector opens an item for editing, **When** they
   cancel without saving, **Then** no changes are persisted.

---

### Edge Cases

- What happens when a module schema has zero custom attributes? The
  form MUST still render with the base fields (title, purchase price)
  and be saveable.
- What happens when a module schema is removed after items were saved
  under it? Existing items MUST still appear in the list and be
  viewable, though the dynamic form cannot render custom fields
  without the schema.
- What happens when an attribute type is unrecognized? The form MUST
  fall back to rendering a text input and log a warning.
- What happens when the backend returns an error on save? The form
  MUST display the error message to the collector and not clear the
  form data.
- What happens when the module list is empty (no schemas loaded)? The
  application MUST display a message directing the user to add module
  schemas, not an empty or broken UI.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The application MUST fetch and cache available module
  schemas from the backend on startup.
- **FR-002**: The application MUST fetch and cache collection items
  from the backend, refreshing after save operations.
- **FR-003**: The application MUST render a dynamic form based on
  the selected module's schema, mapping attribute types to
  appropriate input controls.
- **FR-004**: The type-to-input mapping MUST include at minimum:
  "string" to text input, "number" to number input, "boolean" to
  checkbox, "date" to date picker, "enum" to dropdown (populated
  from the attribute's options list).
- **FR-005**: The form MUST apply display hints (label, placeholder,
  widget override) from the schema when present.
- **FR-006**: The form MUST enforce required field validation before
  allowing submission.
- **FR-007**: On submission, the form MUST construct a payload where
  base fields (title, purchase price) are top-level and custom
  attribute values are nested in an `attributes` object.
- **FR-008**: The form MUST support both create (new item) and edit
  (existing item with pre-populated values) modes.
- **FR-009**: The application MUST display a list of saved items
  with title, collection type, and last-updated timestamp.
- **FR-010**: The application MUST support filtering the item list
  by collection type and searching by keyword.
- **FR-011**: The application MUST show user-friendly error messages
  when backend operations fail, without losing form state.

### Key Entities

- **Module Store**: Cached collection of module schemas fetched from
  the backend. Provides the schema definitions that drive form
  rendering.
- **Collection Store**: Cached collection of items fetched from the
  backend. Provides the data displayed in the item list and used to
  pre-populate edit forms.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A collector can add a new item to any collection type
  in under 60 seconds, including selecting the type and filling all
  required fields.
- **SC-002**: The form correctly renders all five attribute types
  (string, number, boolean, date, enum) with appropriate input
  controls for each.
- **SC-003**: Adding a new collection type requires only adding a
  new JSON schema file -- zero frontend code changes are needed.
- **SC-004**: Saved items display in the list within 1 second of
  saving, without requiring a page refresh.
- **SC-005**: Search results appear within 1 second of the
  collector finishing their query.
- **SC-006**: Form validation prevents submission of incomplete
  required fields 100% of the time, with clear error indicators.

## Assumptions

- This iteration depends on the Iteration 1 backend (Go + SQLite +
  Wails IPC bindings) being complete and functional.
- The generated TypeScript bindings from Wails provide typed access
  to `SaveItem`, `GetItems`, and `GetActiveModules`.
- No authentication or multi-user support is needed.
- Styling is minimal/functional in this iteration -- visual polish
  is deferred to a future iteration.
- Image management (upload, display, thumbnails) is out of scope
  for this iteration. The `images` field is ignored in the UI.
- The item list is a simple flat list (no pagination, infinite
  scroll, or virtualization) sufficient for collections up to a
  few hundred items. Performance optimization is deferred.
