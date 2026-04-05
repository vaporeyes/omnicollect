# Data Model: Image Processing & Grid Display

**Date**: 2026-04-05
**Feature**: 003-image-processing-grid

## Entities

### Media File (filesystem, not database)

Images are stored as files on the local filesystem, not in SQLite.
Items reference images by filename in their `images` JSON array.

**Directory structure**:
```
~/.omnicollect/media/
  originals/     # Full-resolution copies
    {uuid}.jpg
    {uuid}.png
  thumbnails/    # 300x300 compressed JPEG
    {uuid}.jpg
```

**Naming convention**: Each processed image gets a UUID-based filename
(preserving the original extension for originals, always `.jpg` for
thumbnails).

### Item.images (existing field, new usage)

The `images` field on the Item entity (already defined in Iteration 1)
stores an array of filename strings:

```json
{
  "images": ["a1b2c3d4.jpg", "e5f6g7h8.jpg"]
}
```

Each filename corresponds to:
- `~/.omnicollect/media/originals/{filename}` (full-res original)
- `~/.omnicollect/media/thumbnails/{filename}.jpg` (300x300 thumb,
  always JPEG regardless of original format)

Note: Thumbnail filename appends `.jpg` to the base name to handle
originals that are PNG/WebP but thumbnails are always JPEG.

### ProcessImageResult (Go struct, returned by ProcessImage binding)

| Field | Type | Description |
|-------|------|-------------|
| filename | string | The generated filename (UUID-based) |
| originalPath | string | Relative path within originals/ |
| thumbnailPath | string | Relative path within thumbnails/ |
| width | int | Original image width in pixels |
| height | int | Original image height in pixels |
| format | string | Detected format ("jpeg", "png", "gif", "webp") |

## URL Mapping (Wails AssetServer)

| Frontend URL | Served from |
|-------------|-------------|
| `/thumbnails/{filename}` | `~/.omnicollect/media/thumbnails/{filename}` |
| `/originals/{filename}` | `~/.omnicollect/media/originals/{filename}` |

## Validation Rules

- Input file MUST be a valid image (detected via header bytes).
- Accepted formats: JPEG, PNG, GIF, WebP.
- Non-image files MUST be rejected with an error message.
- Corrupted images (valid header but broken content) MUST be
  reported without crashing.
- Media directories MUST be created automatically on first use.
