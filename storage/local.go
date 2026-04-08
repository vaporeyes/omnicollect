// ABOUTME: LocalMediaStore implements MediaStore using the local filesystem.
// ABOUTME: Stores originals and thumbnails under ~/.omnicollect/media/.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// LocalMediaStore stores images on the local filesystem.
type LocalMediaStore struct {
	baseDir string
}

// NewLocalMediaStore creates a LocalMediaStore at ~/.omnicollect/media/.
func NewLocalMediaStore() (*LocalMediaStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolving home dir: %w", err)
	}
	baseDir := filepath.Join(home, ".omnicollect", "media")

	origDir := filepath.Join(baseDir, "originals")
	thumbDir := filepath.Join(baseDir, "thumbnails")
	if err := os.MkdirAll(origDir, 0755); err != nil {
		return nil, fmt.Errorf("creating originals dir: %w", err)
	}
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		return nil, fmt.Errorf("creating thumbnails dir: %w", err)
	}

	return &LocalMediaStore{baseDir: baseDir}, nil
}

// NewLocalMediaStoreAt creates a LocalMediaStore at a specific base directory.
// Used for testing with temp directories.
func NewLocalMediaStoreAt(baseDir string) *LocalMediaStore {
	return &LocalMediaStore{baseDir: baseDir}
}

// BaseDir returns the base media directory for direct file serving.
func (m *LocalMediaStore) BaseDir() string {
	return m.baseDir
}

// SaveOriginal writes original image bytes to the originals directory.
func (m *LocalMediaStore) SaveOriginal(filename string, data []byte) error {
	path := filepath.Join(m.baseDir, "originals", filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing original: %w", err)
	}
	return nil
}

// SaveThumbnail writes thumbnail image bytes to the thumbnails directory.
func (m *LocalMediaStore) SaveThumbnail(filename string, data []byte) error {
	path := filepath.Join(m.baseDir, "thumbnails", filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing thumbnail: %w", err)
	}
	return nil
}

// OriginalURL returns the URL path for serving an original image.
func (m *LocalMediaStore) OriginalURL(filename string) string {
	return "/originals/" + filename
}

// ThumbnailURL returns the URL path for serving a thumbnail image.
func (m *LocalMediaStore) ThumbnailURL(filename string) string {
	return "/thumbnails/" + filename
}
