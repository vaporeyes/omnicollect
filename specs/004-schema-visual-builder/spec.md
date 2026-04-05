# Feature Specification: Schema Visual Builder

**Feature Branch**: `004-schema-visual-builder`
**Created**: 2026-04-05
**Status**: Draft
**Input**: User description: "Iteration 4: Split-pane visual builder for creating and editing module schema files."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create a New Schema Visually (Priority: P1)

A collector wants to define a new collection type (e.g., "Vinyl Records")
without writing JSON by hand. They open the schema builder, enter a
display name and description, then add fields using a visual interface:
clicking "Add Field", naming it, selecting a type, toggling required,
and adding enum options for dropdown fields. The left pane shows a live
preview of what the collection form will look like. They save, and the
new collection type becomes immediately available.

**Why this priority**: This is the primary value of the builder. Without
visual field creation, users must hand-edit JSON files, which is
error-prone and inaccessible to non-technical collectors.

**Independent Test**: Open the builder, enter a schema name, add three
fields (string, number, enum), save. Verify the schema file is written
to the modules directory and the new collection type appears in the
module selector.

**Acceptance Scenarios**:

1. **Given** the user opens the schema builder, **When** they enter a
   display name, add fields, and click save, **Then** a valid JSON
   schema file is written to the modules directory.
2. **Given** the user adds a field with type "enum", **When** they
   configure the field, **Then** they can add, reorder, and remove
   options from the enum's option list.
3. **Given** the user adds fields in the visual builder, **When** they
   view the left pane, **Then** they see a live preview of the
   resulting form matching the current field definitions.
4. **Given** the user has created a valid schema, **When** they save,
   **Then** the new collection type appears in the module selector
   without requiring an application restart.
5. **Given** the user toggles a field's "required" flag, **When** the
   preview updates, **Then** the field shows a required indicator.

---

### User Story 2 - Edit Schema via JSON Code Editor (Priority: P2)

A power user or developer wants to edit the schema definition directly
as JSON text. The right pane of the split-pane layout shows a code
editor with the current schema as formatted JSON. Edits in the code
editor update the visual preview in real-time. Syntax errors during
typing do not crash the preview -- the preview holds its last valid
state until the JSON becomes valid again.

**Why this priority**: The code editor provides full control over the
schema and serves as a learning tool (users see how visual changes
map to JSON). It is secondary to the visual builder because most
users will use the visual interface.

**Independent Test**: Open the builder, type valid JSON in the code
editor. Verify the visual preview updates. Introduce a syntax error.
Verify the preview does not crash and recovers when the error is fixed.

**Acceptance Scenarios**:

1. **Given** a schema is loaded in the builder, **When** the user
   edits the JSON in the code editor, **Then** the visual preview
   updates to reflect the changes.
2. **Given** the user is typing in the code editor and the JSON is
   temporarily invalid (mid-keystroke), **When** the parser encounters
   an error, **Then** the visual preview retains the last valid state
   and a non-intrusive error indicator appears.
3. **Given** the user fixes a JSON syntax error, **When** the JSON
   becomes valid again, **Then** the visual preview updates immediately.
4. **Given** the user makes changes in the code editor, **When** they
   switch focus to the visual builder, **Then** both views are in sync.

---

### User Story 3 - Edit an Existing Schema (Priority: P3)

A collector wants to modify an existing collection type -- adding a new
field, removing one, or changing display hints. They open the builder
with the existing schema loaded. Both the visual builder and code editor
show the current state. They make changes and save. Existing items
under this collection type remain intact (schema changes do not affect
stored data).

**Why this priority**: Editing builds on the create flow (US1) and
is important for schema evolution, but it is less common than initial
creation.

**Independent Test**: Load an existing schema in the builder. Add a
field, save. Verify the schema file is updated. Verify existing items
still display correctly (no data loss).

**Acceptance Scenarios**:

1. **Given** the user selects an existing schema to edit, **When** the
   builder opens, **Then** both the visual builder and code editor are
   populated with the current schema definition.
