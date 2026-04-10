# AI Analysis Contract: REST API + UI

**Branch**: `016-ai-metadata-extraction` | **Date**: 2026-04-10

## REST API

### Analyze Item Image
`POST /api/v1/ai/analyze`

**Request Body**:
```json
{
  "imageFilename": "abc123.jpg",
  "moduleId": "coins"
}
```

**Response**: `200 OK` -- `AIAnalysisResult`
```json
{
  "title": "1922 Peace Dollar",
  "attributes": {
    "year": 1922,
    "condition": "Fine",
    "country": "United States",
    "isGraded": false
  },
  "warnings": []
}
```

**Errors**:
- `400` -- Missing imageFilename or moduleId
- `404` -- Image not found in storage, or module not found
- `503` -- AI service unavailable or returned an error
- `501` -- AI not configured (AI_PROVIDER empty)

### Check AI Availability
`GET /api/v1/ai/status`

**Response**: `200 OK`
```json
{
  "enabled": true,
  "provider": "openai-compatible",
  "model": "anthropic/claude-sonnet-4.6"
}
```

When AI is not configured:
```json
{
  "enabled": false
}
```

## UI Integration

### DynamicForm Changes

When AI is enabled (checked via `/api/v1/ai/status` on mount):

1. An "Analyze with AI" button appears below the image attachment section
2. Button is disabled when no images are attached (tooltip: "Add a photo first")
3. Clicking the button:
   - Shows a loading spinner on the button
   - Calls `POST /api/v1/ai/analyze` with the first image's filename + active moduleId
   - On success: populates empty form fields with returned attributes; shows toast "AI filled N fields"
   - On error: shows error toast; no fields modified
4. For existing items with values: only empty fields are filled; fields with values are preserved
5. Title suggestion: if title is empty, auto-fill; if title has value, show clickable suggestion below field

## Provider API Formats

### Anthropic Messages API
```
POST https://api.anthropic.com/v1/messages
Headers: x-api-key, anthropic-version
Body: {model, messages: [{role: "user", content: [{type: "image", source: {type: "base64", ...}}, {type: "text", text: prompt}]}]}
```

### OpenAI-Compatible (OpenRouter, Google, etc.)
```
POST {AI_BASE_URL}/chat/completions
Headers: Authorization: Bearer {key}
Body: {model, messages: [{role: "user", content: [{type: "image_url", image_url: {url: "data:image/jpeg;base64,..."}}, {type: "text", text: prompt}]}]}
```
