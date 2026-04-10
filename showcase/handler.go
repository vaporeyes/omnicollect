// ABOUTME: HTTP handler for public showcase gallery pages served at /showcase/{slug}.
// ABOUTME: Looks up showcase by slug, loads items and schema, renders server-side HTML.
package showcase

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"omnicollect/storage"
)

const itemsPerPage = 24

// HandleShowcase returns an http.HandlerFunc that serves public gallery pages.
// The store is used for showcase lookup and item queries. The handler never
// exposes tenant IDs, owner identity, or application navigation.
func HandleShowcase(store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse slug from the URL path: /showcase/{slug}
		slug := r.PathValue("slug")
		if slug == "" {
			// Fallback: extract from path manually
			slug = strings.TrimPrefix(r.URL.Path, "/showcase/")
		}
		if slug == "" {
			w.WriteHeader(http.StatusOK)
			RenderUnavailable(w)
			return
		}

		sc, err := store.GetShowcaseBySlug(slug)
		if err != nil || sc == nil || !sc.Enabled {
			w.WriteHeader(http.StatusOK)
			RenderUnavailable(w)
			return
		}

		// For PostgresStore, temporarily set the search path to the showcase's tenant
		if pgStore, ok := store.(*storage.PostgresStore); ok {
			pgStore.SetTenantSchema(sc.TenantID)
		}

		// Load items for this module
		items, err := store.QueryItems("", sc.ModuleID, "", "")
		if err != nil {
			w.WriteHeader(http.StatusOK)
			RenderUnavailable(w)
			return
		}

		// Load module schema for attribute labels
		var schema *storage.ModuleSchema
		modules, err := store.GetModules()
		if err == nil {
			for _, m := range modules {
				if m.ID == sc.ModuleID {
					schema = &m
					break
				}
			}
		}

		// Determine collection name from schema
		collectionName := sc.ModuleID
		if schema != nil {
			collectionName = schema.DisplayName
		}

		totalItems := len(items)

		// Pagination
		page := 1
		if p := r.URL.Query().Get("page"); p != "" {
			if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
				page = parsed
			}
		}
		totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage
		if totalPages < 1 {
			totalPages = 1
		}
		if page > totalPages {
			page = totalPages
		}

		start := (page - 1) * itemsPerPage
		end := start + itemsPerPage
		if end > totalItems {
			end = totalItems
		}

		pageItems := items[start:end]

		// Build gallery items with attribute display
		galleryItems := make([]GalleryItem, len(pageItems))
		for i, item := range pageItems {
			gi := GalleryItem{
				ID:    item.ID,
				Title: item.Title,
				Tags:  item.Tags,
			}

			if len(item.Images) > 0 {
				gi.PrimaryImage = item.Images[0]
			}

			// Build key attributes (first 2 for card) and all attributes (for detail)
			if schema != nil {
				for j, attr := range schema.Attributes {
					val, ok := item.Attributes[attr.Name]
					if !ok || val == nil || fmt.Sprintf("%v", val) == "" {
						continue
					}
					label := attr.Name
					if attr.Display != nil && attr.Display.Label != "" {
						label = attr.Display.Label
					}
					valStr := fmt.Sprintf("%v", val)
					gi.AllAttrs = append(gi.AllAttrs, AttrPair{Label: label, Value: valStr})
					if j < 2 {
						gi.KeyAttrs = append(gi.KeyAttrs, valStr)
					}
				}
			}

			// Add purchase price if present
			if item.PurchasePrice != nil {
				gi.AllAttrs = append(gi.AllAttrs, AttrPair{
					Label: "Price",
					Value: fmt.Sprintf("$%.2f", *item.PurchasePrice),
				})
			}

			galleryItems[i] = gi
		}

		data := GalleryData{
			CollectionName: collectionName,
			TotalItems:     totalItems,
			Items:          galleryItems,
			Page:           page,
			TotalPages:     totalPages,
			PrevPage:       page - 1,
			NextPage:       page + 1,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		RenderGallery(w, data)
	}
}
