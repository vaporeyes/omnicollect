// ABOUTME: MediaStore interface for image storage abstraction.
// ABOUTME: Implemented by LocalMediaStore (filesystem) and S3MediaStore (cloud).
package storage

// MediaStore defines operations for storing and retrieving media files.
// Implementations: LocalMediaStore (local filesystem) and S3MediaStore (S3-compatible).
type MediaStore interface {
	SaveOriginal(filename string, data []byte) error
	SaveThumbnail(filename string, data []byte) error
	OriginalURL(filename string) string
	ThumbnailURL(filename string) string
}
