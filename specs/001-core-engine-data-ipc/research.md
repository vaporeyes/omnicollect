# Research: Core Engine (Data & IPC)

**Date**: 2026-04-05
**Feature**: 001-core-engine-data-ipc

## R1: CGO-Free SQLite Driver

**Decision**: Use `modernc.org/sqlite` (pure Go, no CGO).

**Rationale**: Simplifies cross-platform builds (no C compiler required).
The driver registers as `"sqlite"` via `import _ "modernc.org/sqlite"`.
FTS5 is fully supported (compiled with `SQLITE_ENABLE_FTS5`). Underlying
SQLite version is 3.51.3. Performance is 2-3x slower than CGO `mattn/go-sqlite3`
on write-heavy workloads, which is negligible for a desktop collection app.

**Alternatives considered**:
- `mattn/go-sqlite3`: Requires CGO, complicates cross-compilation. Better
  raw performance but unnecessary for this use case.

**DSN pattern**:
```
file:collection.db?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)
```

## R2: FTS5 Indexing Strategy for JSON Attributes

**Decision**: External content FTS5 table with triggers. Flatten all
JSON attribute values into a single `attrs_text` column using
`json_each()` in trigger bodies.

**Rationale**: Triggers guarantee index consistency within the same
transaction. External content avoids duplicating text data. A single
flattened column handles arbitrary attribute keys without schema
coupling.

**Alternatives considered**:
- Manual sync from Go code: Risk of missed call paths causing index
  drift. Rejected.
- Separate FTS columns per attribute: Only viable with a small, stable
  attribute set. Incompatible with EAV/arbitrary-key design. Rejected.
- Key-prefixed tokens (`color:red`): Useful for field-scoped search.
  Can be added later without changing the trigger structure. Deferred.

**Sync triggers**: INSERT, UPDATE, and DELETE triggers on the `items`
table populate/remove FTS entries. A `rebuild` command is available as
a safety net after bulk operations.

## R3: Wails v2 Project Structure and Bindings

**Decision**: Standard Wails v2 layout with a single `App` struct
exposing `SaveItem`, `GetItems`, and `GetActiveModules`.

**Rationale**: Wails generates TypeScript bindings automatically from
exported Go methods on structs listed in `options.Bind`. A single
struct keeps the binding surface simple for this iteration. Methods
returning `(value, error)` map to `Promise<value>` on the frontend,
with errors becoming rejected promises.

**Key conventions**:
- Go backend code at project root (`app.go`, `db.go`, `modules.go`,
  `models.go`).
- Vue 3 frontend under `frontend/`.
- Generated bindings in `frontend/wailsjs/` (auto-regenerated, never
  edit manually).
- `wails.json` at project root configures build pipeline.
- `wails dev` provides hot reload for both Go and Vue changes.
- `wails build` embeds `frontend/dist/` into a single binary via
  `//go:embed`.

**Alternatives considered**:
- Multiple bound structs (e.g., `ItemService`, `ModuleService`):
  Premature for three methods. Can split later when the App struct
  grows. Rejected for now.

## R4: Module Schema Representation

**Decision**: Custom Go structs (`ModuleSchema`, `AttributeSchema`,
`DisplayHints`) with `encoding/json` unmarshaling. No third-party
JSON Schema library.

**Rationale**: The Go backend loads schemas and serves them to the
Vue frontend for UI generation. It does not need to validate arbitrary
JSON against arbitrary schemas. Custom structs provide compile-time
type safety, clear documentation, and zero extra dependencies. A
lightweight `Validate()` method checks structural invariants (non-empty
ID, valid attribute types, no duplicates).

**Alternatives considered**:
- `santhosh-tekuri/jsonschema`: Full JSON Schema validator. Overkill
  for loading controlled schema files. Rejected.
- `xeipuuv/gojsonschema`: Same purpose, less actively maintained.
  Rejected.

## R5: UUID Generation

**Decision**: `github.com/google/uuid` for v4 UUIDs.

**Rationale**: De facto standard in Go. Single function call
(`uuid.New().String()`). Widely adopted, well maintained.

**Alternatives considered**:
- `gofrs/uuid`: No practical advantage, less adoption. Rejected.

## Summary of Dependencies

| Package | Purpose |
|---------|---------|
| `modernc.org/sqlite` | CGO-free SQLite driver |
| `github.com/google/uuid` | UUID v4 generation |
| `github.com/wailsapp/wails/v2` | Desktop framework + IPC |

No JSON Schema library needed. Standard library `encoding/json` and
`database/sql` cover the remaining requirements.
