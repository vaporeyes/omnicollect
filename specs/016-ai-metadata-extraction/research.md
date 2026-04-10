# Research: AI Metadata Extraction

**Branch**: `016-ai-metadata-extraction` | **Date**: 2026-04-10

## R1: AI Provider Architecture

**Decision**: Two provider implementations behind a common `AIProvider` interface: (1) Anthropic Messages API client for direct Anthropic usage, (2) OpenAI-compatible chat completions client for OpenRouter, Google, and any other provider exposing the OpenAI API format.

**Rationale**: Anthropic's Messages API has a unique format (messages with content blocks, base64 image in source) that differs from the OpenAI chat completions format. Rather than a translation layer, two lean implementations give clean, maintainable code. The OpenAI-compatible client covers OpenRouter (which exposes the OpenAI format), Google's Gemini (via OpenAI-compatible mode), and any future provider.

**Interface**:
```
type AIProvider interface {
    AnalyzeImage(ctx context.Context, imageBase64 string, prompt string) (string, error)
}
```

**Factory**: `NewAIProvider(provider, apiKey, model, baseURL string) AIProvider` -- returns the right implementation based on provider name.

## R2: Prompt Construction Strategy

**Decision**: Build a structured prompt that includes:
1. System message: "You are a collection item metadata extractor. Analyze the image and return a JSON object."
2. Schema description: list each attribute with its name, type, and options (for enums)
3. Output format: "Return ONLY a valid JSON object with these keys. For enum fields, use only the listed options. If you cannot determine a value, omit the key."
4. Image: attached as base64 in the vision message

**Rationale**: Explicit schema instructions + "omit if unknown" reduces hallucination. Enum constraints in the prompt help the model stay within valid options. Asking for JSON-only output simplifies parsing.

**Example prompt for a Coins module**:
```
Analyze this image of a collection item. Return a JSON object with these fields:
- "title" (string): A descriptive name for this item
- "year" (number): The year of manufacture or minting
- "condition" (enum, options: ["Mint", "Fine", "Very Fine", "Good", "Poor"]): The physical condition
- "country" (string): Country of origin
- "isGraded" (boolean): Whether this item has been professionally graded

Return ONLY valid JSON. Omit any field you cannot determine.
```

## R3: Response Parsing and Validation

**Decision**: Parse the AI response as JSON. For each key in the response:
1. Check if it matches a schema attribute name (or "title")
2. Validate the type matches the schema type
3. For enums: check the value is in the options list; discard if not
4. For numbers: verify the value is numeric
5. For booleans: verify the value is true/false
6. Discard any keys not in the schema

**Rationale**: Strict validation prevents invalid data from reaching the form. The "discard invalid" approach is safer than trying to coerce values.

## R4: OpenRouter Integration

**Decision**: OpenRouter uses the OpenAI chat completions API format (`POST /chat/completions` with model, messages, and base64 images in content blocks). The `openai_compat.go` client sends requests to `AI_BASE_URL + "/chat/completions"` with the configured model and API key in the `Authorization: Bearer` header.

**Configuration example**:
```
AI_PROVIDER=openai-compatible
AI_BASE_URL=https://openrouter.ai/api/v1
AI_API_KEY=sk-or-...
AI_MODEL=anthropic/claude-sonnet-4.6
```

**Rationale**: OpenRouter's API is a direct superset of the OpenAI API. Any model available on OpenRouter (Anthropic, Google, Meta, etc.) works by changing `AI_MODEL`.

## R5: Image Handling

**Decision**: The backend reads the item's primary image from MediaStore, encodes it as base64, and includes it in the AI request. The frontend sends only the item ID (or image filename) to the analyze endpoint, not the image bytes.

**Rationale**: The image is already stored in the backend (MediaStore). Re-uploading it from the frontend would be wasteful. The backend reads the stored image and encodes it for the AI call.

## R6: Configuration Variables

| Variable | Default | Example | Description |
|----------|---------|---------|-------------|
| AI_PROVIDER | (empty = disabled) | "anthropic" | Provider type |
| AI_API_KEY | (empty) | "sk-ant-..." | API key |
| AI_MODEL | (empty) | "claude-sonnet-4-6-20250514" | Model identifier |
| AI_BASE_URL | (provider default) | "https://openrouter.ai/api/v1" | Custom base URL (required for openai-compatible) |

When `AI_PROVIDER` is empty, the feature is disabled (button hidden, endpoint returns 404).
