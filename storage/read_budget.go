// ABOUTME: Encoded-byte and row budgets checked in the same transaction before bulk decoding.
// ABOUTME: Bounds retained metadata, not exact Go heap overhead; errors never return partial snapshots.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrMetadataBudget = errors.New("metadata read exceeds resource budget")

const MaxSnapshotReadBytes = 32 << 20
const MaxModuleReadBytes = 8 << 20
const MaxMetadataRowBytes = 1 << 20
const MaxTagResults = 10000
const MaxShowcaseResults = 10000

func encodedBytes(expression string, pg bool) string {
	if pg {
		return "COALESCE(octet_length(CAST(" + expression + " AS TEXT)),0)"
	}
	return "COALESCE(length(CAST(" + expression + " AS BLOB)),0)"
}
func itemReadBytes(pg bool) string {
	parts := []string{"512"}
	for _, field := range []string{"id", "module_id", "title", "images", "tags", "attributes", "created_at", "updated_at"} {
		parts = append(parts, encodedBytes(field, pg))
	}
	return strings.Join(parts, "+")
}

// A zero aggregate limit permits streaming arbitrary row counts, but the row limit remains enforced.
func checkReadBudget(ctx context.Context, tx *sql.Tx, from, bytes string, maxRows, maxBytes, maxRow int64, args ...any) error {
	var count, total, largest int64
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*), COALESCE(SUM("+bytes+"),0), COALESCE(MAX("+bytes+"),0) FROM "+from, args...).Scan(&count, &total, &largest); err != nil {
		return err
	}
	if (maxRows > 0 && count > maxRows) || (maxBytes > 0 && total > maxBytes) || (maxRow > 0 && largest > maxRow) {
		return fmt.Errorf("%w (rows=%d, encoded bytes=%d, largest row=%d)", ErrMetadataBudget, count, total, largest)
	}
	return nil
}
func readModules(ctx context.Context, db *sql.DB, table string, pg bool) ([]ModuleSchema, error) {
	options := &sql.TxOptions{ReadOnly: true}
	if pg {
		options.Isolation = sql.LevelRepeatableRead
	}
	tx, err := db.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err = checkReadBudget(ctx, tx, table, encodedBytes("schema_json", pg), maxSnapshotModules, MaxModuleReadBytes, 256<<10); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, "SELECT schema_json FROM "+table+" ORDER BY display_name,id")
	if err != nil {
		return nil, err
	}
	modules, err := scanModules(rows)
	closeErr := rows.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return modules, nil
}
