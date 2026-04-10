// ABOUTME: Store interface defining all database operations for OmniCollect.
// ABOUTME: Implemented by SQLiteStore (local) and PostgresStore (cloud).
package storage

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
)

// Item represents a single collectible record.
type Item struct {
	ID            string                 `json:"id"`
	ModuleID      string                 `json:"moduleId"`
	Title         string                 `json:"title"`
	PurchasePrice *float64               `json:"purchasePrice"`
	Images        []string               `json:"images"`
	Tags          []string               `json:"tags"`
	Attributes    map[string]any         `json:"attributes"`
	CreatedAt     string                 `json:"createdAt"`
	UpdatedAt     string                 `json:"updatedAt"`
}

// TagCount holds a tag name and the number of items using it.
type TagCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
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

// Showcase represents a public gallery link for a collection module.
type Showcase struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	TenantID  string `json:"tenantId"`
	ModuleID  string `json:"moduleId"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// Store defines all database operations for items, modules, settings, and showcases.
// Implementations: SQLiteStore (local mode) and PostgresStore (cloud mode).
type Store interface {
	QueryItems(query, moduleID, filtersJSON, tagsJSON string) ([]Item, error)
	InsertItem(item Item) (Item, error)
	UpdateItem(item Item) (Item, error)
	DeleteItem(id string) error
	DeleteItems(ids []string) (int64, error)
	BulkUpdateModule(ids []string, newModuleID string) (int64, error)
	ExportItemsCSV(ids []string, modules []ModuleSchema) (string, error)
	GetAllTags() ([]TagCount, error)
	RenameTag(oldName, newName string) (int64, error)
	DeleteTag(name string) (int64, error)
	GetModules() ([]ModuleSchema, error)
	SaveModule(schema ModuleSchema) error
	LoadModuleFile(id string) (string, error)
	GetSettings() (string, error)
	SaveSettings(json string) error
	GetShowcaseBySlug(slug string) (*Showcase, error)
	GetShowcaseForModule(moduleID string) (*Showcase, error)
	UpsertShowcase(showcase Showcase) error
	ListShowcases() ([]Showcase, error)
	Close() error
}

// slugifyRe matches any character that is not alphanumeric or hyphen.
var slugifyRe = regexp.MustCompile(`[^a-z0-9-]+`)

// GenerateShowcaseSlug creates a URL slug from the module name with a random suffix.
// Format: {slugified-name}-{8-hex-chars}. Name portion is capped at 30 chars.
func GenerateShowcaseSlug(moduleName string) string {
	slug := strings.ToLower(strings.TrimSpace(moduleName))
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = slugifyRe.ReplaceAllString(slug, "")
	// Collapse consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if len(slug) > 30 {
		slug = slug[:30]
	}
	if slug == "" {
		slug = "collection"
	}

	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%s-%x", slug, b)
}
