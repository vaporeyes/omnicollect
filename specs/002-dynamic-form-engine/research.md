# Research: Dynamic Form Engine

**Date**: 2026-04-05
**Feature**: 002-dynamic-form-engine

## R1: State Management with Pinia

**Decision**: Use Pinia for Vue 3 state management.

**Rationale**: Pinia is the official Vue 3 state management library,
replacing Vuex. It integrates natively with Vue 3's Composition API,
provides TypeScript support out of the box, and has a simple
store-per-concern pattern that maps directly to our two data domains
(modules and items).

**Alternatives considered**:
- Vuex 4: Legacy API, more boilerplate. Pinia is the recommended
  replacement. Rejected.
- Composables only (no store library): Would work for this scale but
  lacks devtools integration and standardized patterns. Rejected for
  consistency.

## R2: Dynamic Form Rendering Pattern

**Decision**: A single `DynamicForm.vue` component that receives a
`ModuleSchema` prop and loops over its `attributes` array to render
form fields. A child `FormField.vue` component handles type-to-input
dispatch.

**Rationale**: Separating field rendering into `FormField.vue` keeps
`DynamicForm.vue` focused on layout and submission logic. The type
dispatch is a simple switch on `attribute.type`:
- `"string"` -> `<input type="text">`
- `"number"` -> `<input type="number">`
- `"boolean"` -> `<input type="checkbox">`
- `"date"` -> `<input type="date">`
- `"enum"` -> `<select>` with `<option>` per `attribute.options`

Display hints override defaults: `widget` can force a different
control (e.g., "textarea" for a string field), `label` overrides the
attribute name, `placeholder` sets input placeholder text.

**Alternatives considered**:
- Render function approach (programmatic VNodes): More flexible but
  harder to read and maintain. Template-based approach is sufficient
  for the five supported types. Rejected.
- Third-party form library (FormKit, VeeValidate): Adds dependency
  weight for a simple use case. Our schema format is custom, so a
  library adapter would be needed anyway. Rejected.

## R3: Payload Construction

**Decision**: The form maintains two reactive objects: `baseFields`
(title, purchasePrice) and `attributes` (keyed by attribute name).
On submit, these are merged into an `Item` object matching the Wails
binding type.

**Rationale**: Keeping base fields separate from custom attributes
mirrors the backend data model (base columns + JSON attributes blob).
The merge step is trivial:
```typescript
const item: Item = {
  id: editingItem?.id ?? "",
  moduleId: selectedModule.id,
  title: baseFields.title,
  purchasePrice: baseFields.purchasePrice ?? undefined,
  images: editingItem?.images ?? [],
  attributes: { ...attributes },
  createdAt: editingItem?.createdAt ?? "",
  updatedAt: "",
}
```

## R4: Wails Binding Integration

**Decision**: Import directly from the generated
`wailsjs/go/main/App` module. No wrapper layer.

**Rationale**: The Wails-generated bindings are already typed and
return Promises. Adding a service wrapper would be an unnecessary
abstraction for three functions. The Pinia stores call the bindings
directly.

**Import pattern**:
```typescript
import { SaveItem, GetItems, GetActiveModules } from '../../wailsjs/go/main/App'
```

## Summary of Dependencies

| Package | Purpose |
|---------|---------|
| `pinia` | Vue 3 state management |

No other new dependencies needed. Vue 3, TypeScript, and Vite are
already present from the Wails template.
