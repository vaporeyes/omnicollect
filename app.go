// ABOUTME: App struct that serves as the Wails binding target.
// ABOUTME: Exposes SaveItem, GetItems, and GetActiveModules to the frontend.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct bound to the Wails frontend.
type App struct {
	ctx     context.Context
	db      *sql.DB
	modules []ModuleSchema
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// startup is called by Wails when the application starts.
// It initializes the database and loads module schemas.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	db, err := initDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	a.db = db

	modules, err := loadModuleSchemas()
	if err != nil {
		log.Fatalf("failed to load module schemas: %v", err)
	}
	a.modules = modules
	log.Printf("loaded %d module schema(s)", len(a.modules))
}

// SaveItem creates or updates a collection item.
// If item.ID is empty, a new item is created with a generated UUID.
// If item.ID exists in the database, the item is updated.
func (a *App) SaveItem(item Item) (Item, error) {
	if item.ModuleID == "" {
		return Item{}, fmt.Errorf("module_id is required")
	}
	if item.Title == "" {
		return Item{}, fmt.Errorf("title is required")
	}

	if item.Images == nil {
		item.Images = []string{}
	}
	if item.Attributes == nil {
		item.Attributes = map[string]any{}
	}

	if item.ID == "" {
		return insertItem(a.db, item)
	}
	return updateItem(a.db, item)
}

// GetItems retrieves items with optional full-text search and module filter.
// Pass empty strings to skip filtering.
func (a *App) GetItems(query string, moduleID string) ([]Item, error) {
	return queryItems(a.db, query, moduleID)
}

// GetActiveModules returns all module schemas loaded at startup.
func (a *App) GetActiveModules() ([]ModuleSchema, error) {
	return a.modules, nil
}

// ProcessImage validates and processes a local image file.
// Copies the original and generates a 300x300 JPEG thumbnail.
func (a *App) ProcessImage(sourcePath string) (ProcessImageResult, error) {
	if sourcePath == "" {
		return ProcessImageResult{}, fmt.Errorf("source path is required")
	}
	return processImage(sourcePath)
}

// SaveCustomModule validates and writes a module schema JSON file to disk.
// After saving, reloads the in-memory module list so the schema is
// immediately available via GetActiveModules.
func (a *App) SaveCustomModule(schemaJSON string) (ModuleSchema, error) {
	var schema ModuleSchema
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return ModuleSchema{}, fmt.Errorf("invalid JSON: %w", err)
	}

	if err := validateModuleSchema(&schema); err != nil {
		return ModuleSchema{}, err
	}

	if err := saveModuleFile(&schema); err != nil {
		return ModuleSchema{}, err
	}

	// Reload all module schemas
	modules, err := loadModuleSchemas()
	if err != nil {
		log.Printf("warning: failed to reload modules after save: %v", err)
	} else {
		a.modules = modules
	}

	return schema, nil
}

// LoadModuleFile reads a module schema file from disk and returns its
// raw JSON content for editing in the schema builder.
func (a *App) LoadModuleFile(moduleID string) (string, error) {
	if moduleID == "" {
		return "", fmt.Errorf("module ID is required")
	}

	path, err := findModuleFile(moduleID)
	if err != nil {
		return "", fmt.Errorf("finding module file: %w", err)
	}
	if path == "" {
		return "", fmt.Errorf("module not found: %s", moduleID)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading module file: %w", err)
	}

	return string(data), nil
}

// ExportBackup creates a ZIP archive containing the database, media,
// and module schemas. Opens a save dialog for the user to choose the
// output location. Returns the path to the created archive.
func (a *App) ExportBackup() (string, error) {
	defaultName := fmt.Sprintf("omnicollect-backup-%s.zip",
		time.Now().UTC().Format("20060102-150405"))

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Backup",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "ZIP Archives", Pattern: "*.zip"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("save dialog: %w", err)
	}
	if path == "" {
		return "", nil // user cancelled
	}

	if err := createBackupArchive(path, a.db); err != nil {
		return "", fmt.Errorf("creating backup: %w", err)
	}

	return path, nil
}

// SelectImageFile opens a native file dialog for selecting an image.
// Returns the selected file path, or empty string if cancelled.
func (a *App) SelectImageFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Image",
		Filters: []runtime.FileFilter{
			{DisplayName: "Images", Pattern: "*.jpg;*.jpeg;*.png;*.gif;*.webp"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("file dialog: %w", err)
	}
	return path, nil
}
