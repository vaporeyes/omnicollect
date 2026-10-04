// ABOUTME: Tenant-wide aggregates independent of the current item page or filters.
// ABOUTME: Uses a single read snapshot, bounded module groups and JSON-safe price totals.
package storage

import (
	"context"
	"database/sql"
	"math"
	"strconv"
)

const MaxSummaryModules = 1000

type ModuleCount struct {
	ModuleID string `json:"moduleId"`
	Items    int64  `json:"items"`
}
type CollectionSummary struct {
	Items            int64         `json:"items"`
	PricedItems      int64         `json:"pricedItems"`
	InvalidPrices    int64         `json:"invalidPrices"`
	PurchaseTotal    *float64      `json:"purchaseTotal"`
	ValueAvailable   bool          `json:"valueAvailable"`
	Modules          []ModuleCount `json:"modules"`
	ModulesTruncated bool          `json:"modulesTruncated"`
}

func (s *SQLiteStore) CollectionSummary() (CollectionSummary, error) {
	return readCollectionSummary(s.operationContext(), s.db, "items", false)
}
func (s *PostgresStore) CollectionSummary() (CollectionSummary, error) {
	if err := s.validateTenant(); err != nil {
		return CollectionSummary{}, err
	}
	return readCollectionSummary(s.operationContext(), s.db, s.table("items"), true)
}
func readCollectionSummary(ctx context.Context, db *sql.DB, table string, postgres bool) (CollectionSummary, error) {
	options := &sql.TxOptions{ReadOnly: true}
	if postgres {
		options.Isolation = sql.LevelRepeatableRead
	}
	tx, err := db.BeginTx(ctx, options)
	if err != nil {
		return CollectionSummary{}, err
	}
	defer tx.Rollback()
	result := CollectionSummary{Modules: []ModuleCount{}, ValueAvailable: true}
	validPrice := "CASE WHEN purchase_price >= 0 AND purchase_price <= 1.7976931348623157e308 THEN purchase_price ELSE NULL END"
	sumPrice := validPrice
	// PostgreSQL double-precision SUM raises an error on overflow; NUMERIC lets us
	// detect representability ourselves and return an explicit unavailable value.
	if postgres {
		sumPrice = "CAST(CAST((" + validPrice + ") AS TEXT) AS NUMERIC)"
	}
	var total sql.NullString
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*), COUNT("+validPrice+"), COUNT(purchase_price)-COUNT("+validPrice+"), SUM("+sumPrice+") FROM "+table).Scan(&result.Items, &result.PricedItems, &result.InvalidPrices, &total)
	if err != nil {
		return CollectionSummary{}, err
	}
	if total.Valid {
		value, parseErr := strconv.ParseFloat(total.String, 64)
		if parseErr != nil || math.IsInf(value, 0) || math.IsNaN(value) {
			result.ValueAvailable = false
		} else {
			result.PurchaseTotal = &value
		}
	}
	rows, err := tx.QueryContext(ctx, "SELECT module_id, COUNT(*) FROM "+table+" GROUP BY module_id ORDER BY module_id LIMIT "+strconv.Itoa(MaxSummaryModules+1))
	if err != nil {
		return CollectionSummary{}, err
	}
	for rows.Next() {
		var group ModuleCount
		if err = rows.Scan(&group.ModuleID, &group.Items); err != nil {
			rows.Close()
			return CollectionSummary{}, err
		}
		if len(result.Modules) == MaxSummaryModules {
			result.ModulesTruncated = true
			break
		}
		result.Modules = append(result.Modules, group)
	}
	rowErr := rows.Err()
	closeErr := rows.Close()
	if rowErr != nil {
		return CollectionSummary{}, rowErr
	}
	if closeErr != nil {
		return CollectionSummary{}, closeErr
	}
	if err = tx.Commit(); err != nil {
		return CollectionSummary{}, err
	}
	return result, nil
}
