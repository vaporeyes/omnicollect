// ABOUTME: Context-aware media storage with explicit tenant namespaces.
// ABOUTME: Callers must scope cloud media before reading or writing private files.
package storage

import "context"

// MediaStore reads and writes images within one namespace.
// ForTenant derives an isolated namespace from a trusted tenant identifier.
type MediaStore interface {
	CheckOriginal(ctx context.Context, filename string) error
	SaveOriginal(ctx context.Context, filename string, data []byte) error
	SaveThumbnail(ctx context.Context, filename string, data []byte) error
	GetOriginal(ctx context.Context, filename string) ([]byte, error)
	GetThumbnail(ctx context.Context, filename string) ([]byte, error)
	ForTenant(tenant string) (MediaStore, error)
}
