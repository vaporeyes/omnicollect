# omnicollect Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-04-05

## Active Technologies
- Go 1.21+ + Wails v2, modernc.org/sqlite, github.com/google/uuid (001-core-engine-data-ipc)
- SQLite (local, embedded, CGO-free via modernc.org/sqlite) (001-core-engine-data-ipc)
- TypeScript 4.6+, Vue 3.2+ (Composition API) + Vue 3, Pinia (state management), Wails runtime (IPC) (002-dynamic-form-engine)
- N/A (frontend caches only; persistence via Go backend) (002-dynamic-form-engine)
- Go 1.25+, TypeScript 4.6+, Vue 3.2+ + disintegration/imaging (Go thumbnails), golang.org/x/image/webp (WebP decode) (003-image-processing-grid)
- Local filesystem (`~/.omnicollect/media/originals/` and `thumbnails/`) (003-image-processing-grid)

- (001-core-engine-data-ipc)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for 

## Code Style

: Follow standard conventions

## Recent Changes
- 003-image-processing-grid: Added Go 1.25+, TypeScript 4.6+, Vue 3.2+ + disintegration/imaging (Go thumbnails), golang.org/x/image/webp (WebP decode)
- 002-dynamic-form-engine: Added TypeScript 4.6+, Vue 3.2+ (Composition API) + Vue 3, Pinia (state management), Wails runtime (IPC)
- 001-core-engine-data-ipc: Added Go 1.21+ + Wails v2, modernc.org/sqlite, github.com/google/uuid


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
