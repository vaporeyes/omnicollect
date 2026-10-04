// ABOUTME: Public gallery routing uses bounded projections and strict page parameters.
// ABOUTME: Spy store fails on old full-item/module-list reads; HTML remains minimal.
package showcase

import (
	"context"
	"fmt"
	"net/http/httptest"
	"omnicollect/storage"
	"strings"
	"testing"
)

type gallerySpy struct {
	storage.Store
	calls int
	page  int
}

func (s *gallerySpy) WithContext(context.Context) storage.Store { return s }
func (s *gallerySpy) GetShowcaseBySlug(string) (*storage.Showcase, error) {
	return &storage.Showcase{Slug: "public", ModuleID: "gallery", Enabled: true}, nil
}
func (s *gallerySpy) QueryItems(string, string, string, string) ([]storage.Item, error) {
	panic("unbounded item read")
}
func (s *gallerySpy) GetModules() ([]storage.ModuleSchema, error) { panic("unbounded module read") }
func (s *gallerySpy) ReadGalleryPage(moduleID string, page int) (storage.GalleryPage, error) {
	s.calls++
	s.page = page
	return storage.GalleryPage{CollectionName: "Gallery", TotalItems: 25, Page: 2, TotalPages: 2, Items: []storage.GalleryRecord{{ID: "id", Title: "Visible title", PrimaryImage: "cover.png"}}}, nil
}
func TestGalleryBoundedRouting(t *testing.T) {
	store := &gallerySpy{}
	handler := HandleShowcase(store)
	for _, query := range []string{"page=0", "page=-1", "page=abc", "page=1&page=2", "page=%ZZ", "page=1;sort=x", "sort=id", fmt.Sprintf("page=%d", storage.MaxGalleryPages+1)} {
		response := httptest.NewRecorder()
		handler(response, httptest.NewRequest("GET", "/showcase/public?"+query, nil))
		if response.Code != 400 {
			t.Fatalf("%s: %d", query, response.Code)
		}
	}
	if store.calls != 0 {
		t.Fatal("queried invalid pages")
	}
	response := httptest.NewRecorder()
	handler(response, httptest.NewRequest("GET", "/showcase/public?page=2", nil))
	if response.Code != 200 || store.calls != 1 || store.page != 2 || !strings.Contains(response.Body.String(), "Visible title") || !strings.Contains(response.Body.String(), "Page 2 of 2") {
		t.Fatalf("bounded gallery %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	if err := RenderGallery(response, GalleryData{Truncated: true, TotalPages: storage.MaxGalleryPages, Page: storage.MaxGalleryPages}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(response.Body.String(), "beyond this browsing limit") {
		t.Fatal("hidden truncation")
	}
}
