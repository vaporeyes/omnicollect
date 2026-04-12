# Data Model: Masonry Grid & Item Comparison

**Branch**: `020-masonry-compare` | **Date**: 2026-04-11

## Overview

No new database entities, tables, or backend changes. This feature is entirely frontend. This document describes the client-side data flow for the comparison view.

## Existing Entities Used

### Item (from collectionStore)

```
Item {
  id: string
  moduleId: string
  title: string
  purchasePrice: number | null
  images: string[]
  tags: string[]
  attributes: Record<string, any>
  createdAt: string
  updatedAt: string
}
```

### ModuleSchema (from moduleStore)

```
ModuleSchema {
  id: string
  displayName: string
  description: string
  attributes: AttributeSchema[]
}

AttributeSchema {
  name: string
  type: string          // "string" | "number" | "boolean" | "enum"
  required: boolean
  options: string[]     // For enum type
  display: DisplayHints | null
}
```

## Comparison View Data Flow

### Input

The comparison view receives two items by ID from the selection store. It resolves the full Item objects from the collection store's loaded items array.

```
selectionStore.selectedIdArray() -> [idA, idB]
collectionStore.items.find(id === idA) -> itemA
collectionStore.items.find(id === idB) -> itemB
moduleStore.modules.find(id === itemA.moduleId) -> schemaA
moduleStore.modules.find(id === itemB.moduleId) -> schemaB
```

### Diff Computation

The diff table rows are computed from the union of:

1. **Core fields** (always present):
   - Title (string comparison)
   - Purchase Price (number comparison, null-aware)
   - Tags (sorted array comparison)

2. **Schema attributes** (union of both schemas):
   - If same module: iterate schema attributes, compare values
   - If different modules: merge both attribute lists, mark missing fields as empty

Each row produces:

```
DiffRow {
  label: string          // Display label from schema or hardcoded for core fields
  valueA: string | null  // Formatted value for item A
  valueB: string | null  // Formatted value for item B
  isDifferent: boolean   // Whether the two values differ
}
```

### Gallery Synchronization

```
activeImageIndex: number  // Shared between both sides
maxIndexA: itemA.images.length - 1
maxIndexB: itemB.images.length - 1

effectiveIndexA: min(activeImageIndex, maxIndexA)
effectiveIndexB: min(activeImageIndex, maxIndexB)
```

Navigation advances `activeImageIndex`. Each side clamps to its own max.

## State Management

No new Pinia store needed. The comparison view state is local to the component:

- `activeImageIndex: ref(0)` -- Shared gallery position
- `diffRows: computed(...)` -- Derived from the two items and their schemas

The App.vue view routing uses an existing ref pattern:

- `showComparison: ref(false)` -- Controls whether comparison view is shown
- Set to true when Compare button clicked; false on close
