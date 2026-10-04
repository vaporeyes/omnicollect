// ABOUTME: Bounded, consistent CSV reads reject oversized selections instead of partial exports.
// ABOUTME: Enforces encoded read budgets before decoding; output has independent column/byte caps.
package storage

import (
	"context"
	"database/sql"
	"strings"
)

const MaxCSVColumns = 1024
const MaxCSVBytes = 8 << 20

func exportItemsCSV(ctx context.Context, db *sql.DB, table string, pg bool, ids []string, modules []ModuleSchema) (string, error) {
	if err := validateIDs(ids); err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "", nil
	}
	options := &sql.TxOptions{ReadOnly: true}
	if pg {
		options.Isolation = sql.LevelRepeatableRead
	}
	tx, err := db.BeginTx(ctx, options)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	holders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		holders[i] = parameter(pg, i+1)
		args[i] = id
	}
	from := table + " WHERE id IN (" + strings.Join(holders, ",") + ")"
	if err = checkReadBudget(ctx, tx, from, itemReadBytes(pg), 500, MaxCSVBytes, MaxMetadataRowBytes, args...); err != nil {
		return "", err
	}
	rows, err := tx.QueryContext(ctx, "SELECT id,module_id,title,purchase_price,images,tags,attributes,created_at,updated_at FROM "+from+" ORDER BY updated_at DESC,id", args...)
	if err != nil {
		return "", err
	}
	items, err := scanSnapshotItems(rows)
	closeErr := rows.Close()
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return buildCSV(items, modules)
}
