// ABOUTME: HTTP recovery contracts and a simulated lost deletion acknowledgement.
// ABOUTME: Validates persistent discoverability, explicit disposal, and no unsafe overwrite.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"omnicollect/storage"
	"strings"
	"testing"
)

type lostDeleteAckStore struct{ storage.Store }

func (s *lostDeleteAckStore) WithContext(ctx context.Context) storage.Store {
	return &lostDeleteAckStore{s.Store.WithContext(ctx)}
}
func (s *lostDeleteAckStore) DeleteWithRecovery(ids []string) (storage.DeletionBatch, error) {
	batch, err := s.Store.DeleteWithRecovery(ids)
	if err != nil {
		return batch, err
	}
	return batch, fmt.Errorf("acknowledgement lost")
}
func (s *lostDeleteAckStore) RecoverDeletion(id string) (int, error) {
	count, err := s.Store.RecoverDeletion(id)
	if err != nil {
		return count, err
	}
	return count, fmt.Errorf("recovery acknowledgement lost")
}
func TestRecoveryHTTP(t *testing.T) {
	base, app := newTestServer(t)
	if err := app.store.SaveModule(storage.ModuleSchema{ID: "recover", DisplayName: "Recover"}); err != nil {
		t.Fatal(err)
	}
	send := func(method, path, body string, status int) (http.Header, []byte) {
		t.Helper()
		r, err := http.NewRequest(method, base+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(r)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		raw, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != status {
			t.Fatalf("%s %s: %d %s", method, path, response.StatusCode, raw)
		}
		return response.Header, raw
	}
	_, raw := send("GET", "/api/v1/recovery", "", 200)
	if strings.TrimSpace(string(raw)) != "[]" {
		t.Fatalf("empty %s", raw)
	}
	item, err := app.store.SaveItem(context.Background(), storage.Item{ModuleID: "recover", Title: "One"})
	if err != nil {
		t.Fatal(err)
	}
	header, _ := send("DELETE", "/api/v1/items/"+item.ID, "", 204)
	id := header.Get("X-Recovery-ID")
	if id == "" {
		t.Fatal("missing recovery header")
	}
	send("POST", "/api/v1/recovery/"+id, "{}", 200)
	send("POST", "/api/v1/recovery/"+id, "{}", 404)
	_, raw = send("POST", "/api/v1/items/batch-delete", fmt.Sprintf("{\"ids\":[%q]}", item.ID), 200)
	var batch storage.DeletionBatch
	if err = json.Unmarshal(raw, &batch); err != nil {
		t.Fatal(err)
	}
	if batch.Deleted != 1 || batch.RecoveryID == "" {
		t.Fatalf("batch %+v", batch)
	}
	send("DELETE", "/api/v1/recovery/"+batch.RecoveryID, "", 204)
	send("POST", "/api/v1/recovery/"+batch.RecoveryID, "{}", 404)
	item, err = app.store.SaveItem(context.Background(), storage.Item{ModuleID: "recover", Title: "Lost ack"})
	if err != nil {
		t.Fatal(err)
	}
	app.store = &lostDeleteAckStore{app.store}
	_, raw = send("POST", "/api/v1/items/batch-delete", fmt.Sprintf("{\"ids\":[%q]}", item.ID), 500)
	if !strings.Contains(string(raw), "not confirmed") {
		t.Fatalf("dishonest failure %s", raw)
	}
	_, raw = send("GET", "/api/v1/recovery", "", 200)
	var batches []storage.DeletionBatch
	if err = json.Unmarshal(raw, &batches); err != nil {
		t.Fatal(err)
	}
	if len(batches) != 1 || batches[0].Titles[0] != "Lost ack" {
		t.Fatalf("lost recovery handle: %s", raw)
	}
	_, raw = send("POST", "/api/v1/recovery/"+batches[0].RecoveryID, "{}", 409)
	if !strings.Contains(string(raw), "not confirmed") {
		t.Fatalf("dishonest recovery failure %s", raw)
	}
	_, raw = send("GET", "/api/v1/recovery", "", 200)
	if strings.TrimSpace(string(raw)) != "[]" {
		t.Fatalf("committed recovery retained batch %s", raw)
	}
	summary, err := app.store.CollectionSummary()
	if err != nil || summary.Items != 1 {
		t.Fatalf("lost restored record %+v %v", summary, err)
	}
}
