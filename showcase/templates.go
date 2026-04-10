// ABOUTME: Embeds and renders HTML templates for public showcase gallery pages.
// ABOUTME: Uses go:embed for binary-embedded templates and html/template for rendering.
package showcase

import (
	"embed"
	"html/template"
	"io"
)

//go:embed templates/*.html
var templateFS embed.FS

var (
	galleryTmpl     = template.Must(template.ParseFS(templateFS, "templates/gallery.html"))
	unavailableTmpl = template.Must(template.ParseFS(templateFS, "templates/unavailable.html"))
)

// GalleryItem holds the data for a single item card in the gallery template.
type GalleryItem struct {
	ID           string
	Title        string
	PrimaryImage string
	KeyAttrs     []string
	AllAttrs     []AttrPair
	Tags         []string
}

// AttrPair holds a label-value pair for the detail overlay.
type AttrPair struct {
	Label string
	Value string
}

// GalleryData holds all data passed to the gallery template.
type GalleryData struct {
	CollectionName string
	TotalItems     int
	Items          []GalleryItem
	Page           int
	TotalPages     int
	PrevPage       int
	NextPage       int
}

// RenderGallery writes the gallery HTML page to w.
func RenderGallery(w io.Writer, data GalleryData) error {
	return galleryTmpl.Execute(w, data)
}

// RenderUnavailable writes the "no longer available" page to w.
func RenderUnavailable(w io.Writer) error {
	return unavailableTmpl.Execute(w, nil)
}
