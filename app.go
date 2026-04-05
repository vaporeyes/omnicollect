// ABOUTME: App struct that serves as the Wails binding target.
// ABOUTME: Exposes SaveItem, GetItems, and GetActiveModules to the frontend.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
