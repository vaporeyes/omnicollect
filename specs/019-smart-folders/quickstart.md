# Quickstart: Smart Folders (Saved Views)

**Branch**: `019-smart-folders` | **Date**: 2026-04-10

## Prerequisites

- Node.js (for frontend dev)
- Go 1.25+ (for backend -- no changes needed)
- Wails v2 CLI (for desktop dev mode)

## Setup

No new dependencies required. Uses existing Vue 3, Pinia, and API client.

```bash
wails dev    # Start development server
```

## Development Workflow

1. **Store first**: Implement `smartFolderStore.ts` with CRUD operations and settings persistence. Write Vitest tests.

2. **Sidebar component**: Build `SmartFolders.vue` with folder list, "Save Current View" button, inline text field, and context menu integration.

3. **App integration**: Wire SmartFolders component into App.vue sidebar, connect apply action to set collection store state.

## Key Files to Modify

| File | Change |
|------|--------|
| `frontend/src/stores/smartFolderStore.ts` | New: Pinia store for Smart Folder CRUD + persistence |
| `frontend/src/components/SmartFolders.vue` | New: sidebar section component |
| `frontend/src/App.vue` | Add SmartFolders to sidebar, wire apply handler, clear active on manual filter change |

## Testing

```bash
cd frontend && npm test    # Run all tests
cd frontend && npx vitest run src/stores/__tests__/smartFolderStore.test.ts    # Store tests only
```

## Verification

1. Apply a module, search query, and filters
2. Click "Save Current View" in sidebar, type a name, press Enter
3. Smart Folder appears in sidebar with bookmark icon
4. Change to a different module/view
5. Click the Smart Folder -- original view state restored
6. Right-click Smart Folder -- context menu with Rename/Delete
7. Close and reopen app -- Smart Folder persists
