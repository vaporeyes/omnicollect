// ABOUTME: App struct that serves as the Wails binding target.
// ABOUTME: Exposes SaveItem, DeleteItem, GetItems, and GetActiveModules to the frontend.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"omnicollect/storage"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct bound to the Wails frontend.
type App struct {
	ctx        context.Context
	store      storage.Store
	mediaStore storage.MediaStore
	modules    []ModuleSchema
	config     Config
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// Init initializes the database and loads module schemas.
// Can be called from both Wails startup and standalone HTTP server.
func (a *App) Init() {
	a.InitWithConfig(LoadConfig())
}

// InitWithConfig initializes with an explicit config. Instantiates the
// appropriate Store and MediaStore based on cloud vs local settings.
func (a *App) InitWithConfig(cfg Config) {
	a.config = cfg

	// Initialize database store
	if cfg.IsCloudDB() {
		pgStore, err := storage.NewPostgresStore(cfg.DatabaseURL, cfg.TenantID)
		if err != nil {
			log.Fatalf("failed to initialize PostgreSQL: %v", err)
		}
		a.store = pgStore
	} else {
		sqliteStore, err := storage.NewSQLiteStore()
		if err != nil {
			log.Fatalf("failed to initialize SQLite: %v", err)
		}
		a.store = sqliteStore
	}

	// Initialize media store
	if cfg.IsCloudStorage() {
		s3Store, err := storage.NewS3MediaStore(cfg.S3Endpoint, cfg.S3Bucket, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Region)
		if err != nil {
			log.Fatalf("failed to initialize S3 storage: %v", err)
		}
		a.mediaStore = s3Store
	} else {
		localStore, err := storage.NewLocalMediaStore()
		if err != nil {
			log.Fatalf("failed to initialize local media storage: %v", err)
		}
		a.mediaStore = localStore
	}

	// Load modules from store
	modules, err := a.store.GetModules()
	if err != nil {
		log.Fatalf("failed to load module schemas: %v", err)
	}
	a.modules = toMainModules(modules)
	log.Printf("loaded %d module schema(s)", len(a.modules))
}

// startup is called by Wails when the application starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.Init()
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

	si := toStorageItem(item)
	var result storage.Item
	var err error
	if item.ID == "" {
		result, err = a.store.InsertItem(si)
	} else {
		result, err = a.store.UpdateItem(si)
	}
	if err != nil {
		return Item{}, err
	}
	return fromStorageItem(result), nil
}

// GetItems retrieves items with optional full-text search, module filter,
// and attribute filters. Pass empty strings to skip any filter dimension.
// filtersJSON is a JSON array of {field, op, value/values} objects.
func (a *App) GetItems(query string, moduleID string, filtersJSON string) ([]Item, error) {
	items, err := a.store.QueryItems(query, moduleID, filtersJSON)
	if err != nil {
		return nil, err
	}
	return fromStorageItems(items), nil
}

// DeleteItem removes a collection item by ID.
func (a *App) DeleteItem(id string) error {
	if id == "" {
		return fmt.Errorf("item ID is required")
	}
	return a.store.DeleteItem(id)
}

// BulkDeleteResult holds the count of deleted items.
type BulkDeleteResult struct {
	Deleted int64 `json:"deleted"`
}

// BulkUpdateResult holds the count of updated items.
type BulkUpdateResult struct {
	Updated int64 `json:"updated"`
}

// DeleteItems removes multiple items in a single atomic transaction.
func (a *App) DeleteItems(ids []string) (BulkDeleteResult, error) {
	if len(ids) == 0 {
		return BulkDeleteResult{}, fmt.Errorf("no item IDs provided")
	}
	deleted, err := a.store.DeleteItems(ids)
	if err != nil {
		return BulkDeleteResult{}, err
	}
	return BulkDeleteResult{Deleted: deleted}, nil
}

// ExportItemsCSV generates a CSV file for the given item IDs.
// Opens a save dialog for the user to pick the destination.
// Returns the file path, or empty string if cancelled.
func (a *App) ExportItemsCSV(ids []string) (string, error) {
	if len(ids) == 0 {
		return "", fmt.Errorf("no item IDs provided")
	}

	csv, err := a.store.ExportItemsCSV(ids, toStorageModules(a.modules))
	if err != nil {
		return "", err
	}

	defaultName := fmt.Sprintf("omnicollect-export-%d-items.csv", len(ids))
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Selected Items",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV Files", Pattern: "*.csv"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("save dialog: %w", err)
	}
	if path == "" {
		return "", nil
	}

	if err := os.WriteFile(path, []byte(csv), 0644); err != nil {
		return "", fmt.Errorf("writing CSV: %w", err)
	}

	return path, nil
}

// BulkUpdateModule changes the module assignment of multiple items.
func (a *App) BulkUpdateModule(ids []string, newModuleID string) (BulkUpdateResult, error) {
	if len(ids) == 0 {
		return BulkUpdateResult{}, fmt.Errorf("no item IDs provided")
	}
	if newModuleID == "" {
		return BulkUpdateResult{}, fmt.Errorf("new module ID is required")
	}
	updated, err := a.store.BulkUpdateModule(ids, newModuleID)
	if err != nil {
		return BulkUpdateResult{}, err
	}
	return BulkUpdateResult{Updated: updated}, nil
}

// GetActiveModules returns all module schemas loaded at startup.
func (a *App) GetActiveModules() ([]ModuleSchema, error) {
	return a.modules, nil
}

