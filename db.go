// ABOUTME: Legacy database helpers retained for backup archive support.
// ABOUTME: Core CRUD operations have moved to storage/sqlite.go and storage/postgres.go.
package main

import (
	"os"
	"path/filepath"
)

// dbFilePath returns the path to the SQLite database file.
// Retained for use by backup.go.
func dbFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "OmniCollect", "collection.db"), nil
}
