// ABOUTME: Application settings persistence for theme and user preferences.
// ABOUTME: Reads/writes JSON settings file at the user config directory.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// settingsFilePath returns the path to the settings JSON file.
func settingsFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "OmniCollect", "settings.json"), nil
}

// LoadSettings reads the settings file and returns the raw JSON string.
// Returns an empty JSON object if the file does not exist.
func (a *App) LoadSettings() (string, error) {
	path, err := settingsFilePath()
	if err != nil {
		return "{}", fmt.Errorf("resolving settings path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "{}", nil
		}
		return "{}", fmt.Errorf("reading settings: %w", err)
	}

	return string(data), nil
}

// SaveSettings writes the settings JSON string to disk.
func (a *App) SaveSettings(settingsJSON string) error {
	// Validate JSON
	var check json.RawMessage
	if err := json.Unmarshal([]byte(settingsJSON), &check); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	path, err := settingsFilePath()
	if err != nil {
		return fmt.Errorf("resolving settings path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating settings directory: %w", err)
	}

	// Pretty-print for readability
	var formatted json.RawMessage
	if err := json.Unmarshal([]byte(settingsJSON), &formatted); err == nil {
		pretty, err := json.MarshalIndent(formatted, "", "  ")
		if err == nil {
			settingsJSON = string(pretty)
		}
	}

	if err := os.WriteFile(path, []byte(settingsJSON), 0644); err != nil {
		return fmt.Errorf("writing settings: %w", err)
	}

	return nil
}
