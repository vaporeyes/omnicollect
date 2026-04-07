# Quickstart: Command Palette Implementation

**Branch**: `006-command-palette` | **Date**: 2026-04-07

## Prerequisites

- Wails v2 development environment set up (`wails dev` works)
- Existing codebase on the `006-command-palette` branch
- Familiarity with Vue 3 Composition API and Pinia stores

## Files to Create

1. **`frontend/src/components/CommandPalette.vue`** -- The palette overlay component containing:
   - Blurred glass backdrop overlay (Teleported to body)
   - Large search input with auto-focus
   - Debounced search calling `GetItems(query, "")` for cross-module results
   - Quick action matching against keyword map
   - Keyboard navigation (Up/Down/Enter/Escape)
   - Result rendering with thumbnails, titles, and module badges

## Files to Modify

1. **`frontend/src/App.vue`** -- Add:
   - Import and render `CommandPalette` component
   - Palette visibility state (`showPalette` ref)
   - Cmd/Ctrl+K handler in existing `onGlobalKeydown` to toggle palette
   - Handler for palette `selectItem` event (navigate to ItemDetail)
   - Handler for palette `action` event (dispatch quick actions to existing handlers)

2. **`frontend/src/stores/collectionStore.ts`** -- Add:
   - `searchAllItems(query: string)` action that calls `GetItems(query, "")` without modifying the store's `items` ref (returns results directly, used only by the palette)

## Implementation Order

1. Add `searchAllItems` to collectionStore (small, self-contained)
2. Build CommandPalette.vue (bulk of the work)
3. Wire into App.vue (integration)
4. Update CLAUDE.md and README (documentation, Constitution VI)

## Key Design Decisions

- **No new backend binding**: Existing `GetItems` with empty moduleID searches all modules
- **Palette state is local**: open/closed, query, highlighted index live in the component, not Pinia
- **Results capped at 25**: Prevents DOM bloat
- **200ms debounce**: Balances responsiveness and query frequency
- **Thumbnails only**: Constitution Principle IV compliance
- **Z-index 3000**: Above lightbox (1000) and context menu (2000), below toast (9999)

## Acceptance Test Flow

1. Press Cmd/Ctrl+K -- palette opens, input focused
2. Type item name -- results appear with thumbnails and module badges
3. Arrow Down/Up -- highlight moves through results
4. Press Enter -- palette closes, item detail view opens
5. Press Cmd/Ctrl+K -- palette opens again
6. Type "new" -- quick actions appear above results
7. Select "Add New Item" -- palette closes, new item form opens
8. Press Escape -- palette closes without action
9. Click outside palette -- palette closes without action
10. Press Cmd/Ctrl+K while palette is open -- palette closes (toggle)
