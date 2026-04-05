# Quickstart: Core Engine (Data & IPC)

**Feature**: 001-core-engine-data-ipc

## Prerequisites

- Go 1.21+ installed
- Node.js 18+ and npm installed
- Wails CLI v2 installed (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

## Setup

1. Clone the repository and checkout the feature branch:
   ```bash
   git clone <repo-url> omnicollect
   cd omnicollect
   git checkout 001-core-engine-data-ipc
   ```

2. Install Go dependencies:
   ```bash
   go mod tidy
   ```

3. Install frontend dependencies:
   ```bash
   cd frontend && npm install && cd ..
   ```

## Development

Run the app in development mode (hot reload for Go and Vue):
```bash
wails dev
```

This will:
- Generate TypeScript bindings in `frontend/wailsjs/`
- Start the Vite dev server for the Vue frontend
- Build and run the Go backend
- Open the application window

## Verify Core Functionality

### 1. Module Schema Loading

Place a test schema file at `~/.omnicollect/modules/coins.json`:

```json
{
  "id": "coins",
  "displayName": "Coins",
  "description": "Coin collection",
  "attributes": [
    {
      "name": "year",
      "type": "number",
      "required": true,
      "display": { "label": "Mint Year", "widget": "text" }
    },
    {
      "name": "country",
      "type": "string",
      "required": true,
      "display": { "label": "Country of Origin", "widget": "text" }
    },
    {
      "name": "condition",
      "type": "enum",
      "options": ["Poor", "Fair", "Good", "Very Good", "Fine", "Very Fine", "Extremely Fine", "Uncirculated"],
      "display": { "label": "Condition", "widget": "dropdown" }
    }
  ]
}
```

Call `GetActiveModules()` from the frontend console or a test Vue
component. Expect one module with id "coins" and three attributes.

### 2. Item CRUD

Save an item via `SaveItem()`:

```typescript
const item = await SaveItem({
  id: "",
  moduleId: "coins",
  title: "1921 Morgan Silver Dollar",
  purchasePrice: 45.00,
  images: [],
  attributes: {
    year: 1921,
    country: "United States",
    condition: "Very Fine"
  },
  createdAt: "",
  updatedAt: ""
});
// item.id is now a UUID, createdAt/updatedAt are populated
```

Retrieve items:
```typescript
const allCoins = await GetItems("", "coins");
const searchResults = await GetItems("morgan", "");
```

### 3. Full-Text Search

After saving several items, verify FTS5 search:
- Search by title keyword returns matching items.
- Search by attribute value (e.g., "United States") returns items
  with that value in any attribute.
- Empty search string returns all items (filtered by moduleID if
  provided).

## Build for Production

```bash
wails build
```

The output binary is in `build/bin/`. It is a single self-contained
executable with the Vue frontend embedded.

## Database Location

The SQLite database is created at the application's user data
directory (platform-dependent, resolved via `os.UserConfigDir()`):
- macOS: `~/Library/Application Support/OmniCollect/collection.db`
- Linux: `~/.config/OmniCollect/collection.db`
- Windows: `%APPDATA%\OmniCollect\collection.db`
