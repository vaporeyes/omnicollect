// ABOUTME: Module directory path helper retained for backup archive support.
// ABOUTME: Core module CRUD operations have moved to storage/sqlite.go and storage/postgres.go.
package main

import (
	"os"
	"path/filepath"
)

// modulesDir returns the path to the modules directory.
// Retained for use by backup.go.
func modulesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".omnicollect", "modules"), nil
}
