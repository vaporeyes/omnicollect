// ABOUTME: Shared decoding for database-backed module schemas.
// ABOUTME: Corrupt schemas are surfaced as errors rather than silently discarded.
package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func scanModules(rows *sql.Rows) ([]ModuleSchema, error) {
	modules := []ModuleSchema{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var schema ModuleSchema
		if err := json.Unmarshal([]byte(raw), &schema); err != nil {
			return nil, fmt.Errorf("invalid stored module: %w", err)
		}
		if err := ValidateModuleSchema(&schema); err != nil {
			return nil, err
		}
		modules = append(modules, schema)
	}
	return modules, rows.Err()
}
