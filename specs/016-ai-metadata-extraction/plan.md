# Implementation Plan: AI Metadata Extraction

**Branch**: `016-ai-metadata-extraction` | **Date**: 2026-04-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/016-ai-metadata-extraction/spec.md`

## Summary

Add AI-powered metadata extraction that analyzes uploaded collection item photos and auto-fills form fields. The backend constructs a prompt from the module schema + image, calls a configurable AI vision provider (Anthropic direct, Google direct, or any OpenAI-compatible endpoint including OpenRouter), validates the response against the schema, and returns structured attribute values. The frontend adds an "Analyze with AI" button to the item form.

## Technical Context

**Language/Version**: Go 1.25+ (backend AI client + handler), TypeScript + Vue 3 (frontend button + form integration)
**Primary Dependencies**: No new Go dependencies (uses `net/http` for AI API calls + `encoding/json`); no new frontend dependencies
**Storage**: No database changes; AI results populate existing Item attributes
**Testing**: Unit tests for prompt construction + response parsing; mock AI responses
**Target Platform**: Docker container (cloud) + macOS desktop (local)
**Project Type**: Multi-tenant SaaS + desktop hybrid
**Performance Goals**: AI response under 10 seconds; schema validation under 100ms
**Constraints**: Configurable provider via env vars; graceful degradation when not configured; OpenRouter support via OpenAI-compatible base URL

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Local-First | PASS (mitigated) | AI requires network; feature is optional. App works fully without AI. |
| II. Schema-Driven UI | PASS | AI prompt is dynamically built from the module schema; no hardcoded type-specific logic |
| III. Flat Data Architecture | PASS | AI results populate existing attributes JSON; no schema changes |
| IV. Performance & Memory | PASS | Image sent as base64 in the API call; no persistent memory impact |
| V. Type-Safe IPC | PASS | New REST endpoint with typed request/response |
| VI. Documentation | PASS | Spec artifacts produced |

All gates pass. Principle I mitigated: AI is an additive feature that requires network, but the app is fully functional without it.

## Project Structure

### Source Code (repository root)

```text
# Backend (Go)
ai/
  provider.go          # New: AI provider interface + factory function
  anthropic.go         # New: Anthropic Messages API client (direct)
  openai_compat.go     # New: OpenAI-compatible client (works with OpenRouter, Google, etc.)
  prompt.go            # New: schema-to-prompt builder + response parser/validator
config.go              # Modified: add AI_PROVIDER, AI_API_KEY, AI_MODEL, AI_BASE_URL env vars
handlers.go            # Modified: add handleAnalyzeItem endpoint
server.go              # Modified: register AI analysis route

# Frontend (Vue/TypeScript)
frontend/src/
  api/client.ts        # Modified: add analyzeItem function
  api/types.ts         # Modified: add AIAnalysisResult type
  components/
    DynamicForm.vue    # Modified: add "Analyze with AI" button, loading state, result application
```

**Structure Decision**: A new `ai/` Go package encapsulates all AI provider logic. Two provider implementations (Anthropic direct + OpenAI-compatible) behind a common interface. The prompt builder is provider-agnostic. One new REST endpoint. Minimal frontend changes (button + result handling in DynamicForm).

## Complexity Tracking

| Concern | Justification | Simpler Alternative Rejected Because |
|---------|--------------|-------------------------------------|
| Constitution I (Local-First) | AI is optional; app works fully offline | Removing the feature would eliminate significant user value |
| Two provider implementations | Anthropic has a unique API format; OpenAI-compatible covers OpenRouter + Google + others | A single OpenAI-compatible client would require an Anthropic-to-OpenAI translation layer that's fragile |
