# Component Interface Contracts

**Branch**: `020-masonry-compare` | **Date**: 2026-04-11

## CollectionGrid.vue (Modified)

No prop or emit changes. Only CSS layout changes (masonry).

**Props** (unchanged):
- `items: Item[]`
- `modules: ModuleSchema[]`

**Emits** (unchanged):
- `select(item: Item)`
- `viewImage(item: Item, filename: string)`
- `addItem()`
- `itemContextMenu(item: Item, x: number, y: number)`

## BulkActionBar.vue (Modified)

**Props** (modified):
- `count: number` (existing)

**Emits** (added):
- `delete()` (existing)
- `export()` (existing)
- `editModule()` (existing)
- `deselectAll()` (existing)
- `compare()` **NEW** -- Emitted when Compare button is clicked; only available when count === 2

**Behavior**: Compare button renders conditionally when `count === 2`.

## ComparisonView.vue (New)

**Props**:
- `itemA: Item` -- First item to compare
- `itemB: Item` -- Second item to compare
- `schemaA: ModuleSchema | null` -- Schema for item A (null if module not found)
- `schemaB: ModuleSchema | null` -- Schema for item B (null if module not found)

**Emits**:
- `close()` -- User clicks back/close button; parent returns to grid view
- `viewImage(filename: string)` -- User clicks image for full-size lightbox (optional, reuses existing lightbox)

**Internal State**:
- `activeImageIndex: ref(0)` -- Synchronized gallery position

**Computed**:
- `diffRows: ComputedRef<DiffRow[]>` -- Array of diff rows for the attribute table

**DiffRow Type**:
```typescript
interface DiffRow {
  label: string
  valueA: string | null
  valueB: string | null
  isDifferent: boolean
}
```

## App.vue (Modified)

**New State**:
- `showComparison: ref(false)` -- Whether comparison view is active
- `comparisonItems: ref<[Item, Item] | null>(null)` -- The two items being compared

**New Handler**:
- `onCompare()` -- Reads two selected IDs from selectionStore, resolves Item objects from collectionStore.items, sets comparisonItems and showComparison = true

**View Priority** (updated):
```
showComparison ? ComparisonView
: showForm ? DynamicForm
: showDetail ? ItemDetail
: showDashboard ? DashboardView
: viewMode === 'list' ? ItemList
: CollectionGrid
```
