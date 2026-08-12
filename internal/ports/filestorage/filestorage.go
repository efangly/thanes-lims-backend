package filestorage

import (
	"context"
	"io"
	"time"
)

// FileStorage is the port for binary object storage (documents, reports).
// Implemented against MinIO/S3, swappable to any S3-compatible provider
// without touching domain/application code.
type FileStorage interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}