// ProcessImage validates and processes a local image file.
// Generates thumbnail bytes and persists via MediaStore.
func (a *App) ProcessImage(sourcePath string) (ProcessImageResult, error) {
	if sourcePath == "" {
		return ProcessImageResult{}, fmt.Errorf("source path is required")
	}

	data, err := processImageToBytes(sourcePath)
	if err != nil {
		return ProcessImageResult{}, err
	}

	// Persist via MediaStore
	if err := a.mediaStore.SaveOriginal(data.OriginalFile, data.OriginalBytes); err != nil {
		return ProcessImageResult{}, fmt.Errorf("saving original: %w", err)
	}
	if err := a.mediaStore.SaveThumbnail(data.ThumbFile, data.ThumbBytes); err != nil {
		return ProcessImageResult{}, fmt.Errorf("saving thumbnail: %w", err)
	}

	return ProcessImageResult{
		Filename:      data.Filename,
		OriginalPath:  data.OriginalFile,
		ThumbnailPath: data.ThumbFile,
		Width:         data.Width,
		Height:        data.Height,
		Format:        data.Format,
	}, nil
}

// SaveCustomModule validates and saves a module schema.
// After saving, reloads the in-memory module list so the schema is
// immediately available via GetActiveModules.
func (a *App) SaveCustomModule(schemaJSON string) (ModuleSchema, error) {
	var schema storage.ModuleSchema
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return ModuleSchema{}, fmt.Errorf("invalid JSON: %w", err)
	}

	if err := storage.ValidateModuleSchema(&schema); err != nil {
		return ModuleSchema{}, err
	}

	if err := a.store.SaveModule(schema); err != nil {
		return ModuleSchema{}, err
	}

	// Reload all module schemas
	modules, err := a.store.GetModules()
	if err != nil {
		log.Printf("warning: failed to reload modules after save: %v", err)
	} else {
		a.modules = toMainModules(modules)
	}

	return fromStorageModule(schema), nil
}

// LoadModuleFile reads a module schema and returns its raw JSON content
// for editing in the schema builder.
func (a *App) LoadModuleFile(moduleID string) (string, error) {
	if moduleID == "" {
		return "", fmt.Errorf("module ID is required")
	}
	return a.store.LoadModuleFile(moduleID)
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

	// Backup only works with SQLiteStore (local mode)
	sqliteStore, ok := a.store.(*storage.SQLiteStore)
	if !ok {
		return "", fmt.Errorf("backup export is only available in local mode")
	}

	if err := createBackupArchive(path, sqliteStore.DB()); err != nil {
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

// --- Type conversion helpers between main.* and storage.* types ---

func toStorageItem(item Item) storage.Item {
	return storage.Item{
		ID:            item.ID,
		ModuleID:      item.ModuleID,
		Title:         item.Title,
		PurchasePrice: item.PurchasePrice,
		Images:        item.Images,
		Attributes:    item.Attributes,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func fromStorageItem(item storage.Item) Item {
	return Item{
		ID:            item.ID,
		ModuleID:      item.ModuleID,
		Title:         item.Title,
		PurchasePrice: item.PurchasePrice,
		Images:        item.Images,
		Attributes:    item.Attributes,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func fromStorageItems(items []storage.Item) []Item {
	result := make([]Item, len(items))
	for i, item := range items {
		result[i] = fromStorageItem(item)
	}
	return result
}

func toStorageModule(m ModuleSchema) storage.ModuleSchema {
	attrs := make([]storage.AttributeSchema, len(m.Attributes))
	for i, a := range m.Attributes {
		attrs[i] = storage.AttributeSchema{
			Name:     a.Name,
			Type:     a.Type,
			Required: a.Required,
			Options:  a.Options,
		}
		if a.Display != nil {
			attrs[i].Display = &storage.DisplayHints{
				Label:       a.Display.Label,
				Placeholder: a.Display.Placeholder,
				Widget:      a.Display.Widget,
				Group:       a.Display.Group,
				Order:       a.Display.Order,
			}
		}
	}
	return storage.ModuleSchema{
		ID:          m.ID,
		DisplayName: m.DisplayName,
		Description: m.Description,
		Attributes:  attrs,
	}
}

func toStorageModules(modules []ModuleSchema) []storage.ModuleSchema {
	result := make([]storage.ModuleSchema, len(modules))
	for i, m := range modules {
		result[i] = toStorageModule(m)
	}
	return result
}

func fromStorageModule(m storage.ModuleSchema) ModuleSchema {
	attrs := make([]AttributeSchema, len(m.Attributes))
	for i, a := range m.Attributes {
		attrs[i] = AttributeSchema{
			Name:     a.Name,
			Type:     a.Type,
			Required: a.Required,
			Options:  a.Options,
		}
		if a.Display != nil {
			attrs[i].Display = &DisplayHints{
				Label:       a.Display.Label,
				Placeholder: a.Display.Placeholder,
				Widget:      a.Display.Widget,
				Group:       a.Display.Group,
				Order:       a.Display.Order,
			}
		}
	}
	return ModuleSchema{
		ID:          m.ID,
		DisplayName: m.DisplayName,
		Description: m.Description,
		Attributes:  attrs,
	}
}

func toMainModules(modules []storage.ModuleSchema) []ModuleSchema {
	result := make([]ModuleSchema, len(modules))
	for i, m := range modules {
		result[i] = fromStorageModule(m)
	}
	return result
}
