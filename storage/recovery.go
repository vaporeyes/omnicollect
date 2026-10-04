// ABOUTME: Durable tenant-scoped deletion recovery with atomic capture and restore.
// ABOUTME: Bounded batches never overwrite live IDs or silently expire recovery data.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrDeleteSelectionMissing = errors.New("one or more selected items no longer exist; refresh first")
var ErrInvalidDeletion = errors.New("deletion rejected")

const MaxRecoveryBatches = 100
const MaxRecoveryBatchBytes = 8 << 20
const MaxRecoveryBytes = 64 << 20

type DeletionBatch struct {
	RecoveryID string   `json:"recoveryId"`
	Deleted    int      `json:"deleted"`
	CreatedAt  string   `json:"createdAt"`
	Titles     []string `json:"titles"`
}

func recoveryTx(ctx context.Context, db *sql.DB, table func(string) string, pg bool) (*sql.Tx, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	if pg {
		_, err = tx.ExecContext(ctx, "LOCK TABLE "+table("modules")+" IN SHARE MODE")
		if err == nil {
			_, err = tx.ExecContext(ctx, "LOCK TABLE "+table("items")+", "+table("deletion_batches")+" IN SHARE ROW EXCLUSIVE MODE")
		}
	} else {
		// Reserve the SQLite writer before reading snapshots; never upgrade a stale read.
		_, err = tx.ExecContext(ctx, "UPDATE deletion_batches SET item_count=item_count WHERE id='__lock__'")
	}
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return tx, nil
}
func deleteWithRecovery(ctx context.Context, db *sql.DB, table func(string) string, pg bool, ids []string) (DeletionBatch, error) {
	zero := DeletionBatch{}
	if err := validateIDs(ids); err != nil {
		return zero, fmt.Errorf("%w: %v", ErrInvalidDeletion, err)
	}
	if len(ids) == 0 {
		return zero, fmt.Errorf("%w: select at least one item", ErrInvalidDeletion)
	}
	tx, err := recoveryTx(ctx, db, table, pg)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback()
	holders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		holders[i] = parameter(pg, i+1)
		args[i] = id
	}
	predicate := " WHERE id IN (" + strings.Join(holders, ",") + ")"
	if err = checkReadBudget(ctx, tx, table("items")+predicate, itemReadBytes(pg), 500, MaxRecoveryBatchBytes, 0, args...); err != nil {
		if errors.Is(err, ErrMetadataBudget) {
			return zero, fmt.Errorf("%w: %w", ErrInvalidDeletion, err)
		}
		return zero, err
	}
	rows, err := tx.QueryContext(ctx, "SELECT id,module_id,title,purchase_price,images,tags,attributes,created_at,updated_at FROM "+table("items")+predicate+" ORDER BY id", args...)
	if err != nil {
		return zero, err
	}
	items, err := scanSnapshotItems(rows)
	closeErr := rows.Close()
	if err != nil {
		return zero, err
	}
	if closeErr != nil {
		return zero, closeErr
	}
	if len(items) != len(ids) {
		return zero, ErrDeleteSelectionMissing
	}
	payload, err := json.Marshal(items)
	if err != nil {
		return zero, err
	}
	if len(payload) > MaxRecoveryBatchBytes {
		return zero, fmt.Errorf("%w: recovery batch exceeds 8 MB; select fewer items", ErrInvalidDeletion)
	}
	result := DeletionBatch{RecoveryID: uuid.NewString(), Deleted: len(items), CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	for _, item := range items[:min(5, len(items))] {
		result.Titles = append(result.Titles, item.Title)
	}
	titles, _ := json.Marshal(result.Titles)
	var count, bytes int64
	length := "length(CAST(payload AS BLOB)) + length(CAST(titles AS BLOB))"
	if pg {
		length = "octet_length(payload) + octet_length(titles)"
	}
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*), COALESCE(SUM("+length+"),0) FROM "+table("deletion_batches")).Scan(&count, &bytes); err != nil {
		return zero, err
	}
	if count >= MaxRecoveryBatches || bytes+int64(len(payload)+len(titles)) > MaxRecoveryBytes {
		return zero, fmt.Errorf("%w: recovery storage is full; recover or explicitly discard batches before deleting more", ErrInvalidDeletion)
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO "+table("deletion_batches")+" (id,payload,item_count,created_at,titles) VALUES ("+parameter(pg, 1)+","+parameter(pg, 2)+","+parameter(pg, 3)+","+parameter(pg, 4)+","+parameter(pg, 5)+")", result.RecoveryID, string(payload), result.Deleted, result.CreatedAt, string(titles)); err != nil {
		return zero, err
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM "+table("items")+predicate, args...); err != nil {
		return zero, err
	}
	if err = tx.Commit(); err != nil {
		return zero, err
	}
	return result, nil
}
func recoverDeletion(ctx context.Context, db *sql.DB, table func(string) string, pg bool, id string) (int, error) {
	if _, err := uuid.Parse(id); err != nil {
		return 0, fmt.Errorf("invalid recovery id")
	}
	tx, err := recoveryTx(ctx, db, table, pg)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var raw string
	if err = tx.QueryRowContext(ctx, "SELECT payload FROM "+table("deletion_batches")+" WHERE id="+parameter(pg, 1), id).Scan(&raw); err != nil {
		return 0, err
	}
	if len(raw) > MaxRecoveryBatchBytes {
		return 0, fmt.Errorf("invalid recovery payload size")
	}
	var items []Item
	if err = json.Unmarshal([]byte(raw), &items); err != nil {
		return 0, err
	}
	if len(items) == 0 || len(items) > 500 {
		return 0, fmt.Errorf("invalid recovery item count")
	}
	modules := map[string]ModuleSchema{}
	stamp := time.Now().UTC().Truncate(time.Microsecond)
	// Validate every record first. Never restore obsolete versions that a stale editor can reuse.
	for _, item := range items {
		schema, ok := modules[item.ModuleID]
		if !ok {
			var schemaJSON string
			if err = tx.QueryRowContext(ctx, "SELECT schema_json FROM "+table("modules")+" WHERE id="+parameter(pg, 1), item.ModuleID).Scan(&schemaJSON); err != nil {
				return 0, fmt.Errorf("recovery module unavailable: %w", err)
			}
			if err = json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
				return 0, err
			}
			modules[item.ModuleID] = schema
		}
		if err = ValidateItem(item, schema); err != nil {
			return 0, fmt.Errorf("recovery conflicts with current schema: %w", err)
		}
		previous, e := time.Parse(time.RFC3339Nano, item.UpdatedAt)
		if e != nil {
			return 0, e
		}
		if !stamp.After(previous) {
			stamp = previous.Add(time.Microsecond)
		}
	}
	holders := make([]string, 9)
	for i := range holders {
		holders[i] = parameter(pg, i+1)
	}
	query := "INSERT INTO " + table("items") + " (id,module_id,title,purchase_price,images,tags,attributes,created_at,updated_at) VALUES (" + strings.Join(holders, ",") + ")"
	for _, item := range items {
		if item.Images == nil {
			item.Images = []string{}
		}
		if item.Tags == nil {
			item.Tags = []string{}
		}
		if item.Attributes == nil {
			item.Attributes = map[string]any{}
		}
		images, _ := json.Marshal(item.Images)
		tags, _ := json.Marshal(item.Tags)
		attrs, _ := json.Marshal(item.Attributes)
		if _, err = tx.ExecContext(ctx, query, item.ID, item.ModuleID, item.Title, item.PurchasePrice, string(images), string(tags), string(attrs), item.CreatedAt, stamp.Format(time.RFC3339Nano)); err != nil {
			return 0, fmt.Errorf("recovery refused; existing IDs are never overwritten: %w", err)
		}
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM "+table("deletion_batches")+" WHERE id="+parameter(pg, 1), id); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return len(items), nil
}
func listDeletions(ctx context.Context, db *sql.DB, table string) ([]DeletionBatch, error) {
	rows, err := db.QueryContext(ctx, "SELECT id,item_count,created_at,titles FROM "+table+" ORDER BY created_at DESC,id LIMIT 100")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	batches := []DeletionBatch{}
	for rows.Next() {
		var b DeletionBatch
		var titles string
		if err = rows.Scan(&b.RecoveryID, &b.Deleted, &b.CreatedAt, &titles); err != nil {
			return nil, err
		}
		if err = json.Unmarshal([]byte(titles), &b.Titles); err != nil {
			return nil, err
		}
		batches = append(batches, b)
	}
	return batches, rows.Err()
}
func (s *SQLiteStore) DeleteWithRecovery(ids []string) (DeletionBatch, error) {
	return deleteWithRecovery(s.operationContext(), s.db, func(n string) string { return n }, false, ids)
}
func (s *PostgresStore) DeleteWithRecovery(ids []string) (DeletionBatch, error) {
	if err := s.validateTenant(); err != nil {
		return DeletionBatch{}, err
	}
	return deleteWithRecovery(s.operationContext(), s.db, s.table, true, ids)
}
func (s *SQLiteStore) RecoverDeletion(id string) (int, error) {
	return recoverDeletion(s.operationContext(), s.db, func(n string) string { return n }, false, id)
}
func (s *PostgresStore) RecoverDeletion(id string) (int, error) {
	if err := s.validateTenant(); err != nil {
		return 0, err
	}
	return recoverDeletion(s.operationContext(), s.db, s.table, true, id)
}
func (s *SQLiteStore) ListDeletions() ([]DeletionBatch, error) {
	return listDeletions(s.operationContext(), s.db, "deletion_batches")
}
func (s *PostgresStore) ListDeletions() ([]DeletionBatch, error) {
	if err := s.validateTenant(); err != nil {
		return nil, err
	}
	return listDeletions(s.operationContext(), s.db, s.table("deletion_batches"))
}
