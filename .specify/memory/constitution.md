<!--
  Sync Impact Report
  ====================
  Version change: 1.0.0 -> 1.1.0
  Modified principles:
    - Principle VI: Documentation is Paramount (materially expanded)
  Added sections: None
  Removed sections: None
  Templates requiring updates:
    - .specify/templates/plan-template.md: ✅ No updates needed
    - .specify/templates/spec-template.md: ✅ No updates needed
    - .specify/templates/tasks-template.md: ✅ No updates needed
  Follow-up TODOs: None
-->

# OmniCollect Constitution

## Core Principles

### I. Local-First Mandate

The primary source of truth is the local SQLite database. The application
MUST provide 100% functionality without an active internet connection.
No centralized cloud accounts are required for core operation.

- All data MUST be readable and writable offline.
- Network features (sync, backup, sharing) are supplementary and MUST
  degrade gracefully when connectivity is unavailable.
- No feature may depend on a remote service for its primary function.

### II. Schema-Driven UI

UI forms for collection items MUST be generated at runtime from JSON
schemas. There MUST NOT be hardcoded Vue templates for specific item
types.

- Files such as `BookForm.vue`, `CoinForm.vue`, or any type-specific
  form component are explicitly forbidden.
- A single, generic form renderer MUST consume JSON schema definitions
  to produce the appropriate input fields, validation, and layout.
- Adding a new collection type MUST require only a new JSON schema
  definition, not new UI code.

### III. Flat Data Architecture

All item metadata MUST be stored as flat JSON documents within a single
SQLite table using an Entity-Attribute-Value (EAV) pattern.

- Complex SQL JOIN operations for item attributes are prohibited.
- Queries against item metadata MUST operate on the flat JSON structure.
- Schema evolution (adding new fields to a collection type) MUST NOT
  require database migrations.

### IV. Performance & Memory Protection

The frontend MUST NOT load original, high-resolution media files into
list or grid views. All list renders MUST use generated, highly
compressed local thumbnails.

- Thumbnail generation MUST occur at import time or on first access.
- Original media files are loaded only in detail/edit views on explicit
  user action.
- The frontend MUST NOT hold references to full-resolution image data
  in list/grid rendering paths.

### V. Type-Safe IPC

All communication between the Vue 3 frontend and Go backend MUST use
Wails-generated TypeScript bindings. Raw, untyped IPC calls are
prohibited.

- Every backend function exposed to the frontend MUST have a
  corresponding generated TypeScript binding.
- Direct `window.go` or equivalent untyped calls are forbidden.
- Type mismatches between frontend and backend MUST be caught at
  compile time, not runtime.

### VI. Documentation is Paramount

All features, APIs, and architectural decisions MUST be documented.
Documentation is a deliverable, not an afterthought.

- Public functions and exported types MUST have meaningful
  documentation.
- Architectural decisions MUST be recorded with rationale.
- User-facing features MUST include usage documentation before a
  feature is considered complete.
- The project README MUST be updated as part of every iteration to
  reflect new features, changed project structure, and updated
  dependency lists. A feature is not complete until the README
  accurately describes the current state of the application.
- CLAUDE.md (or equivalent agent context file) MUST be updated to
  reflect new bindings, components, conventions, and data locations
  added by each iteration.
- Each iteration MUST produce spec artifacts (spec.md, plan.md,
  research.md, data-model.md, contracts/, quickstart.md, tasks.md)
  in the `specs/` directory before implementation begins.

## Technology Stack

- **Frontend**: Vue 3 (Composition API, TypeScript)
- **Backend**: Go
- **Desktop Framework**: Wails (provides native desktop shell and
  type-safe IPC bindings)
- **Database**: SQLite (local, embedded)
- **Schema Format**: JSON Schema (drives UI generation per Principle II)

## Compliance Review

Any PR or feature addition that violates these principles MUST be
refactored before merge.

- Code reviews MUST include a constitution compliance check against
  all six principles.
- Automated checks SHOULD enforce Principles IV (no raw media in
  lists) and V (no untyped IPC) where feasible.
- Violations discovered post-merge MUST be tracked as high-priority
  issues and resolved in the next development cycle.

## Governance

This constitution supersedes all other development practices for the
OmniCollect project. Amendments follow this procedure:

1. Propose the change with rationale in a dedicated PR.
2. Document the specific principle(s) affected and the motivation.
3. Update this file, increment the version per semver rules:
   - **MAJOR**: Principle removal or incompatible redefinition.
   - **MINOR**: New principle added or existing principle materially
     expanded.
   - **PATCH**: Clarifications, wording, or non-semantic refinements.
4. Update any dependent templates or documentation as identified in
   the Sync Impact Report.

**Version**: 1.1.0 | **Ratified**: 2026-04-04 | **Last Amended**: 2026-04-05
