# Feature Specification: AI Metadata Extraction

**Feature Branch**: `016-ai-metadata-extraction`  
**Created**: 2026-04-10  
**Status**: Draft  
**Input**: User description: "When a user uploads a photo, pipe it to an AI vision model that analyzes the image and returns a JSON payload matching the module schema, auto-filling attributes like manufacturer, year, and condition."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Auto-Fill Attributes from Uploaded Photo (Priority: P1)

A collector creating a new item uploads a photo of a vintage typewriter. After the image uploads, they click an "Analyze with AI" button. The system sends the image along with the active module's schema to a vision AI service. Within seconds, the form fields (manufacturer, year, condition, etc.) are auto-populated with values extracted from the image. The user reviews the suggestions, corrects any mistakes, and saves the item.

**Why this priority**: This is the core value -- transforming a photo into structured metadata without manual typing. It delivers the "wow" moment and covers 90% of the use case.

**Independent Test**: Upload a photo of a recognizable collectible (e.g., a US quarter). Click "Analyze with AI". Verify the form fields are populated with plausible values matching the module schema attributes.

**Acceptance Scenarios**:

1. **Given** a user is creating a new item with a module that has enum/string/number attributes, **When** they upload a photo and click "Analyze with AI", **Then** the AI returns values for each attribute and the form fields are pre-filled.
2. **Given** the AI returns values, **Then** all returned values respect the schema constraints (enum values are from the defined options list, numbers are numeric, booleans are true/false).
3. **Given** the AI cannot determine a value for an attribute, **Then** that field is left empty rather than filled with a guess.
4. **Given** the form is pre-filled by AI, **When** the user reviews and modifies any value, **Then** the manual edit takes precedence and the item saves with the user's version.
5. **Given** no AI service is configured (no API key), **Then** the "Analyze with AI" button is hidden and the form works exactly as before.

---

### User Story 2 - AI Analysis on Existing Items (Priority: P2)

A collector has items that were added without full metadata. They open an existing item's edit form, where the item already has photos attached. They click "Analyze with AI" to have the AI examine the existing photos and suggest attribute values for empty fields. Only empty fields are auto-filled; fields the user has already set remain unchanged.

**Why this priority**: Many users have collections with photos but sparse metadata. This lets them retroactively enrich existing items without re-entering data.

**Independent Test**: Create an item with a photo but only a title (no attributes filled). Edit the item. Click "Analyze with AI". Verify empty attribute fields are populated while the title remains unchanged.

**Acceptance Scenarios**:

1. **Given** an existing item with photos and some attributes already filled, **When** the user clicks "Analyze with AI", **Then** only empty attributes are populated (existing values are preserved).
2. **Given** an item with multiple photos, **When** "Analyze with AI" is clicked, **Then** the first (primary) image is sent to the AI for analysis.
3. **Given** an item with no photos, **When** the form renders, **Then** the "Analyze with AI" button is disabled with a tooltip "Add a photo first".

---

### User Story 3 - Title Suggestion from AI (Priority: P3)

In addition to attributes, the AI suggests a title for the item based on the image content. If the title field is empty, the suggestion is auto-filled. If the title already has a value, the suggestion is shown as a subtle hint below the field that the user can click to apply.

**Why this priority**: Title is the most visible field and benefits from AI assistance, but it's lower priority than structured attributes because users often already know what they're adding.

**Independent Test**: Upload a photo with an empty title field. Click "Analyze with AI". Verify a title is suggested. Verify that if a title already exists, it's not overwritten.

**Acceptance Scenarios**:

1. **Given** the title field is empty, **When** AI analysis completes, **Then** the title is auto-filled with the AI's suggestion.
2. **Given** the title field already has a value, **When** AI analysis completes, **Then** the title is NOT overwritten; a clickable suggestion appears below the field.
3. **Given** the AI suggestion for title, **When** the user clicks the suggestion, **Then** the title field is updated with the suggestion.

---

### Edge Cases

