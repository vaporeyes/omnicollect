// ABOUTME: Store interface defining all database operations for OmniCollect.
// ABOUTME: Implemented by SQLiteStore (local) and PostgresStore (cloud).
package storage

// Item represents a single collectible record.
type Item struct {
	ID            string                 `json:"id"`
	ModuleID      string                 `json:"moduleId"`
	Title         string                 `json:"title"`
	PurchasePrice *float64               `json:"purchasePrice"`
	Images        []string               `json:"images"`
	Attributes    map[string]any         `json:"attributes"`
	CreatedAt     string                 `json:"createdAt"`
	UpdatedAt     string                 `json:"updatedAt"`
}

// ModuleSchema defines a collection type with its attribute definitions.
type ModuleSchema struct {
	ID          string            `json:"id"`
	DisplayName string            `json:"displayName"`
	Description string            `json:"description,omitempty"`
	Attributes  []AttributeSchema `json:"attributes"`
}

// AttributeSchema defines a single field within a module schema.
type AttributeSchema struct {
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	Required bool          `json:"required,omitempty"`
	Options  []string      `json:"options,omitempty"`
	Display  *DisplayHints `json:"display,omitempty"`
}

// DisplayHints controls how the frontend renders an attribute field.
type DisplayHints struct {
	Label       string `json:"label,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Widget      string `json:"widget,omitempty"`
	Group       string `json:"group,omitempty"`
	Order       int    `json:"order,omitempty"`
}

// Store defines all database operations for items, modules, and settings.
// Implementations: SQLiteStore (local mode) and PostgresStore (cloud mode).
type Store interface {
	QueryItems(query, moduleID, filtersJSON string) ([]Item, error)
	InsertItem(item Item) (Item, error)
	UpdateItem(item Item) (Item, error)
	DeleteItem(id string) error
	DeleteItems(ids []string) (int64, error)
	BulkUpdateModule(ids []string, newModuleID string) (int64, error)
	ExportItemsCSV(ids []string, modules []ModuleSchema) (string, error)
	GetModules() ([]ModuleSchema, error)
	SaveModule(schema ModuleSchema) error
	LoadModuleFile(id string) (string, error)
	GetSettings() (string, error)
	SaveSettings(json string) error
	Close() error
}
