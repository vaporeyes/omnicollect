// ABOUTME: Module schema loader that reads JSON definitions from disk.
// ABOUTME: Scans ~/.omnicollect/modules/ at startup and parses each .json file.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// modulesDir returns the path to the modules directory.
func modulesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".omnicollect", "modules"), nil
}

// loadModuleSchemas scans the modules directory and parses all valid JSON schemas.
// Creates the directory if it does not exist. Skips and logs malformed files.
// Returns an error only if the directory cannot be created or read.
func loadModuleSchemas() ([]ModuleSchema, error) {
	dir, err := modulesDir()
	if err != nil {
		return nil, fmt.Errorf("resolving modules dir: %w", err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating modules dir: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading modules dir: %w", err)
	}

	var schemas []ModuleSchema
	seen := make(map[string]string) // module ID -> filename

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		schema, err := parseModuleSchema(path)
		if err != nil {
			log.Printf("skipping malformed schema %s: %v", entry.Name(), err)
			continue
		}

		if existing, ok := seen[schema.ID]; ok {
			log.Printf("skipping duplicate module ID %q in %s (already defined in %s)", schema.ID, entry.Name(), existing)
			continue
		}

		seen[schema.ID] = entry.Name()
		schemas = append(schemas, *schema)
	}

	if schemas == nil {
		schemas = []ModuleSchema{}
	}

	return schemas, nil
}

// parseModuleSchema reads and validates a single JSON schema file.
func parseModuleSchema(path string) (*ModuleSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var schema ModuleSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	if err := validateModuleSchema(&schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

// findModuleFile returns the file path for a given module ID by scanning
// the modules directory. Returns empty string if not found.
func findModuleFile(moduleID string) (string, error) {
	dir, err := modulesDir()
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		schema, err := parseModuleSchema(path)
		if err != nil {
			continue
		}
		if schema.ID == moduleID {
			return path, nil
		}
	}
	return "", nil
}

// saveModuleFile writes a ModuleSchema as formatted JSON to the modules directory.
func saveModuleFile(schema *ModuleSchema) error {
	dir, err := modulesDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating modules dir: %w", err)
	}

	// Check for ID conflict (same ID in a different file)
	existingPath, err := findModuleFile(schema.ID)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(dir, schema.ID+".json")
	if existingPath != "" && existingPath != targetPath {
		// Same ID exists in a differently-named file
		targetPath = existingPath
	}

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling schema: %w", err)
	}

	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("writing schema file: %w", err)
	}

	return nil
}

// validateModuleSchema checks required fields and attribute uniqueness.
func validateModuleSchema(schema *ModuleSchema) error {
	if schema.ID == "" {
		return fmt.Errorf("missing required field: id")
	}
	if schema.DisplayName == "" {
		return fmt.Errorf("missing required field: displayName")
	}

	validTypes := map[string]bool{
		"string": true, "number": true, "boolean": true, "date": true, "enum": true,
	}

	attrNames := make(map[string]bool)
	for _, attr := range schema.Attributes {
		if attr.Name == "" {
			return fmt.Errorf("attribute missing name")
		}
		if !validTypes[attr.Type] {
			return fmt.Errorf("attribute %q has unrecognized type %q", attr.Name, attr.Type)
		}
		if attrNames[attr.Name] {
			return fmt.Errorf("duplicate attribute name: %q", attr.Name)
		}
		attrNames[attr.Name] = true

		if attr.Type == "enum" && len(attr.Options) == 0 {
			return fmt.Errorf("enum attribute %q must have at least one option", attr.Name)
		}
	}

	return nil
}
