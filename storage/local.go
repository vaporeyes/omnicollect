// ABOUTME: LocalMediaStore implements MediaStore using the local filesystem.
// ABOUTME: Stores originals and thumbnails under ~/.omnicollect/media/.
package storage

import (
	"context"
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
func (m *LocalMediaStore) SaveOriginal(ctx context.Context, filename string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := ValidateFilename(filename); err != nil {
		return err
	}
	path := filepath.Join(m.baseDir, "originals", filename)
	if err := AtomicWriteFile(path, data); err != nil {
		return fmt.Errorf("writing original: %w", err)
	}
	return nil
}

// SaveThumbnail writes thumbnail image bytes to the thumbnails directory.
func (m *LocalMediaStore) SaveThumbnail(ctx context.Context, filename string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := ValidateFilename(filename); err != nil {
		return err
	}
	path := filepath.Join(m.baseDir, "thumbnails", filename)
	if err := AtomicWriteFile(path, data); err != nil {
		return fmt.Errorf("writing thumbnail: %w", err)
	}
	return nil
}

// CheckOriginal verifies a confined regular file without loading its contents.
func (m *LocalMediaStore) CheckOriginal(ctx context.Context, filename string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := ValidateFilename(filename); err != nil {
		return err
	}
	root, err := os.OpenRoot(filepath.Join(m.baseDir, "originals"))
	if err != nil {
		return err
	}
	defer root.Close()
	info, err := root.Stat(filename)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() == 0 || info.Size() > 30<<20 {
		return fmt.Errorf("invalid original file")
	}
	return nil
}

// GetOriginal reads a confined, size-bounded original image.
func (m *LocalMediaStore) GetOriginal(ctx context.Context, filename string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return ReadFileAt(filepath.Join(m.baseDir, "originals"), filename, 30<<20)
}

// GetThumbnail reads a confined, size-bounded thumbnail.
func (m *LocalMediaStore) GetThumbnail(ctx context.Context, filename string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return ReadFileAt(filepath.Join(m.baseDir, "thumbnails"), filename, 2<<20)
}

// ForTenant creates a filesystem namespace without mutating the base store.
func (m *LocalMediaStore) ForTenant(tenant string) (MediaStore, error) {
	if err := ValidateFilename(tenant); err != nil {
		return nil, err
	}
	return NewLocalMediaStoreAt(filepath.Join(m.baseDir, "tenants", tenant)), nil
}
