// ABOUTME: Image processing for thumbnail generation and original archival.
// ABOUTME: Uses disintegration/imaging (pure Go, CGO-free) for resize/crop.
package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

// mediaDir returns the base media directory path.
func mediaDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".omnicollect", "media"), nil
}

// ensureMediaDirs creates the originals and thumbnails directories if needed.
func ensureMediaDirs() (origDir, thumbDir string, err error) {
	base, err := mediaDir()
	if err != nil {
		return "", "", err
	}
	origDir = filepath.Join(base, "originals")
	thumbDir = filepath.Join(base, "thumbnails")
	if err := os.MkdirAll(origDir, 0755); err != nil {
		return "", "", fmt.Errorf("creating originals dir: %w", err)
	}
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		return "", "", fmt.Errorf("creating thumbnails dir: %w", err)
	}
	return origDir, thumbDir, nil
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

// processImage validates, copies the original, and generates a thumbnail.
func processImage(sourcePath string) (ProcessImageResult, error) {
	format, width, height, err := validateImage(sourcePath)
	if err != nil {
		return ProcessImageResult{}, err
	}

	origDir, thumbDir, err := ensureMediaDirs()
	if err != nil {
		return ProcessImageResult{}, err
	}

	// Generate UUID-based filename
	id := uuid.New().String()
	ext := filepath.Ext(sourcePath)
	if ext == "" {
		ext = "." + format
	}
	origFilename := id + strings.ToLower(ext)
	thumbFilename := id + ".jpg"

	// Copy original
	origPath := filepath.Join(origDir, origFilename)
	if err := copyFile(sourcePath, origPath); err != nil {
		return ProcessImageResult{}, fmt.Errorf("copying original: %w", err)
	}

	// Generate thumbnail
	thumbPath := filepath.Join(thumbDir, thumbFilename)
	if err := generateThumbnail(sourcePath, thumbPath); err != nil {
		os.Remove(origPath) // clean up on failure
		return ProcessImageResult{}, fmt.Errorf("generating thumbnail: %w", err)
	}

	return ProcessImageResult{
		Filename:      thumbFilename,
		OriginalPath:  origFilename,
		ThumbnailPath: thumbFilename,
		Width:         width,
		Height:        height,
		Format:        format,
	}, nil
}

// generateThumbnail creates a 300x300 center-cropped JPEG thumbnail.
func generateThumbnail(srcPath, dstPath string) error {
	src, err := imaging.Open(srcPath, imaging.AutoOrientation(true))
	if err != nil {
		return fmt.Errorf("opening image: %w", err)
	}

	thumb := imaging.Fill(src, 300, 300, imaging.Center, imaging.Lanczos)
	return imaging.Save(thumb, dstPath, imaging.JPEGQuality(80))
}

// copyFile copies a file from src to dst.
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
