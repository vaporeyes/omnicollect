# Tasks: AI Metadata Extraction

**Input**: Design documents from `/specs/016-ai-metadata-extraction/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/ai-contract.md, quickstart.md

**Tests**: Not explicitly requested.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Configuration and types shared by all stories

- [x] T001 Update `config.go`: add `AIProvider`, `AIAPIKey`, `AIModel`, `AIBaseURL` fields to Config struct; read from `AI_PROVIDER`, `AI_API_KEY`, `AI_MODEL`, `AI_BASE_URL` env vars; add `IsAIEnabled()` helper returning true when AIProvider is non-empty
- [x] T002 [P] Add `AIAnalysisResult` and `AIStatus` types to `frontend/src/api/types.ts`: AIAnalysisResult (`title?: string`, `attributes: Record<string, any>`, `warnings: string[]`), AIStatus (`enabled: boolean`, `provider?: string`, `model?: string`)
- [x] T003 [P] Add `analyzeItem(imageFilename: string, moduleId: string)` and `getAIStatus()` functions to `frontend/src/api/client.ts`

**Checkpoint**: Config reads AI env vars; frontend types and API functions defined

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: AI provider abstraction and prompt builder

- [x] T004 Create `ai/provider.go`: define `AIProvider` interface with `AnalyzeImage(ctx context.Context, imageBase64 string, prompt string) (string, error)` method; define `AIConfig` struct; implement `NewAIProvider(cfg AIConfig) (AIProvider, error)` factory that returns the correct implementation based on provider name
- [x] T005 Create `ai/anthropic.go`: implement `AnthropicProvider` struct with `AnalyzeImage` method; construct Anthropic Messages API request with model, base64 image in content block, text prompt; parse response to extract text content; handle errors (rate limit, auth failure, timeout)
- [x] T006 Create `ai/openai_compat.go`: implement `OpenAICompatProvider` struct with `AnalyzeImage` method; construct OpenAI chat completions request with model, image_url (data: base64), text prompt; send to `baseURL + "/chat/completions"`; parse response to extract message content; handle errors
- [x] T007 Create `ai/prompt.go`: implement `BuildPrompt(schema ModuleSchema) string` that generates the analysis prompt from schema attributes (name, type, options for enums); implement `ParseAndValidateResponse(jsonStr string, schema ModuleSchema) (map[string]any, string, []string)` that parses AI JSON response, validates enum values against schema options, discards invalid types, returns (attributes, title, warnings)

**Checkpoint**: AI package compiles; `go vet ./ai/...` passes; providers can be instantiated

---

## Phase 3: User Story 1 - Auto-Fill from Photo (Priority: P1) MVP

**Goal**: Upload photo, click "Analyze with AI", form fields auto-populated from vision model.

**Independent Test**: Upload a photo of a recognizable collectible, click analyze, verify form fields populated with plausible values matching schema.

### Implementation for User Story 1

- [x] T008 [US1] Add `handleAnalyzeItem` handler in `handlers.go`: parse JSON body (imageFilename + moduleId); read image bytes from MediaStore; encode as base64; look up module schema; call `BuildPrompt`; call `aiProvider.AnalyzeImage`; call `ParseAndValidateResponse`; return AIAnalysisResult JSON
- [x] T009 [US1] Add `handleAIStatus` handler in `handlers.go`: return AIStatus JSON based on config.IsAIEnabled(), provider name, model
- [x] T010 [US1] Register AI routes in `server.go`: `POST /api/v1/ai/analyze` and `GET /api/v1/ai/status`
- [x] T011 [US1] Initialize AIProvider in `app.go`: if config.IsAIEnabled(), call `ai.NewAIProvider(cfg)` and store on App struct; pass to handlers
- [x] T012 [US1] Modify `frontend/src/components/DynamicForm.vue`: on mount, call `getAIStatus()` to check if AI is enabled; if enabled and item has images, show "Analyze with AI" button below images section; on click, call `analyzeItem(firstImage, moduleId)`; on success, populate empty form fields from result attributes; show toast "AI filled N fields"; on error, show error toast
- [x] T013 [US1] Handle loading state in `DynamicForm.vue`: while AI analysis is in progress, show spinner on the button text ("Analyzing..."), disable the button, but allow editing other fields
- [x] T014 [US1] Handle no-image state in `DynamicForm.vue`: when item has no images, disable the "Analyze with AI" button with title tooltip "Add a photo first"

**Checkpoint**: Full AI analysis flow works: upload photo -> click analyze -> fields populated

---

## Phase 4: User Story 2 - Enrich Existing Items (Priority: P2)

**Goal**: Analyze photos on existing items, filling only empty attributes.

**Independent Test**: Create item with photo but sparse attributes. Edit, click analyze, verify only empty fields filled.

### Implementation for User Story 2

- [x] T015 [US2] Update AI result application logic in `DynamicForm.vue`: when populating form fields from AI result, check each field -- if the form field already has a value (non-empty, non-null, non-default), skip it; only populate fields that are empty
- [x] T016 [US2] Handle multiple images in `DynamicForm.vue`: when item has multiple images, always send the first (primary) image to the AI; no multi-image analysis in v1
- [x] T017 [US2] Show field-level feedback in `DynamicForm.vue`: after AI analysis, briefly highlight fields that were auto-filled (subtle background flash or border color) so the user can see what changed

**Checkpoint**: Existing items can be enriched; manual values preserved; visual feedback on filled fields

---

## Phase 5: User Story 3 - Title Suggestion (Priority: P3)

**Goal**: AI suggests a title; auto-fill if empty, show as clickable hint if title has value.

**Independent Test**: Analyze with empty title -> title auto-filled. Analyze with existing title -> suggestion shown below field.

### Implementation for User Story 3

- [x] T018 [US3] Update AI result handling in `DynamicForm.vue` for title: if `result.title` is non-empty and the title form field is empty, auto-fill the title field
- [x] T019 [US3] Add title suggestion UI in `DynamicForm.vue`: if `result.title` is non-empty and the title field already has a value, show a small clickable suggestion text below the title input (e.g., "AI suggestion: 1922 Peace Dollar") that replaces the title when clicked; dismiss suggestion on save or manual title edit
- [x] T020 [US3] Style the title suggestion in `DynamicForm.vue`: subtle appearance (smaller font, muted color, clickable with hover underline) that doesn't compete with the main form

**Checkpoint**: Title suggestions work in both empty and pre-filled scenarios

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Docker config, documentation, error handling

- [x] T021 Update `docker-compose.yml`: add AI_PROVIDER, AI_API_KEY, AI_MODEL, AI_BASE_URL env vars to the app service with empty defaults and comments
- [x] T022 Update `Dockerfile`: pass AI env vars through (they're runtime, not build-time, so no ARG needed -- just document)
- [x] T023 [P] Update `CLAUDE.md`: document ai/ package, AIProvider interface, env vars, REST endpoints, DynamicForm AI integration
- [x] T024 [P] Update `README.md`: add "AI Metadata Extraction" section with provider setup (Anthropic + OpenRouter), env vars table, usage description, iteration 16 history
- [x] T025 Run quickstart.md acceptance test flow (all 14 steps) and fix any issues

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies -- start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (needs config for provider instantiation)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (needs AI providers + prompt builder)
- **User Story 2 (Phase 4)**: Depends on Phase 3 (extends DynamicForm AI logic)
- **User Story 3 (Phase 5)**: Depends on Phase 3 (extends DynamicForm AI result handling)
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational. MVP -- full AI analysis pipeline.
- **US2 (P2)**: Depends on US1 (extends the "fill only empty" logic).
- **US3 (P3)**: Depends on US1 (extends title handling in the same DynamicForm).

### Parallel Opportunities

- T002 and T003 (Phase 1 frontend) and T001 (Phase 1 backend) -- different languages
- T005 and T006 (Phase 2 providers) -- different files
- T023 and T024 (docs) -- different files

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 + Phase 2 = AI package + config
2. Phase 3 (US1) = full pipeline: analyze endpoint + DynamicForm button
3. **STOP and VALIDATE**: upload photo, click analyze, verify fields populated
4. This delivers the core "AI auto-fill" experience

### Incremental Delivery

1. Phase 1 + Phase 2 = AI abstraction layer
2. Phase 3 (US1) = auto-fill from new photos (MVP)
3. Phase 4 (US2) = enrich existing items (fill-only-empty logic)
4. Phase 5 (US3) = title suggestion
5. Phase 6 = docs + Docker config

---

## Notes

- Two AI provider implementations: Anthropic direct + OpenAI-compatible (OpenRouter + others)
- Prompt dynamically built from module schema (attribute names, types, enum options)
- Response strictly validated: enum values checked against schema, invalid values discarded
- Backend reads image from MediaStore (no re-upload from frontend)
- Feature hidden when AI_PROVIDER is empty (graceful degradation)
- No new Go or npm dependencies (uses net/http + encoding/json)
- OpenRouter config: AI_PROVIDER=openai-compatible, AI_BASE_URL=https://openrouter.ai/api/v1
