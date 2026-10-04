// ABOUTME: Explicit bounded-page and tenant-wide-summary HTTP read contracts.
// ABOUTME: Additive endpoints leave legacy array consumers intact until UI migration.
package main

import (
	"net/http"
	"net/url"
	"omnicollect/storage"
	"strconv"
)

func (s *Server) handleItemPage(w http.ResponseWriter, r *http.Request) {
	values, parseErr := url.ParseQuery(r.URL.RawQuery)
	if parseErr != nil {
		writeError(w, 400, "invalid query encoding")
		return
	}
	for key, entries := range values {
		switch key {
		case "query", "moduleId", "filters", "tags", "limit", "offset":
		default:
			writeError(w, 400, "unsupported page parameter")
			return
		}
		if len(entries) != 1 {
			writeError(w, 400, "duplicate page parameter")
			return
		}
	}
	limit, offset := 100, 0
	for key, target := range map[string]*int{"limit": &limit, "offset": &offset} {
		if entries, ok := values[key]; ok {
			value, err := strconv.Atoi(entries[0])
			if err != nil {
				writeError(w, 400, "invalid "+key)
				return
			}
			*target = value
		}
	}
	if err := storage.ValidateItemPage(limit, offset); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	query, moduleID, filters, tags := values.Get("query"), values.Get("moduleId"), values.Get("filters"), values.Get("tags")
	if err := storage.ValidateQuery(query, moduleID, filters, tags); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	page, err := s.requestStore(r).QueryItemPage(query, moduleID, filters, tags, limit, offset)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, page)
}
func (s *Server) handleCollectionSummary(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery != "" {
		writeError(w, 400, "summary is tenant-wide and accepts no filters or page parameters")
		return
	}
	summary, err := s.requestStore(r).CollectionSummary()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, summary)
}
