// ABOUTME: HTTP handler for public showcase gallery pages served at /showcase/{slug}.
// ABOUTME: Looks up showcase by slug, loads items and schema, renders server-side HTML.
package showcase

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"omnicollect/mediahttp"
	"omnicollect/storage"
)

// HandleMedia exposes only media referenced by this currently enabled gallery.
func HandleMedia(store storage.Store, base storage.MediaStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := store.WithContext(r.Context())
		sc, err := store.GetShowcaseBySlug(r.PathValue("slug"))
		if err != nil || sc == nil || !sc.Enabled {
			http.NotFound(w, r)
			return
		}
		scoped := store
		media := base
		if pg, ok := store.(*storage.PostgresStore); ok {
			scoped = pg.WithTenantSchema(sc.TenantID)
			media, err = base.ForTenant(sc.TenantID)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		name := r.PathValue("filename")
		if err := storage.ValidateFilename(name); err != nil {
			http.NotFound(w, r)
			return
		}
		allowed, err := scoped.HasShowcaseImage(r.Context(), sc.ModuleID, name)
		if err != nil || !allowed {
			http.NotFound(w, r)
			return
		}
		mediahttp.Serve(w, r, media, r.PathValue("kind"), name)
	}
}

// HandleShowcase returns an http.HandlerFunc that serves public gallery pages.
// The store is used for showcase lookup and item queries. The handler never
// exposes tenant IDs, owner identity, or application navigation.
func HandleShowcase(store storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := store.WithContext(r.Context())
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; base-uri 'none'; frame-ancestors 'none'")
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

		// Use a per-request scoped store for the showcase's tenant to avoid
		// racing with concurrent requests on shared store state.
		tenantStore := store
		if pgStore, ok := store.(*storage.PostgresStore); ok {
			tenantStore = pgStore.WithTenantSchema(sc.TenantID)
		}

		params, err := url.ParseQuery(r.URL.RawQuery)
		page := 1
		if err == nil {
			for key, values := range params {
				if key != "page" || len(values) != 1 {
					err = strconv.ErrSyntax
					break
				}
				page, err = strconv.Atoi(values[0])
				if err != nil {
					break
				}
			}
		}
		if err != nil || page < 1 || page > storage.MaxGalleryPages {
			http.Error(w, "Invalid gallery page", http.StatusBadRequest)
			return
		}
		result, err := tenantStore.ReadGalleryPage(sc.ModuleID, page)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			RenderUnavailable(w)
			return
		}
		galleryItems := make([]GalleryItem, len(result.Items))
		for i, item := range result.Items {
			galleryItems[i] = GalleryItem{ID: item.ID, Title: item.Title, PrimaryImage: item.PrimaryImage}
		}
		data := GalleryData{
			Slug: sc.Slug, CollectionName: result.CollectionName, TotalItems: result.TotalItems,
			Items: galleryItems, Page: result.Page, TotalPages: result.TotalPages,
			PrevPage: result.Page - 1, NextPage: result.Page + 1, Truncated: result.Truncated,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		RenderGallery(w, data)
	}
}
