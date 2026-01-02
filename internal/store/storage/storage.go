// internal/storage/storage.go
package storage

import (
	"context"
	"io"
)

type Storage interface {
	Exists(ctx context.Context, key string) (bool, error)
	Get(ctx context.Context, key string) (io.ReadCloser, string, error)
	Put(ctx context.Context, key string, contentType string, body []byte) error
	PublicURL(key string) string
}
