# Research: Schema Visual Builder

**Date**: 2026-04-05
**Feature**: 004-schema-visual-builder

## R1: Code Editor Component

**Decision**: Use `vue-codemirror` (CodeMirror 6 wrapper) with
`@codemirror/lang-json` for JSON syntax highlighting.

**Rationale**: CodeMirror 6 provides real editor behavior (undo/redo,
bracket matching, proper keyboard shortcuts, line numbers) out of the
box. Bundle size (~80-120 kB gzipped) is acceptable for a Wails desktop
app where assets load from local filesystem. Clean v-model integration
with Vue 3. ESM-native, works with Vite. Actively maintained.

**Usage pattern**:
```typescript
import { Codemirror } from 'vue-codemirror'
import { json } from '@codemirror/lang-json'
const extensions = [json()]
// <Codemirror v-model="jsonContent" :extensions="extensions" />
```

**Alternatives considered**:
- `vue-prism-editor`: Lighter (~12 kB) but textarea-based -- no undo
  history, no bracket matching, no code folding. Rejected for poor
  editing experience.
- Monaco (`@guolao/vue-monaco-editor`): 2-4 MB bundle. Far too heavy.
  Rejected.
- Raw textarea + Prism overlay: Requires building scroll sync, line
  numbers, cursor positioning. Rejected -- reinvents what
  vue-prism-editor already does.

## R2: Bidirectional State Sync Strategy

**Decision**: Single reactive `draftSchema` object as source of truth.
Visual builder mutates the object directly. Code editor receives a
computed JSON string and emits text changes. A `watch` on the code
editor text attempts `JSON.parse` -- on success, updates the draft
object; on failure, sets an error string and leaves the object unchanged.

**Rationale**: Making the structured object the primary state avoids
the complexity of merging two representations. The code editor is a
"derived view" that can temporarily diverge during typing errors. The
visual builder always reflects the last valid state.

**Error recovery**: The code editor maintains its own text even when
parsing fails (the user's cursor position and partial edits are not
lost). Only the visual preview freezes at the last valid state. When
parsing succeeds again, the object updates and the preview resumes.

## R3: Slug Generation for Schema ID

**Decision**: Auto-generate ID from displayName using:
lowercase, replace non-alphanumeric with hyphens, collapse consecutive
hyphens, trim leading/trailing hyphens.

**Rationale**: Users should not have to think about machine-readable
IDs. The slug is visible in the code editor for power users to
override. Simple regex-based implementation.

**Example**: "Vinyl Records" -> "vinyl-records"

## R4: Schema Save and Hot Reload

**Decision**: The `SaveCustomModule` Go binding writes the file to
disk, then calls `loadModuleSchemas()` to reload all schemas into
the App's in-memory slice. The frontend then calls
`moduleStore.fetchModules()` to refresh.

**Rationale**: Reloading all schemas (not just the new one) is
simpler and handles edge cases like renamed/deleted files. The
modules directory typically has <50 files, so full reload is fast
(< 100ms per research from Iteration 1).

## R5: Field Reordering

**Decision**: Move up/down buttons on each field row. Drag-and-drop
deferred as a stretch goal.

**Rationale**: Move buttons are trivial to implement (array splice)
and accessible. Drag-and-drop requires a library (e.g., vuedraggable)
which adds dependency weight for marginal UX improvement. Can be added
later without architecture changes.

## Summary of New Dependencies

| Package | Purpose |
|---------|---------|
| `vue-codemirror` | CodeMirror 6 Vue 3 wrapper |
| `@codemirror/lang-json` | JSON syntax highlighting |
