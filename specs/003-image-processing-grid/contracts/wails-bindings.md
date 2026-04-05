# Wails IPC Contracts: Image Processing

**Date**: 2026-04-05
**Feature**: 003-image-processing-grid

## ProcessImage

Processes a local image file: copies original, generates thumbnail.

**Go signature**:
```go
func (a *App) ProcessImage(sourcePath string) (ProcessImageResult, error)
```

**Frontend call**:
```typescript
import { ProcessImage } from '../../wailsjs/go/main/App'
const result = await ProcessImage("/path/to/photo.jpg")
```

**Behavior**:
- Validates the file is a supported image format (JPEG, PNG, GIF, WebP).
- Copies the original to `~/.omnicollect/media/originals/{uuid}.{ext}`.
- Generates a 300x300 JPEG thumbnail to `~/.omnicollect/media/thumbnails/{uuid}.jpg`.
- Creates media directories if they do not exist.
- Returns metadata about the processed image.
- Rejects with error if file is not a valid image or cannot be processed.

**Input**: `sourcePath` -- absolute path to the image file on disk.
**Output**: `ProcessImageResult` with filename, paths, dimensions, format.
**Error**: Invalid image, read/write failure, unsupported format.

## ProcessImageResult (Go struct)

```go
type ProcessImageResult struct {
    Filename      string `json:"filename"`
    OriginalPath  string `json:"originalPath"`
    ThumbnailPath string `json:"thumbnailPath"`
    Width         int    `json:"width"`
    Height        int    `json:"height"`
    Format        string `json:"format"`
}
```
