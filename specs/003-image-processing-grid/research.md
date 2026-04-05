# Research: Image Processing & Grid Display

**Date**: 2026-04-05
**Feature**: 003-image-processing-grid

## R1: Image Processing Library

**Decision**: Use `github.com/disintegration/imaging` for thumbnail
generation. Add `golang.org/x/image/webp` for WebP input support.

**Rationale**: `disintegration/imaging` is 100% CGO-free (compatible
with `modernc.org/sqlite`), well-maintained, and supports JPEG, PNG,
GIF, TIFF, BMP natively. WebP input requires a blank import of
`golang.org/x/image/webp` which registers the decoder with Go's
standard `image.Decode`. Thumbnails are always output as JPEG.

**Alternatives considered**:
- `nfnt/resize`: Older, fewer features, no EXIF support. Rejected.
- Go standard library only: No convenient resize/crop API. Rejected.
- CGO-based libraries (libvips bindings): Violates CGO-free constraint. Rejected.

## R2: Thumbnail Strategy

**Decision**: Use `imaging.Fill` with `imaging.Center` anchor for
300x300 square thumbnails. JPEG quality 80. Auto-orientation enabled.

**Rationale**: `Fill` scales and center-crops to produce exactly
300x300 pixels regardless of input aspect ratio. This gives a
uniform grid layout. Quality 80 on 300x300 produces 15-40KB
thumbnails (well under the 100KB requirement). `AutoOrientation(true)`
handles EXIF rotation from phone/DSLR cameras.

**Alternatives considered**:
- `imaging.Fit` (preserves full image, non-square output): Better for
  detail but creates uneven grid. Rejected for grid use case.
- `imaging.Thumbnail` (alias for Fill+Center): Same result, less
  explicit. Rejected for clarity.

**Resampling filter**: `imaging.Lanczos` for highest quality. Can
switch to `imaging.CatmullRom` if performance is a concern (200-500ms
vs 100-300ms for a 20MB input).

## R3: Image Validation

**Decision**: Use `image.DecodeConfig` to validate files before full
processing. This reads only the file header (magic bytes + dimensions)
without decoding the full image.

**Rationale**: Fast detection (< 1ms) catches non-image files and
corrupted headers before committing to the expensive full decode.
Returns the format string ("jpeg", "png", "gif", "webp") for
validation against accepted formats.

## R4: Wails AssetServer for Local Media

**Decision**: Configure `assetserver.Options.Handler` with an
`http.ServeMux` that maps `/thumbnails/` and `/originals/` to
local filesystem directories using `http.StripPrefix` +
`http.FileServer`.

**Rationale**: Wails webview blocks `file://` URLs from the
`http://wails.localhost/` origin. The `Handler` field receives
requests that the embedded `Assets` FS cannot serve (fallback
behavior). `http.FileServer` handles content-type detection, range
requests, and path traversal protection automatically.

**URL pattern** (frontend):
```html
<img :src="'/thumbnails/' + filename" loading="lazy" />
```

Resolves to `http://wails.localhost/thumbnails/filename.jpg`.
The embedded assets have no `/thumbnails/` path, so the request
falls through to the custom handler.

**Alternatives considered**:
- Middleware approach: Wraps all requests, not just media. Overkill.
  Rejected.
- Base64-encoding images in Wails runtime calls: High memory overhead,
  defeats lazy loading. Rejected.
- Custom protocol handler: Not supported in Wails v2. Rejected.

## R5: Lazy Loading

**Decision**: Use native HTML `loading="lazy"` attribute on all grid
`<img>` tags.

**Rationale**: Supported by all modern webview engines (WebKit, Blink,
WebView2). Zero JavaScript needed. The browser handles viewport
intersection detection automatically. No IntersectionObserver polyfill
or library required.

## Summary of New Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/disintegration/imaging` | Image resize, crop, JPEG encode |
| `golang.org/x/image/webp` | WebP format decoder (blank import) |
