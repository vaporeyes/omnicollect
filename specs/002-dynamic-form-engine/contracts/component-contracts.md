# Component Contracts: Dynamic Form Engine

**Date**: 2026-04-05
**Feature**: 002-dynamic-form-engine

## DynamicForm.vue

Renders a form from a module schema. Supports create and edit modes.

**Props**:
- `schema: ModuleSchema` (required) -- The module schema driving form
  generation.
- `item: Item | null` (optional, default null) -- When provided,
  pre-populates form for editing. When null, form is in create mode.

**Emits**:
- `save(item: Item)` -- Fired when the form is submitted with valid
  data. Payload is a fully constructed Item object ready for
  `SaveItem()`.
- `cancel()` -- Fired when the user cancels editing.

**Behavior**:
- Renders base fields (title, purchase price) at the top.
- Loops over `schema.attributes` and renders a `FormField` for each.
- Validates required fields on submit, showing errors inline.
- Constructs the `Item` payload with attributes nested in the
  `attributes` object.

## FormField.vue

Renders a single form field based on its schema attribute definition.

**Props**:
- `attribute: AttributeSchema` (required) -- The attribute definition.
- `modelValue: any` (required) -- Current field value (v-model).

**Emits**:
- `update:modelValue(value: any)` -- Standard v-model update.

**Behavior**:
- Dispatches on `attribute.type` to render the correct input control.
- Applies `attribute.display.label` as the field label (falls back to
  `attribute.name`).
- Applies `attribute.display.placeholder` if present.
- If `attribute.display.widget` is set, overrides the default control
  (e.g., "textarea" for a string field).
- For "enum" type, renders `<select>` with `attribute.options`.
- For unrecognized types, falls back to text input with console warning.

## ItemList.vue

Displays a list of collection items with filtering and search.

**Props**:
- `items: Item[]` (required) -- Items to display.
- `modules: ModuleSchema[]` (required) -- Available modules (for
  displaying collection type names).

**Emits**:
- `select(item: Item)` -- Fired when the user clicks an item to edit.
- `filterChange(moduleId: string)` -- Fired when module filter changes.
- `search(query: string)` -- Fired when search text changes.

**Behavior**:
- Renders each item as a row showing title, collection type name,
  and last-updated timestamp.
- Shows an empty state message when no items match.
- Provides a module filter dropdown and a search text input.

## ModuleSelector.vue

Allows the user to select a collection type for adding a new item.

**Props**:
- `modules: ModuleSchema[]` (required) -- Available modules.

**Emits**:
- `select(module: ModuleSchema)` -- Fired when a module is selected.

**Behavior**:
- Renders a list/dropdown of available modules showing displayName.
- When no modules are available, shows a message directing the user
  to add schema files.
