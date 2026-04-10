# Showcase Contract: REST API + Public Pages

**Branch**: `017-public-showcase` | **Date**: 2026-04-10

## Authenticated API (for collection owners)

### List Showcases
`GET /api/v1/showcases`

**Response**: `200 OK` -- `Showcase[]`
```json
[
  {"id": "uuid", "slug": "rare-coins-a3f7b2c1", "moduleId": "coins", "enabled": true, "createdAt": "...", "updatedAt": "..."}
]
```

### Toggle Showcase
`POST /api/v1/showcases/toggle`

**Request Body**:
```json
{"moduleId": "coins", "enabled": true}
```

**Response**: `200 OK` -- `Showcase` (created or updated)
```json
{"id": "uuid", "slug": "rare-coins-a3f7b2c1", "moduleId": "coins", "enabled": true, "url": "/showcase/rare-coins-a3f7b2c1"}
```

On first toggle to public: generates slug and creates showcase record.
On subsequent toggles: updates `enabled` field; slug remains stable.

## Public Pages (no auth, server-rendered HTML)

### Gallery Page
`GET /showcase/{slug}`

**Query Params**: `?page=2` (pagination, default page 1, 24 items per page)

**Response**: `200 OK` -- HTML page with:
- Header: collection name, item count
- Responsive grid of item cards (thumbnail, title, key attributes)
- Pagination links
- Google Fonts link for Outfit + Instrument Serif

**Errors**:
- Showcase not found or disabled: renders `unavailable.html` (200 status, friendly message)

### Item Detail (CSS :target overlay)
Item detail is shown inline on the gallery page via CSS `:target` pseudo-class. Each card links to `#item-{id}`, revealing a detail section with full-resolution image, all attributes, and tags.

No separate URL for item detail -- it's all on the gallery page.

## Template Structure

```
showcase/templates/
  gallery.html       -- Full gallery page with embedded CSS
  unavailable.html   -- "This showcase is no longer available" page
```

Templates are embedded in the Go binary via `//go:embed`.

## HTML Gallery Page Structure

```html
<!DOCTYPE html>
<html>
<head>
  <title>{{.CollectionName}} - OmniCollect Showcase</title>
  <link href="fonts.googleapis.com/..." rel="stylesheet">
  <style>/* embedded CSS using app's design variables */</style>
</head>
<body>
  <header>
    <h1>{{.CollectionName}}</h1>
    <p>{{.ItemCount}} items</p>
  </header>
  <main class="grid">
    {{range .Items}}
    <a href="#item-{{.ID}}" class="card">
      <img src="/thumbnails/{{.PrimaryImage}}" loading="lazy">
      <h3>{{.Title}}</h3>
      <span class="module-badge">{{$.CollectionName}}</span>
    </a>
    <div id="item-{{.ID}}" class="detail-overlay">
      <a href="#" class="close">&times;</a>
      <img src="/originals/{{.PrimaryImage}}">
      <h2>{{.Title}}</h2>
      <!-- attributes, tags -->
    </div>
    {{end}}
  </main>
  <nav class="pagination"><!-- page links --></nav>
</body>
</html>
```
