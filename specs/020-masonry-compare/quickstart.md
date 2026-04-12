# Quickstart: Masonry Grid & Item Comparison

**Branch**: `020-masonry-compare` | **Date**: 2026-04-11

## Prerequisites

- Node.js 18+ and npm
- Go 1.25+ (for Wails dev mode, though no Go changes in this feature)
- Wails CLI (`wails dev` for desktop) or standalone HTTP mode (`go run . --serve`)

## Development Setup

```bash
git checkout 020-masonry-compare

# Frontend only (fastest iteration)
cd frontend
npm install
npm run dev

# Full desktop mode
wails dev

# Standalone HTTP mode
go run . --serve
```

## Testing the Masonry Grid

1. Start the app and navigate to any collection with multiple items
2. Switch to Grid view if not already active
3. Verify cards display at varying heights based on image aspect ratios
4. Resize the browser window and confirm columns reflow without gaps
5. Hover over cards to confirm scale + shadow effects work
6. Verify frosted glass captions show title, module name, and date
7. Test with a collection that has items with no images (placeholder cards)

## Testing Comparison Mode

1. In any collection view, select exactly two items (click checkboxes)
2. Verify the "Compare" button appears in the floating bulk action bar
3. Click Compare to enter the comparison view
4. Verify both items appear side-by-side with their image galleries
5. Click next/prev on one gallery and confirm both advance together
6. Scroll down to the diff table; verify differing values are highlighted
7. Click close/back to return to the grid with selection preserved

## Testing Edge Cases

- Select 1 item: Compare button should not appear
- Select 3 items: Compare button should disappear
- Compare two items from different modules: union of attributes shown
- Compare two items with identical attributes: no highlighting
- Compare items where one has no images: placeholder on that side
- Narrow the viewport below 768px in comparison mode: should stack vertically

## Running Automated Tests

```bash
cd frontend
npm test              # Run all Vitest tests
npm test -- --watch   # Watch mode during development
```

## Key Files to Modify

| File | Change |
|------|--------|
| `frontend/src/components/CollectionGrid.vue` | CSS: column-count masonry, remove aspect-ratio: 1 |
| `frontend/src/components/BulkActionBar.vue` | Add Compare button + emit |
| `frontend/src/components/ComparisonView.vue` | New: side-by-side layout, synced galleries, diff table |
| `frontend/src/App.vue` | Add comparison view routing, showComparison state, onCompare handler |
