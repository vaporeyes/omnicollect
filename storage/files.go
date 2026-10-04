// ABOUTME: Safe filename validation and atomic, directory-confined filesystem operations.
// ABOUTME: Shared by module, settings, media, and backup boundaries.
package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var safeModuleID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)
var safeFilename = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,254}$`)

// ValidateFilename accepts only a single portable filename, never a path.
func ValidateFilename(name string) error {
	if !safeFilename.MatchString(name) || strings.Contains(name, "..") {
		return fmt.Errorf("invalid filename")
	}
	return nil
}

// ReadFileAt prevents traversal and symlink escapes even for existing files.
func ReadFileAt(dir, name string, limit int64) ([]byte, error) {
	if err := ValidateFilename(name); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	f, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("file exceeds size limit")
	}
	return data, nil
}

// AtomicWriteFile publishes a complete file without following destination symlinks.
func AtomicWriteFile(path string, data []byte) error {
	dir, name := filepath.Dir(path), filepath.Base(path)
	if err := ValidateFilename(name); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	tmp := ".write-" + uuid.NewString()
	f, err := root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer root.Remove(tmp)
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return root.Rename(tmp, name)
}
