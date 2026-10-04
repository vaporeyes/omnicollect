// ABOUTME: Bounded item-page contracts shared by SQLite and PostgreSQL.
// ABOUTME: Applies SQL limits before scanning and uses one lookahead row for hasMore.
package storage

import (
	"errors"
	"fmt"
)

var ErrPaginationRequired = errors.New("result exceeds 200 items; use /api/v1/items/page with limit and offset")

func readLegacyItems(read itemPageQuery, query, moduleID, filters, tags string) ([]Item, error) {
	page, err := readItemPage(read, query, moduleID, filters, tags, MaxItemPageSize, 0)
	if err != nil {
		return nil, err
	}
	if page.HasMore {
		return nil, ErrPaginationRequired
	}
	return page.Items, nil
}

const MaxItemPageSize = 200
const MaxItemPageOffset = 1000000

type ItemPage struct {
	Items   []Item `json:"items"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	HasMore bool   `json:"hasMore"`
}

func ValidateItemPage(limit, offset int) error {
	if limit < 1 || limit > MaxItemPageSize || offset < 0 || offset > MaxItemPageOffset {
		return fmt.Errorf("limit must be 1–%d and offset 0–%d", MaxItemPageSize, MaxItemPageOffset)
	}
	return nil
}

func appendItemPage(query string, args []any, limit, offset int, postgres bool) (string, []any) {
	if postgres {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	} else {
		query += " LIMIT ? OFFSET ?"
	}
	return query, append(args, limit, offset)
}

type itemPageQuery func(string, string, string, string, int, int) ([]Item, error)

func readItemPage(read itemPageQuery, query, moduleID, filters, tags string, limit, offset int) (ItemPage, error) {
	if err := ValidateItemPage(limit, offset); err != nil {
		return ItemPage{}, err
	}
	items, err := read(query, moduleID, filters, tags, limit+1, offset)
	if err != nil {
		return ItemPage{}, err
	}
	if items == nil {
		items = []Item{}
	}
	page := ItemPage{Items: items, Limit: limit, Offset: offset, HasMore: len(items) > limit}
	if page.HasMore {
		page.Items = items[:limit]
	}
	return page, nil
}
func (s *SQLiteStore) QueryItemPage(query, moduleID, filters, tags string, limit, offset int) (ItemPage, error) {
	return readItemPage(s.queryItems, query, moduleID, filters, tags, limit, offset)
}
func (s *PostgresStore) QueryItemPage(query, moduleID, filters, tags string, limit, offset int) (ItemPage, error) {
	return readItemPage(s.queryItems, query, moduleID, filters, tags, limit, offset)
}
