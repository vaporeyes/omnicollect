// ABOUTME: Snapshot-consistent public gallery reads project only titles and cover references.
// ABOUTME: Counts and module names share the transaction; row and offset budgets bound paging.
package storage

import (
	"context"
	"database/sql"
	"fmt"
)

const GalleryPageSize = 24
const MaxGalleryPages = MaxItemPageOffset/GalleryPageSize + 1

type GalleryRecord struct{ ID, Title, PrimaryImage string }
type GalleryPage struct {
	CollectionName               string
	Items                        []GalleryRecord
	TotalItems, Page, TotalPages int
	Truncated                    bool
}

func (s *SQLiteStore) ReadGalleryPage(moduleID string, page int) (GalleryPage, error) {
	return readGalleryPage(s.operationContext(), s.db, func(n string) string { return n }, false, moduleID, page)
}
func (s *PostgresStore) ReadGalleryPage(moduleID string, page int) (GalleryPage, error) {
	if err := s.validateTenant(); err != nil {
		return GalleryPage{}, err
	}
	return readGalleryPage(s.operationContext(), s.db, s.table, true, moduleID, page)
}
func readGalleryPage(ctx context.Context, db *sql.DB, table func(string) string, pg bool, moduleID string, page int) (GalleryPage, error) {
	result := GalleryPage{Items: []GalleryRecord{}}
	if !safeModuleID.MatchString(moduleID) || page < 1 || page > MaxGalleryPages {
		return result, fmt.Errorf("invalid gallery page")
	}
	options := &sql.TxOptions{ReadOnly: true}
	if pg {
		options.Isolation = sql.LevelRepeatableRead
	}
	tx, err := db.BeginTx(ctx, options)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	if err = tx.QueryRowContext(ctx, "SELECT display_name FROM "+table("modules")+" WHERE id="+parameter(pg, 1), moduleID).Scan(&result.CollectionName); err != nil {
		return result, err
	}
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table("items")+" WHERE module_id="+parameter(pg, 1), moduleID).Scan(&result.TotalItems); err != nil {
		return result, err
	}
	result.TotalPages = 1
	if result.TotalItems > 0 {
		result.TotalPages = (result.TotalItems-1)/GalleryPageSize + 1
	}
	result.Truncated = result.TotalPages > MaxGalleryPages
	result.TotalPages = min(result.TotalPages, MaxGalleryPages)
	result.Page = min(page, result.TotalPages)
	cover := "COALESCE(json_extract(images,'$[0]'),'')"
	if pg {
		cover = "COALESCE(images->>0,'')"
	}
	rows, err := tx.QueryContext(ctx, "SELECT id,title,"+cover+" FROM "+table("items")+" WHERE module_id="+parameter(pg, 1)+" ORDER BY updated_at DESC,id ASC LIMIT "+parameter(pg, 2)+" OFFSET "+parameter(pg, 3), moduleID, GalleryPageSize, (result.Page-1)*GalleryPageSize)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var item GalleryRecord
		if err = rows.Scan(&item.ID, &item.Title, &item.PrimaryImage); err != nil {
			rows.Close()
			return result, err
		}
		result.Items = append(result.Items, item)
	}
	err = rows.Err()
	closeErr := rows.Close()
	if err != nil {
		return result, err
	}
	if closeErr != nil {
		return result, closeErr
	}
	if err = tx.Commit(); err != nil {
		return result, err
	}
	return result, nil
}
