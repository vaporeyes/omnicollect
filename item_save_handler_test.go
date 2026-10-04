// ABOUTME: HTTP save validation, strict JSON, and optimistic conflict response tests.
// ABOUTME: Verifies stale edits never overwrite committed items through the public API.
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestItemSaveHTTPVersionsAndValidation(t *testing.T) {
	url, _ := newTestServer(t)
	post := func(body []byte, status int) Item {
		t.Helper()
		response, err := http.Post(url+"/api/v1/items", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var result Item
		if response.StatusCode != status {
			var problem any
			json.NewDecoder(response.Body).Decode(&problem)
			t.Fatalf("got %d want %d: %v", response.StatusCode, status, problem)
		}
		if status == 200 {
			if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
				t.Fatal(err)
			}
		}
		return result
	}
	original := post([]byte(`{"moduleId":"comics","title":"First","attributes":{"legacy":"keep"}}`), 200)
	noVersion := original
	noVersion.UpdatedAt = ""
	data, _ := json.Marshal(noVersion)
	post(data, 428)
	update := original
	update.Title = "Second"
	data, _ = json.Marshal(update)
	saved := post(data, 200)
	if saved.UpdatedAt == original.UpdatedAt || saved.CreatedAt != original.CreatedAt {
		t.Fatal("version/history not preserved")
	}
	data, _ = json.Marshal(original)
	post(data, 409)
	post([]byte(`{"moduleId":"comics","title":"Valid"} {"title":"trailing"}`), 400)
	post([]byte(`{"moduleId":"missing","title":"Invalid"}`), 400)
	post([]byte(`{"moduleId":"comics","title":"   "}`), 400)
	post([]byte(`{"moduleId":"comics","title":"Invalid","images":["another-tenant.png"]}`), 400)
}
