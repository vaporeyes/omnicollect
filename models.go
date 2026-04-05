// ABOUTME: Shared data types for OmniCollect items and module schemas.
// ABOUTME: These structs are used by Wails to generate TypeScript bindings.
package main

// Item represents a single collectible record stored in SQLite.
type Item struct {
	ID            string                 `json:"id"`
	ModuleID      string                 `json:"moduleId"`
	Title         string                 `json:"title"`
	PurchasePrice *float64               `json:"purchasePrice"`
	Images        []string               `json:"images"`
	Attributes    map[string]any `json:"attributes"`
	CreatedAt     string                 `json:"createdAt"`
	UpdatedAt     string                 `json:"updatedAt"`
}

// ModuleSchema defines a collection type loaded from a JSON file.
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
