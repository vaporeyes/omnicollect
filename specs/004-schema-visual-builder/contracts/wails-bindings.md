# Wails IPC Contracts: Schema Visual Builder

**Date**: 2026-04-05
**Feature**: 004-schema-visual-builder

## SaveCustomModule

Validates and writes a module schema JSON file to disk.

**Go signature**:
```go
func (a *App) SaveCustomModule(schemaJSON string) (ModuleSchema, error)
```

**Frontend call**:
```typescript
import { SaveCustomModule } from '../../wailsjs/go/main/App'
const saved = await SaveCustomModule(JSON.stringify(schema, null, 2))
```

**Behavior**:
- Parses the JSON string into a ModuleSchema struct.
- Validates: non-empty ID, non-empty displayName, valid attribute
  types, unique attribute names, enum attributes have options.
- Writes formatted JSON to `~/.omnicollect/modules/{id}.json`.
- Reloads the in-memory module schemas so the new/updated schema is
  immediately available via `GetActiveModules()`.
- Returns the parsed ModuleSchema on success.
- Rejects with error if validation fails or file write fails.

**Input**: `schemaJSON` -- JSON string of the module schema.
**Output**: Parsed `ModuleSchema` struct.
**Error**: Validation failure, write failure, duplicate ID conflict.

## LoadModuleFile

Reads a specific module schema file from disk for editing.

**Go signature**:
```go
func (a *App) LoadModuleFile(moduleID string) (string, error)
```

**Frontend call**:
```typescript
import { LoadModuleFile } from '../../wailsjs/go/main/App'
const json = await LoadModuleFile("coins")
```

**Behavior**:
- Finds the schema file for the given module ID in the modules
  directory.
- Returns the raw JSON string (preserving formatting).
- Rejects with error if module not found.

**Input**: `moduleID` -- the module's ID string.
**Output**: Raw JSON string of the schema file.
**Error**: Module not found, read failure.
