// ABOUTME: Schema-to-prompt builder and AI response parser/validator.
// ABOUTME: Constructs structured prompts from module schemas and validates AI output.
package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"omnicollect/storage"
)

// BuildPrompt generates an analysis prompt from a module schema.
// The prompt instructs the AI to return a JSON object matching the schema fields.
func BuildPrompt(schema storage.ModuleSchema) string {
	var b strings.Builder
	b.WriteString("Analyze this image of a collection item. Return a JSON object with these fields:\n")
	b.WriteString("- \"title\" (string): A descriptive name for this item\n")

	for _, attr := range schema.Attributes {
		line := fmt.Sprintf("- %q (%s)", attr.Name, attr.Type)
		if len(attr.Options) > 0 {
			line += fmt.Sprintf(", options: %s", formatOptions(attr.Options))
		}
		if attr.Display != nil && attr.Display.Label != "" {
			line += fmt.Sprintf(": %s", attr.Display.Label)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\nReturn ONLY valid JSON. No markdown fences, no explanation. Omit any field you cannot determine from the image.")
	return b.String()
}

func formatOptions(opts []string) string {
	quoted := make([]string, len(opts))
	for i, o := range opts {
		quoted[i] = fmt.Sprintf("%q", o)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// ParseAndValidateResponse parses AI JSON output and validates it against
// the module schema. Returns validated attributes, suggested title, and
// any warnings about discarded fields.
func ParseAndValidateResponse(jsonStr string, schema storage.ModuleSchema) (map[string]any, string, []string) {
	// Strip markdown code fences if present
	jsonStr = stripCodeFences(jsonStr)
	jsonStr = strings.TrimSpace(jsonStr)

	var raw map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return map[string]any{}, "", []string{fmt.Sprintf("Failed to parse AI response as JSON: %v", err)}
	}

	// Build a lookup of schema attributes by name
	attrByName := make(map[string]storage.AttributeSchema, len(schema.Attributes))
	for _, attr := range schema.Attributes {
		attrByName[attr.Name] = attr
	}

	validated := make(map[string]any)
	var warnings []string

	// Extract title separately
	title := ""
	if t, ok := raw["title"]; ok {
		if s, ok := t.(string); ok {
			title = s
		}
		delete(raw, "title")
	}

	for key, val := range raw {
		attr, exists := attrByName[key]
		if !exists {
			// Not in schema, discard silently
			continue
		}

		validVal, ok := validateValue(val, attr)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("Discarded invalid value for %q", key))
			continue
		}
		validated[key] = validVal
	}

	return validated, title, warnings
}

// validateValue checks that a value matches the expected schema type.
// For enums, verifies the value is in the allowed options list.
func validateValue(val any, attr storage.AttributeSchema) (any, bool) {
	switch attr.Type {
	case "string", "enum":
		s, ok := val.(string)
		if !ok {
			return nil, false
		}
		if attr.Type == "enum" && len(attr.Options) > 0 {
			found := false
			for _, opt := range attr.Options {
				if strings.EqualFold(s, opt) {
					s = opt // normalize to the exact option casing
					found = true
					break
				}
			}
			if !found {
				return nil, false
			}
		}
		return s, true

	case "number":
		// JSON numbers come as float64
		switch v := val.(type) {
		case float64:
			return v, true
		case json.Number:
			f, err := v.Float64()
			if err != nil {
				return nil, false
			}
			return f, true
		default:
			return nil, false
		}

	case "boolean":
		b, ok := val.(bool)
		if !ok {
			return nil, false
		}
		return b, true

	default:
		// Unknown type, accept strings
		if s, ok := val.(string); ok {
			return s, true
		}
		return nil, false
	}
}

// stripCodeFences removes markdown code block delimiters from the response.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	// Remove opening fence (```json or ```)
	if strings.HasPrefix(s, "```") {
		idx := strings.Index(s, "\n")
		if idx != -1 {
			s = s[idx+1:]
		}
	}
	// Remove closing fence
	if strings.HasSuffix(s, "```") {
		s = s[:len(s)-3]
	}
	return strings.TrimSpace(s)
}
