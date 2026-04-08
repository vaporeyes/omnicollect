// ABOUTME: Application settings persistence via the Store interface.
// ABOUTME: Wails-bound methods that delegate to the active Store implementation.
package main

// LoadSettings reads settings from the active store.
// Returns an empty JSON object if no settings exist.
func (a *App) LoadSettings() (string, error) {
	return a.store.GetSettings()
}

// SaveSettings writes settings JSON to the active store.
func (a *App) SaveSettings(settingsJSON string) error {
	return a.store.SaveSettings(settingsJSON)
}
