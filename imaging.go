// ABOUTME: Image processing for thumbnail generation and original archival.
// ABOUTME: Returns processed bytes for callers to persist via MediaStore.
package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

// ProcessImageResult contains metadata about a processed image.
// Defined here to keep imaging self-contained; re-exported from models.go.
type processedImageData struct {
	Filename      string
	OriginalFile  string
	ThumbFile     string
	OriginalBytes []byte
	ThumbBytes    []byte
	Width         int
	Height        int
	Format        string
}

// validateImage checks that a file is a supported image by reading its header.
// Returns the detected format string ("jpeg", "png", "gif", "webp") and dimensions.
func validateImage(path string) (format string, width, height int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, 0, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	cfg, fmt_str, err := image.DecodeConfig(f)
	if err != nil {
		return "", 0, 0, fmt.Errorf("not a valid image: %w", err)
	}

	supported := map[string]bool{"jpeg": true, "png": true, "gif": true, "webp": true}
	if !supported[fmt_str] {
		return "", 0, 0, fmt.Errorf("unsupported image format: %s", fmt_str)
	}

	return fmt_str, cfg.Width, cfg.Height, nil
}

// maxImageFileSize is the maximum file size (in bytes) accepted for
// image processing. Larger files would cause excessive memory usage
// during full-image decode. 30 MB covers high-res DSLR JPEGs; users
// with larger TIFFs should convert before importing.
const maxImageFileSize = 30 * 1024 * 1024

// processImageToBytes validates an image, reads the original bytes, and generates
// thumbnail bytes in memory. The caller persists via MediaStore.
func processImageToBytes(sourcePath string) (processedImageData, error) {
	// Check file size before expensive decode
	info, err := os.Stat(sourcePath)
	if err != nil {
		return processedImageData{}, fmt.Errorf("reading file info: %w", err)
	}
	if info.Size() > maxImageFileSize {
		sizeMB := info.Size() / (1024 * 1024)
		return processedImageData{}, fmt.Errorf(
			"image too large (%d MB). Maximum supported size is %d MB",
			sizeMB, maxImageFileSize/(1024*1024))
	}

	format, width, height, err := validateImage(sourcePath)
	if err != nil {
		return processedImageData{}, err
	}

	// Generate UUID-based filename
	id := uuid.New().String()
	// Use the same filename for both original and thumbnail so the frontend
	// can reference a single filename for both /originals/ and /thumbnails/ paths.
	// The original bytes are stored as-is regardless of the .jpg extension.
	filename := id + ".jpg"
	origFilename := filename
	thumbFilename := filename

	// Read original bytes
	origBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return processedImageData{}, fmt.Errorf("reading original: %w", err)
	}

	// Generate thumbnail bytes in memory
	thumbBytes, err := generateThumbnailBytes(sourcePath)
	if err != nil {
		return processedImageData{}, fmt.Errorf("generating thumbnail: %w", err)
	}

	return processedImageData{
		Filename:      thumbFilename,
		OriginalFile:  origFilename,
		ThumbFile:     thumbFilename,
		OriginalBytes: origBytes,
		ThumbBytes:    thumbBytes,
		Width:         width,
		Height:        height,
		Format:        format,
	}, nil
}

// generateThumbnailBytes creates a 300x300 center-cropped JPEG thumbnail and returns the bytes.
func generateThumbnailBytes(srcPath string) ([]byte, error) {
	src, err := imaging.Open(srcPath, imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("opening image: %w", err)
	}

	thumb := imaging.Fill(src, 300, 300, imaging.Center, imaging.Lanczos)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 80}); err != nil {
		return nil, fmt.Errorf("encoding thumbnail: %w", err)
	}
	return buf.Bytes(), nil
}

// copyFile copies a file from src to dst. Retained for backup and other uses.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
