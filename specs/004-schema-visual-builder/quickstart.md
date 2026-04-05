# Quickstart: Schema Visual Builder

**Feature**: 004-schema-visual-builder

## Prerequisites

- Iterations 1-3 must be complete.
- `wails dev` runs successfully.

## Setup

```bash
git checkout 004-schema-visual-builder
cd frontend && npm install && cd ..
```

## Development

```bash
wails dev
```

## Verify Core Functionality

### 1. Create a New Schema

1. Click "Schema Builder" in the sidebar.
2. Enter display name: "Vinyl Records".
3. Verify the ID auto-generates as "vinyl-records".
4. Click "Add Field":
   - Name: "artist", Type: string, Required: yes
5. Click "Add Field":
   - Name: "year", Type: number, Required: yes
6. Click "Add Field":
   - Name: "genre", Type: enum
   - Add options: "Rock", "Jazz", "Classical", "Electronic"
7. Verify the live preview shows three form fields with correct types.
8. Click Save.
9. Verify `~/.omnicollect/modules/vinyl-records.json` exists.
10. Verify "Vinyl Records" appears in the module selector immediately.

### 2. Code Editor Sync

1. Open the schema builder (new or existing schema).
2. Add a field in the visual builder.
3. Verify the JSON in the code editor updates immediately.
4. Edit the JSON directly -- change a field name.
5. Verify the visual builder updates.
6. Type invalid JSON (e.g., remove a closing brace).
7. Verify the visual preview does not crash -- it holds last state.
8. Fix the JSON.
9. Verify the preview recovers.

### 3. Edit Existing Schema

1. Open the builder for the "Coins" schema (from Iteration 1).
2. Verify both panes show the current schema definition.
3. Add a new field: "denomination", type: string.
4. Save.
5. Verify the file on disk is updated.
6. Go to the collection view -- existing coin items still display.
7. Edit a coin item -- verify the new "denomination" field appears.

### 4. Validation

1. Try to save with an empty display name -- verify error shown.
2. Try to add two fields with the same name -- verify error shown.
3. Try to add an enum field with no options -- verify error shown.

### 5. Unsaved Changes

1. Make changes in the builder without saving.
2. Click close/cancel.
3. Verify a confirmation prompt appears.

## Build

```bash
wails build
```
