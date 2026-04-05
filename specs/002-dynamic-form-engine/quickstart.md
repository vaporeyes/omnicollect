# Quickstart: Dynamic Form Engine

**Feature**: 002-dynamic-form-engine

## Prerequisites

- Iteration 1 (Core Engine) must be complete and functional.
- Go 1.21+, Node.js 18+, Wails CLI v2 installed.
- At least one module schema in `~/.omnicollect/modules/` (e.g.,
  the `coins.json` from Iteration 1).

## Setup

1. Checkout the feature branch:
   ```bash
   git checkout 002-dynamic-form-engine
   ```

2. Install the new frontend dependency:
   ```bash
   cd frontend && npm install && cd ..
   ```

## Development

```bash
wails dev
```

The app opens with the dynamic form engine. The Vite dev server
provides hot reload for Vue component changes.

## Verify Core Functionality

### 1. Module Loading

On startup, the sidebar should display all available collection
types loaded from `~/.omnicollect/modules/`. With the sample
`coins.json`, you should see "Coins" listed.

### 2. Dynamic Form Rendering

1. Click "Coins" in the module selector.
2. Verify the form shows:
   - Base fields: Title (text input), Purchase Price (number input)
   - Custom fields: Mint Year (number input), Country of Origin
     (text input), Condition (dropdown with options from Poor to
     Uncirculated)
3. Labels should match the `display.label` values from the schema.

### 3. Save a New Item

1. Fill in: Title = "1921 Morgan Silver Dollar", Purchase Price = 45,
   Mint Year = 1921, Country = "United States", Condition = "Very Fine"
2. Click Save.
3. Verify the item appears in the item list with the correct title.

### 4. Search

1. Type "morgan" in the search box.
2. Verify the 1921 Morgan Silver Dollar appears.
3. Clear the search, type "xyz".
4. Verify the empty state message appears.

### 5. Edit an Item

1. Click the item in the list.
2. Verify all fields are pre-populated with saved values.
3. Change the condition to "Extremely Fine".
4. Save and verify the change persists.

### 6. Multiple Collection Types

1. Add a second schema file (e.g., `books.json`) to
   `~/.omnicollect/modules/`.
2. Restart the app (`wails dev` will restart).
3. Verify both "Coins" and "Books" appear in the module selector.
4. Add an item under "Books".
5. Filter the item list by "Coins" -- verify only coin items show.
6. Filter by "Books" -- verify only book items show.

## Build

```bash
wails build
```

The production binary in `build/bin/` includes the complete frontend
with dynamic form engine.
