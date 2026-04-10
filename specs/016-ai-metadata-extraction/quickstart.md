# Quickstart: AI Metadata Extraction

**Branch**: `016-ai-metadata-extraction` | **Date**: 2026-04-10

## Prerequisites

- Existing codebase on the `016-ai-metadata-extraction` branch
- An AI API key (Anthropic, OpenRouter, or any OpenAI-compatible provider)
- Items with photos for testing

## No New Dependencies

Go `net/http` + `encoding/json` for AI API calls. No new npm packages.

## Files to Create

### Backend
1. **`ai/provider.go`** -- AIProvider interface, NewAIProvider factory, config struct
2. **`ai/anthropic.go`** -- Anthropic Messages API client (direct)
3. **`ai/openai_compat.go`** -- OpenAI-compatible client (OpenRouter, Google, etc.)
4. **`ai/prompt.go`** -- Schema-to-prompt builder, response JSON parser, schema validator

## Files to Modify

### Backend
1. **`config.go`** -- Add AI_PROVIDER, AI_API_KEY, AI_MODEL, AI_BASE_URL env vars
2. **`handlers.go`** -- Add handleAnalyzeItem and handleAIStatus endpoints
3. **`server.go`** -- Register AI routes
4. **`app.go`** -- Initialize AIProvider based on config

### Frontend
5. **`frontend/src/api/types.ts`** -- Add AIAnalysisResult and AIStatus types
6. **`frontend/src/api/client.ts`** -- Add analyzeItem and getAIStatus functions
7. **`frontend/src/components/DynamicForm.vue`** -- Add "Analyze with AI" button, loading state, result application, title suggestion

## Implementation Order

1. Create ai/ package (interface + providers + prompt builder)
2. Add config env vars
3. Add REST endpoints (analyze + status)
4. Wire AI provider in app.go
5. Frontend: add API client functions + types
6. Frontend: add button + form integration in DynamicForm
7. Test with Anthropic direct
8. Test with OpenRouter
9. Update CLAUDE.md and README

## Environment Variables

```bash
# Anthropic direct
AI_PROVIDER=anthropic
AI_API_KEY=sk-ant-api03-...
AI_MODEL=claude-sonnet-4-6-20250514

# OpenRouter (access to hundreds of models)
AI_PROVIDER=openai-compatible
AI_BASE_URL=https://openrouter.ai/api/v1
AI_API_KEY=sk-or-v1-...
AI_MODEL=anthropic/claude-sonnet-4.6

# Disabled (default)
AI_PROVIDER=
```

## Acceptance Test Flow

1. Start server without AI env vars -- verify no "Analyze with AI" button visible
2. Set AI env vars (e.g., Anthropic direct), restart server
3. Create a new item in a module with enum/number/string attributes
4. Upload a photo of a recognizable collectible
5. Click "Analyze with AI" -- verify loading spinner appears
6. Verify form fields are populated with plausible values
7. Verify enum values are from the schema's options list
8. Modify an AI-suggested value, save -- verify manual edit persists
9. Open an existing item with photo + some filled attributes
10. Click "Analyze with AI" -- verify only empty fields are filled
11. Test with OpenRouter: change env vars to OpenRouter config, restart
12. Repeat steps 4-6 with OpenRouter -- verify same behavior
13. Test error handling: set invalid API key, click analyze -- verify error toast
14. Test no-image state: create item without photo -- verify button is disabled
