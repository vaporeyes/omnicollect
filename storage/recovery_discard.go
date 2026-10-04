// ABOUTME: Explicit permanent disposal of a saved deletion batch, never live items.
// ABOUTME: Recovery and disposal serialize through the database; media bytes are retained.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)

func discardDeletion(ctx context.Context, db *sql.DB, table string, pg bool, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid recovery id")
	}
	result, err := db.ExecContext(ctx, "DELETE FROM "+table+" WHERE id="+parameter(pg, 1), id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}
func (s *SQLiteStore) DiscardDeletion(id string) error {
	return discardDeletion(s.operationContext(), s.db, "deletion_batches", false, id)
}
func (s *PostgresStore) DiscardDeletion(id string) error {
	if err := s.validateTenant(); err != nil {
		return err
	}
	return discardDeletion(s.operationContext(), s.db, s.table("deletion_batches"), true, id)
}
