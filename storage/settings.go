// ABOUTME: Shared settings contract for SQLite and PostgreSQL.
// ABOUTME: Settings updates replace supplied top-level keys and preserve unrelated keys.
package storage

import (
	"encoding/json"
	"fmt"
)

func settingsObject(raw string) (map[string]json.RawMessage, error) {
	if len(raw) > 1<<20 {
		return nil, fmt.Errorf("settings exceed 1 MB limit")
	}
	var value map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &value); err != nil || value == nil {
		return nil, fmt.Errorf("settings must be a JSON object")
	}
	return value, nil
}

func mergeSettings(current, patch string) (string, error) {
	base, err := settingsObject(current)
	if err != nil {
		return "", err
	}
	changes, err := settingsObject(patch)
	if err != nil {
		return "", err
	}
	for key, value := range changes {
		base[key] = value
	}
	data, err := json.Marshal(base)
	if len(data) > 1<<20 {
		return "", fmt.Errorf("merged settings exceed 1 MB limit")
	}
	return string(data), err
}