- What happens when the AI service is unavailable or returns an error? A toast notification informs the user ("AI analysis unavailable, please try again later") and no fields are modified.
- What happens when the image is not a photo of a collectible (e.g., a screenshot, a blank image)? The AI returns empty/partial results; the user sees which fields were not filled.
- What happens when the AI returns an enum value that is not in the schema's options list? The value is discarded for that field (only valid enum options are accepted).
- What happens with rate limiting from the AI provider? The system respects rate limits and shows a message if the user hits them ("AI analysis limit reached, please wait").
- What happens when the AI analysis takes a long time? A loading spinner is shown on the "Analyze with AI" button. The user can continue editing other fields while waiting.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The item creation/edit form MUST provide an "Analyze with AI" button when AI is configured and at least one photo is attached.
- **FR-002**: Clicking "Analyze with AI" MUST send the item's primary image and the active module's schema to a backend endpoint.
- **FR-003**: The backend MUST construct a prompt that includes the schema's attribute definitions (name, type, options for enums) and instruct the AI to return a JSON object matching the schema.
- **FR-004**: The AI response MUST be parsed and validated against the schema before populating form fields.
- **FR-005**: Enum attribute values from the AI MUST be validated against the schema's options list; invalid values MUST be discarded.
- **FR-006**: When analyzing an existing item, only empty attributes MUST be populated; attributes with existing values MUST be preserved.
- **FR-007**: The "Analyze with AI" button MUST be hidden when no AI service is configured (no API key set).
- **FR-008**: The feature MUST show a loading state while the AI processes the image and MUST NOT block the user from editing other fields.
- **FR-009**: AI service errors MUST be surfaced as toast notifications; no fields MUST be modified on error.
- **FR-010**: The AI service configuration (provider, API key, model) MUST be configurable via environment variables.
- **FR-011**: The system MUST support multiple AI vision providers via a configurable base URL + API key pattern. Supported providers include Anthropic (direct), Google (direct), and any OpenAI-compatible API endpoint (including OpenRouter, which provides access to hundreds of models via a single API key and endpoint).
- **FR-012**: The AI provider configuration MUST support a custom base URL (`AI_BASE_URL`) to enable routing requests through OpenRouter or other OpenAI-compatible proxies.

### Key Entities

- **AI Analysis Request**: An image (bytes or URL) paired with a module schema, sent to the AI provider. Contains the prompt instructing the AI to return structured JSON.
- **AI Analysis Response**: A JSON object with keys matching the module schema's attribute names. Values conform to the attribute types. May include a suggested title.
- **AI Provider Configuration**: Environment variables specifying the provider (e.g., "anthropic", "google", or "openai-compatible"), API key, model name, and optionally a custom base URL for OpenRouter or other proxies.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: AI analysis returns structured metadata within 10 seconds of the user clicking "Analyze with AI".
- **SC-002**: At least 70% of auto-filled attributes are correct for recognizable collectible items (coins, stamps, books, hardware).
- **SC-003**: Zero invalid enum values appear in the form after AI analysis (all values validated against schema).
- **SC-004**: The feature degrades gracefully when AI is not configured -- zero UI changes, zero errors, identical behavior to before.
- **SC-005**: Users can override any AI-suggested value before saving, with zero data loss from manual edits.

## Clarifications

### Session 2026-04-10

- Q: Should the system support OpenRouter in addition to direct Anthropic/Google? -> A: Yes. Support a configurable base URL (`AI_BASE_URL`) for OpenAI-compatible endpoints including OpenRouter. Set `AI_PROVIDER=openai-compatible` to use any OpenRouter model.

## Assumptions

- The AI vision service is called via a standard HTTP API from the Go backend. The backend proxies the request to avoid exposing API keys to the frontend.
- The default AI provider is Anthropic (Claude) with vision capabilities. The provider is configurable via `AI_PROVIDER` (e.g., "anthropic", "google", "openai-compatible"), `AI_API_KEY`, `AI_MODEL`, and optionally `AI_BASE_URL` environment variables. For OpenRouter, set `AI_PROVIDER=openai-compatible`, `AI_BASE_URL=https://openrouter.ai/api/v1`, `AI_API_KEY` to the OpenRouter key, and `AI_MODEL` to the desired model (e.g., `anthropic/claude-sonnet-4.6` or `google/gemma-4-31b-it`).
- The prompt sent to the AI includes the module schema as JSON, the image as base64 (or URL), and instructions to return a JSON object with attribute names as keys.
- The AI response is expected to be a JSON object. If the AI returns non-JSON or malformed JSON, the response is treated as an error (no fields populated).
- Only one image is analyzed per request (the first/primary image). Multi-image analysis is out of scope for v1.
- The AI analysis is an optional enhancement. All existing item creation/editing workflows continue to work without AI. There is no requirement for AI analysis to create an item.
- The cost per AI analysis call is borne by the application operator (the API key owner). No per-user billing or quotas are implemented in v1.
- The AI feature works in both local mode (desktop) and cloud mode (Docker). In local mode, the user provides their own API key via environment variable.
