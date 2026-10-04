// ABOUTME: Authenticated private media and live, reference-checked public showcase delivery.
// ABOUTME: Exercises tenant-scoped upload/read/AI paths without real external services.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"omnicollect/auth"
	"omnicollect/storage"
	"omnicollect/tenantid"
)

type recordingAI struct {
	calls int
	mime  string
}

func (p *recordingAI) AnalyzeImage(ctx context.Context, image, mime, prompt string) (string, error) {
	p.calls++
	p.mime = mime
	return `{"title":"Example","attributes":{}}`, nil
}

func TestPrivateMediaRoutesRequireAuthentication(t *testing.T) {
	app := testApp(t)
	app.config = Config{AuthIssuer: "https://issuer.example/", AuthAudience: "test"}
	handler := NewServer(app).buildHandler()
	for _, method := range []string{"GET", "HEAD"} {
		for _, path := range []string{"/originals/secret.png", "/thumbnails/secret.png", "/originals/", "/thumbnails/"} {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(method, path, nil))
			if response.Code != 401 {
				t.Fatalf("anonymous media %s %s returned %d", method, path, response.Code)
			}
		}
	}
}
func TestTenantMediaUploadReadAndAIIsolation(t *testing.T) {
	app := testApp(t)
	app.config = Config{AuthIssuer: "https://issuer.example/", AuthAudience: "test"}
	if err := app.store.SaveModule(storage.ModuleSchema{ID: "books", DisplayName: "Books"}); err != nil {
		t.Fatal(err)
	}
	provider := &recordingAI{}
	app.aiProvider = provider
	server := NewServer(app)
	handler := server.tenantScopeMiddleware(server.mux)
	request := func(tenant string, req *http.Request) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req.WithContext(auth.SetTenantID(req.Context(), tenantid.Local(tenant))))
		return response
	}
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	file, err := form.CreateFormFile("image", "misnamed.jpg")
	if err != nil {
		t.Fatal(err)
	}
	file.Write(pngFixture(t, 100, 50))
	form.Close()
	req := httptest.NewRequest("POST", "/api/v1/images/upload", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	response := request("alice", req)
	if response.Code != 200 {
		t.Fatalf("upload failed: %d %s", response.Code, response.Body.String())
	}
	var image ProcessImageResult
	if err := json.Unmarshal(response.Body.Bytes(), &image); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(image.Filename, ".png") {
		t.Fatal("original MIME extension lost")
	}
	for _, kind := range []string{"originals", "thumbnails"} {
		response = request("alice", httptest.NewRequest("GET", "/"+kind+"/"+image.Filename, nil))
		expectedType := "image/png"
		if kind == "thumbnails" {
			expectedType = "image/jpeg"
		}
		if response.Code != 200 || response.Header().Get("Content-Type") != expectedType || response.Header().Get("Cache-Control") != "private, no-store" {
			t.Fatalf("bad media response: %d %v", response.Code, response.Header())
		}
		response = request("bob", httptest.NewRequest("GET", "/"+kind+"/"+image.Filename, nil))
		if response.Code != 404 {
			t.Fatalf("cross-tenant media exposed: %d", response.Code)
		}
	}
	aiBody, _ := json.Marshal(map[string]string{"imageFilename": image.Filename, "moduleId": "books"})
	response = request("bob", httptest.NewRequest("POST", "/api/v1/ai/analyze", bytes.NewReader(aiBody)))
	if response.Code != 404 || provider.calls != 0 {
		t.Fatal("AI could read another tenant's image")
	}
	response = request("alice", httptest.NewRequest("POST", "/api/v1/ai/analyze", bytes.NewReader(aiBody)))
	if response.Code != 200 || provider.calls != 1 || provider.mime != "image/png" {
		t.Fatalf("AI MIME/ownership failure: %d %+v", response.Code, provider)
	}
	if _, err := app.mediaStore.GetOriginal(context.Background(), image.Filename); err == nil {
		t.Fatal("tenant upload leaked to base namespace")
	}
}
func TestPublicMediaRequiresLiveShowcaseReference(t *testing.T) {
	app, item := seedBackupApp(t)
	price := 12345.67
	item.PurchasePrice = &price
	item.Images = append(item.Images, "private.png")
	item.Tags = []string{"private-tag"}
	item.Attributes = map[string]any{"notes": "private-note"}
	if _, err := app.store.UpdateItem(item); err != nil {
		t.Fatal(err)
	}
	if err := app.store.SaveModule(storage.ModuleSchema{ID: "books", DisplayName: "Books", Attributes: []storage.AttributeSchema{{Name: "notes", Type: "string"}}}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	sc := storage.Showcase{ID: "gallery", Slug: "books-public", TenantID: "local", ModuleID: "books", Enabled: true, CreatedAt: now, UpdatedAt: now}
	if err := app.store.UpsertShowcase(sc); err != nil {
		t.Fatal(err)
	}
	if err := app.mediaStore.SaveOriginal(context.Background(), "private.png", pngFixture(t, 1, 1)); err != nil {
		t.Fatal(err)
	}
	handler := NewServer(app).buildHandler()
	get := func(path string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		return response
	}
	mediaPath := "/showcase/books-public/media/originals/image.jpg"
	if response := get(mediaPath); response.Code != 200 || response.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("public referenced image unavailable: %d", response.Code)
	}
	if response := get("/showcase/books-public"); !strings.Contains(response.Body.String(), "/showcase/books-public/media/thumbnails/image.jpg") {
		t.Fatal("gallery still points at private media")
	}
	gallery := get("/showcase/books-public").Body.String()
	for _, private := range []string{"12345.67", "private-tag", "private-note"} {
		if strings.Contains(gallery, private) {
			t.Fatalf("private field leaked: %s", private)
		}
	}
	if response := get("/showcase/books-public/media/originals/private.png"); response.Code != 404 {
		t.Fatal("unreferenced original exposed")
	}
	sc.Enabled = false
	if err := app.store.UpsertShowcase(sc); err != nil {
		t.Fatal(err)
	}
	if response := get(mediaPath); response.Code != 404 {
		t.Fatal("disabled gallery image still public")
	}
	sc.Enabled = true
	if err := app.store.UpsertShowcase(sc); err != nil {
		t.Fatal(err)
	}
	if err := app.store.DeleteItem(item.ID); err != nil {
		t.Fatal(err)
	}
	if response := get(mediaPath); response.Code != 404 {
		t.Fatal("removed gallery image still public")
	}
}