2. **Given** the user adds a field to an existing schema, **When** they
   save, **Then** the schema file on disk is updated and existing items
   under this schema are unaffected.
3. **Given** the user removes a field from an existing schema, **When**
   they save, **Then** existing items retain their stored attribute
   data for the removed field (data is preserved, just not displayed
   in forms).

---

### Edge Cases

- What happens when the user saves a schema with a duplicate ID
  (matching an existing schema)? The system MUST warn the user and
  prevent overwriting a different schema file.
- What happens when the user tries to save with an empty display name?
  The system MUST show a validation error and block the save.
- What happens when the user enters deeply nested or circular JSON in
  the code editor? The parser MUST handle this gracefully without
  freezing the UI.
- What happens when the user reorders fields in the visual builder?
  The field order MUST be preserved in the saved JSON.
- What happens when the modules directory is not writable? The system
  MUST display a clear error message.
- What happens when the user closes the builder with unsaved changes?
  The system MUST prompt for confirmation before discarding.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a split-pane interface with a
  visual form preview on the left and a JSON code editor on the right.
- **FR-002**: The visual builder MUST allow adding fields with a name,
  type (string, number, boolean, date, enum), required flag, and
  display hints (label, placeholder, widget).
- **FR-003**: For enum-type fields, the visual builder MUST allow
  adding, removing, and reordering option values.
- **FR-004**: The visual preview MUST show a live rendering of the
  form that would be generated from the current schema definition.
- **FR-005**: The code editor MUST display the current schema as
  formatted JSON and allow direct text editing.
- **FR-006**: Changes in either pane (visual or code) MUST be
  reflected in the other pane in real-time via shared reactive state.
- **FR-007**: JSON parse errors in the code editor MUST NOT crash the
  visual preview. The preview MUST hold its last valid state until
  valid JSON is restored.
- **FR-008**: The system MUST validate the schema before saving:
  non-empty ID, non-empty display name, unique attribute names,
  recognized attribute types.
- **FR-009**: On save, the system MUST write the schema as a formatted
  JSON file to the modules directory on disk.
- **FR-010**: After saving, the new or updated schema MUST be
  available in the module selector without requiring an application
  restart.
- **FR-011**: The builder MUST support loading an existing schema for
  editing, pre-populating both panes.
- **FR-012**: The system MUST prompt the user before discarding
  unsaved changes.
- **FR-013**: The visual builder MUST support reordering fields via
  drag-and-drop or move up/down controls.

### Key Entities

- **Draft Schema**: The in-progress schema being constructed or edited
  in the builder. Exists in memory as shared reactive state between
  both panes. Not persisted until the user explicitly saves.
- **Saved Schema**: A module schema file on disk in the modules
  directory. Created or updated when the user saves from the builder.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user with no JSON knowledge can create a new collection
  type with 5 fields in under 3 minutes using only the visual builder.
- **SC-002**: Changes in either pane (visual or code) are reflected in
  the other pane within 500 milliseconds.
- **SC-003**: JSON syntax errors during typing never cause the visual
  preview to crash or show a blank state.
- **SC-004**: After saving a new schema, the collection type appears in
  the module selector within 1 second without application restart.
- **SC-005**: Editing an existing schema and saving preserves all
  existing items under that collection type with zero data loss.
- **SC-006**: Schema validation catches 100% of invalid configurations
  (empty names, duplicate attributes, unrecognized types) before save.

## Assumptions

- Iterations 1-3 are complete and functional.
- The builder is accessed from a dedicated navigation entry (e.g., a
  "Schema Builder" button in the sidebar), not from within the item
  management flow.
- The schema ID is auto-generated from the display name (slugified)
  for new schemas. Users can override it in the code editor.
- Field reordering uses move up/down buttons for this iteration.
  Full drag-and-drop is a stretch goal -- if simple to implement,
  include it; otherwise defer.
- The code editor provides basic syntax highlighting and line numbers
  but does not need full IDE features (autocomplete, linting).
- The "live preview" shows the form as it would appear in the
  DynamicForm component (reusing the existing FormField rendering).
- Schema deletion is out of scope for this iteration.
