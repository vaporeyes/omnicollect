# Data Model: AI Metadata Extraction

**Branch**: `016-ai-metadata-extraction` | **Date**: 2026-04-10

## No Database Schema Changes

AI analysis populates existing Item attributes. No new columns, tables, or indexes.

## New Types

### AIAnalysisRequest (backend internal)

| Field | Type | Description |
|-------|------|-------------|
| imageFilename | string | Filename of the item's primary image (read from MediaStore) |
| schema | ModuleSchema | Active module schema for prompt construction |
| existingAttributes | map[string]any | Current attribute values (for "fill only empty" logic) |

### AIAnalysisResult (returned to frontend)

| Field | Type | Description |
|-------|------|-------------|
| title | string (optional) | Suggested title; empty if not determined |
| attributes | map[string]any | Extracted attribute values, validated against schema |
| warnings | string[] | Non-fatal issues (e.g., "Could not determine 'year'") |

### AIProviderConfig (from environment)

| Field | Source | Description |
|-------|--------|-------------|
| Provider | AI_PROVIDER | "anthropic", "openai-compatible", or empty (disabled) |
| APIKey | AI_API_KEY | API authentication key |
| Model | AI_MODEL | Model identifier |
| BaseURL | AI_BASE_URL | Custom endpoint URL (required for openai-compatible) |

## Go Interface

```
type AIProvider interface {
    AnalyzeImage(ctx context.Context, imageBase64 string, prompt string) (string, error)
}
```

Two implementations:
- `AnthropicProvider` -- calls Anthropic Messages API directly
- `OpenAICompatProvider` -- calls any OpenAI-compatible endpoint (OpenRouter, Google, etc.)
