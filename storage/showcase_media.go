// ABOUTME: Checks public cover-image references without loading a tenant's collection.
// ABOUTME: Only the first image actually displayed by a showcase is publishable by default.
package storage

import "context"

func (s *SQLiteStore) HasShowcaseImage(ctx context.Context, moduleID, filename string) (bool, error) {
	if err := ValidateFilename(filename); err != nil {
		return false, err
	}
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM items WHERE module_id=? AND json_extract(images,'$[0]')=?)`, moduleID, filename).Scan(&exists)
	return exists, err
}
func (s *PostgresStore) HasShowcaseImage(ctx context.Context, moduleID, filename string) (bool, error) {
	if err := s.validateTenant(); err != nil {
		return false, err
	}
	if err := ValidateFilename(filename); err != nil {
		return false, err
	}
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM `+s.table("items")+` WHERE module_id=$1 AND images->>0=$2)`, moduleID, filename).Scan(&exists)
	return exists, err
}
