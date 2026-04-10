# Feature Specification: Public Showcase URLs

**Feature Branch**: `017-public-showcase`  
**Created**: 2026-04-10  
**Status**: Draft  
**Input**: User description: "Allow users to toggle a collection to Public. Generate a beautiful, read-only gallery link they can share on forums or with prospective buyers."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Make a Collection Public and Get a Shareable Link (Priority: P1)

A collector has a curated coin collection and wants to share it with fellow enthusiasts on a forum. In the collection settings (or module selector), they toggle their "Coins" collection to "Public." The system generates a unique, shareable URL (e.g., `/showcase/{slug}`). They copy the link and paste it in a forum post. Anyone who visits the link sees a beautiful, read-only gallery of the public collection with item images, titles, and attributes -- without needing to log in.

**Why this priority**: The shareable link is the core value. Without it, nothing else in this feature works. This story delivers the ability to share and the public-facing gallery view.

**Independent Test**: Toggle a collection to public. Copy the generated link. Open it in an incognito browser (no auth). Verify the gallery displays all items from that collection with images and metadata, read-only.

**Acceptance Scenarios**:

1. **Given** a collection with items, **When** the user toggles it to "Public", **Then** a unique shareable URL is generated and displayed to the user.
2. **Given** a public showcase URL, **When** anyone visits it without logging in, **Then** they see a read-only gallery of the collection's items with thumbnails, titles, and key attributes.
3. **Given** a public showcase, **When** a visitor tries to edit, delete, or modify any item, **Then** no edit controls are available -- the view is strictly read-only.
4. **Given** a public showcase URL, **Then** it does NOT expose the owner's identity, private settings, or any data from other collections.
5. **Given** the showcase URL includes the collection module's display name and a slug, **Then** the URL is human-readable (e.g., `/showcase/coins-abc123` not `/showcase/8f3a-4b2c`).

---

### User Story 2 - Toggle a Collection Back to Private (Priority: P2)

A collector who previously shared their collection decides they no longer want it public. They toggle the collection back to "Private." The showcase URL immediately stops working -- visitors see a "This showcase is no longer available" message.

**Why this priority**: Users must have full control over their shared data. Without the ability to revoke, users would hesitate to make anything public.

**Independent Test**: Make a collection public, copy the URL. Toggle it back to private. Visit the URL. Verify it returns a "not available" page.

**Acceptance Scenarios**:

1. **Given** a public collection, **When** the user toggles it to "Private", **Then** the showcase URL immediately stops serving the gallery.
2. **Given** a previously public URL that is now private, **When** anyone visits it, **Then** they see a friendly "This showcase is no longer available" page.
3. **Given** a collection is toggled public -> private -> public again, **Then** the same URL works again (the slug is stable, not regenerated).

---

### User Story 3 - Showcase Gallery Design (Priority: P3)

The public showcase page is a visually polished, standalone gallery page designed for external viewers. It features a hero header with the collection name, a responsive grid of item cards with large thumbnails, and an item detail overlay when clicking a card. The design stands on its own without the application's sidebar, toolbar, or navigation.

**Why this priority**: The gallery must look good enough to share publicly. A raw data dump would reflect poorly on the collector. This story ensures the showcase is a "showcase."

**Independent Test**: Visit a showcase URL. Verify the page has a clean header, responsive image grid, item detail on click, and no application chrome (no sidebar, no toolbar, no edit buttons).

**Acceptance Scenarios**:

1. **Given** a showcase page, **Then** it displays a header with the collection name (module display name) and item count.
2. **Given** the item grid, **Then** items display as cards with large thumbnails, title, and key attributes (matching the module schema).
3. **Given** a visitor clicks an item card, **Then** a detail overlay shows the full-resolution image, all attributes, and any tags.
4. **Given** the page is viewed on mobile, **Then** the grid and detail overlay are responsive and usable.
5. **Given** the showcase page, **Then** it has NO sidebar, toolbar, edit buttons, login prompts, or application navigation.

---

### Edge Cases

