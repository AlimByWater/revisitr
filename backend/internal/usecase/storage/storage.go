package storage

import (
	"context"
	"io"
)

// FileInfo contains metadata about an uploaded file.
type FileInfo struct {
	Key         string `json:"key"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// Storage abstracts file storage operations.
type Storage interface {
	Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) (*FileInfo, error)
	Delete(ctx context.Context, bucket, key string) error
	GetURL(ctx context.Context, bucket, key string) (string, error)
}
