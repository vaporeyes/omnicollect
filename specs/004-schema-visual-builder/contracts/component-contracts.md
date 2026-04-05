# Component Contracts: Schema Visual Builder

**Date**: 2026-04-05
**Feature**: 004-schema-visual-builder

## SchemaBuilder.vue

Top-level split-pane layout for the schema builder.

**Props**:
- `moduleId: string | null` (optional) -- If provided, loads existing
  schema for editing. If null, starts with empty template.

**Emits**:
- `saved(schema: ModuleSchema)` -- Fired after successful save.
- `close()` -- Fired when user closes the builder.

**Behavior**:
- Manages the central draft schema reactive state.
- Renders SchemaVisualEditor on the left and SchemaCodeEditor on
  the right in a resizable split-pane.
- Shows save/cancel buttons in a toolbar.
- Validates before save, showing errors inline.
- Prompts on close if unsaved changes exist.

## SchemaVisualEditor.vue

Visual form builder for adding/editing fields.

**Props**:
- `schema: DraftSchema` (required, reactive) -- The draft schema.

**Emits**:
- `update:schema(schema: DraftSchema)` -- On any visual change.

**Behavior**:
- Editable fields for displayName and description at the top.
- "Add Field" button that appends a new attribute with defaults.
- Each field row shows: name input, type dropdown, required toggle,
  display hints (collapsible), remove button, move up/down buttons.
- For enum fields, shows an options editor (add/remove/reorder).
- Live form preview section at the bottom using FormField components.

## SchemaCodeEditor.vue

JSON text editor pane.

**Props**:
- `modelValue: string` (required) -- JSON string content.
- `error: string | null` (optional) -- Parse error message to display.

**Emits**:
- `update:modelValue(value: string)` -- On text change.

**Behavior**:
- Renders a code editor component with JSON syntax highlighting
  and line numbers.
- Emits text changes on input.
- Displays error indicator when `error` prop is set.
- Does not attempt to parse JSON itself (parent handles that).

## SchemaFormPreview.vue

Live preview of the form that would be generated from the draft schema.

**Props**:
- `schema: DraftSchema` (required) -- Current draft to preview.

**Behavior**:
- Renders FormField components for each attribute in the schema.
- Read-only (inputs are disabled or visual-only).
- Updates reactively as the schema changes.
- Shows "Add fields to see preview" when attributes array is empty.
