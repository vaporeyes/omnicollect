// ABOUTME: Regression tests for opaque import handles and safe retry semantics.
// ABOUTME: Verifies tenant ownership, exclusive execution, and path rejection.
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportSessionOwnershipAndRetry(t *testing.T) {
	sessions := importSessions{}
	path := filepath.Join(t.TempDir(), "backup.zip")
	if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	id, err := sessions.register(path, "owner")
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][2]string{{"../backup.zip", "owner"}, {id, "other"}} {
		if _, err := sessions.claim(args[0], args[1]); err == nil {
			t.Fatal("accepted unauthorized handle")
		}
	}
	if _, err := sessions.claim(id, "owner"); err != nil {
		t.Fatal(err)
	}
	if _, err := sessions.claim(id, "owner"); err == nil {
		t.Fatal("accepted duplicate execution")
	}
	sessions.release(id, false)
	if _, err := os.Stat(path); err != nil {
		t.Fatal("failed import removed retry archive", err)
	}
	if _, err := sessions.claim(id, "owner"); err != nil {
		t.Fatal(err)
	}
	sessions.release(id, true)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("successful import retained archive")
	}
}
