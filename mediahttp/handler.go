// ABOUTME: Confined image delivery shared by private API and explicitly authorized showcases.
// ABOUTME: Detects MIME from content, rejects non-images, and prevents public cache leakage.
package mediahttp

import (
	"bytes"
	"net/http"
	"omnicollect/storage"
	"time"
)

// Serve requires the caller to authorize the namespace and (for public media)
// the live showcase reference before calling. Filenames never select a tenant.
func Serve(w http.ResponseWriter, r *http.Request, media storage.MediaStore, kind, filename string) {
	if err := storage.ValidateFilename(filename); err != nil {
		http.NotFound(w, r)
		return
	}
	var data []byte
	var err error
	switch kind {
	case "originals":
		data, err = media.GetOriginal(r.Context(), filename)
	case "thumbnails":
		data, err = media.GetThumbnail(r.Context(), filename)
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.NotFound(w, r)
		return
	}
	contentType := http.DetectContentType(data)
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, no-store")
	http.ServeContent(w, r, filename, time.Time{}, bytes.NewReader(data))
}