- What happens when a public collection has zero items? The showcase page shows the collection name and a message "This collection is empty."
- What happens when an item in a public collection has no images? The item card shows a placeholder thumbnail (same as the grid view).
- What happens when the same user makes multiple collections public? Each gets its own unique showcase URL.
- What happens with the showcase URL if the user deletes their account or the tenant is removed? The URL stops working (returns 404).
- What happens when a very large collection is shared (1000+ items)? The showcase page should paginate or lazy-load items to maintain performance.
- What happens in local/desktop mode (no public URLs possible)? The showcase feature is disabled -- toggle is hidden. It requires the cloud/server deployment to serve public pages.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Users MUST be able to toggle a collection (module) to "Public" or "Private" via the application UI.
- **FR-002**: When a collection is toggled to "Public", the system MUST generate a unique, human-readable showcase URL.
- **FR-003**: The showcase URL MUST serve a read-only gallery page accessible without authentication.
- **FR-004**: The showcase page MUST display the collection name, item count, and a responsive grid of item cards (thumbnail, title, key attributes).
- **FR-005**: Clicking an item card on the showcase page MUST show a detail overlay with full-resolution image, all attributes, and tags.
- **FR-006**: The showcase page MUST NOT expose any edit controls, application navigation, owner identity, or data from other collections.
- **FR-007**: When a collection is toggled back to "Private", the showcase URL MUST immediately stop serving the gallery.
- **FR-008**: A revoked showcase URL MUST display a friendly "no longer available" message.
- **FR-009**: The showcase URL slug MUST be stable across public/private toggles (same URL works if re-enabled).
- **FR-010**: The showcase page MUST be responsive and usable on mobile devices.
- **FR-011**: Large public collections MUST paginate or lazy-load to maintain performance.
- **FR-012**: The showcase feature MUST be disabled in local/desktop mode (only available in cloud deployments with public-facing URLs).

### Key Entities

- **Showcase**: A public-facing view of a single collection (module). Has a unique slug, references a tenant and module ID, and has an enabled/disabled state.
- **Showcase Slug**: A URL-safe, human-readable identifier combining the module display name and a short unique suffix (e.g., "coins-abc123").
- **Showcase Page**: A standalone, server-rendered or client-rendered gallery page served at `/showcase/{slug}` without authentication.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can make a collection public and share the link in under 30 seconds (toggle + copy URL).
- **SC-002**: The showcase page loads and renders the gallery within 3 seconds for collections up to 500 items.
- **SC-003**: 100% of showcase pages are accessible without login -- zero authentication barriers for visitors.
- **SC-004**: Toggling a collection to private revokes access within 1 second -- zero stale public access.
- **SC-005**: The showcase page scores at least 90 on Lighthouse mobile usability.
- **SC-006**: Zero data leakage: no owner identity, private settings, or cross-collection data visible on the showcase page.

## Clarifications

### Session 2026-04-10

- Q: How should the showcase page be rendered (server HTML, separate Vue SPA, or same app)? -> A: Server-rendered HTML via Go templates. Zero JavaScript for visitors. Fast, lightweight, SEO-friendly.

## Assumptions

- The showcase page is served by the same backend that serves the API. A dedicated `/showcase/{slug}` route serves the gallery (either server-rendered HTML or a dedicated client-side page).
- Showcase metadata (slug, moduleId, tenantId, enabled) is stored in a `showcases` table in the database. One row per public collection.
- The slug format is `{module-name-slugified}-{short-random-suffix}` (e.g., "rare-coins-x7k9"). The random suffix prevents URL guessing.
- Image access for showcase pages uses the same media serving paths (thumbnails/originals). No separate CDN or public bucket is needed -- the existing media routes already work without authentication.
- The showcase page is server-rendered HTML via Go templates -- zero JavaScript required for visitors. This keeps the page fast, lightweight, SEO-friendly, and truly standalone. The gallery grid and item detail overlay are built with pure HTML and CSS.
- Showcase analytics (view counts, visitor tracking) are out of scope for v1.
- Custom domains for showcase URLs are out of scope for v1.
- The showcase page inherits the application's visual design language (Outfit + Instrument Serif fonts, theme variables) but does not require the full application bundle.
