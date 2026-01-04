// internal/store/storage/storage.go
package storage

import (
	"context"
	"io"
)

type Storage interface {
	Get(ctx context.Context, key string) (io.ReadCloser, string, error)
	Put(ctx context.Context, key string, contentType string, body []byte) error
	PublicURL(key string) string
	Ensure(ctx context.Context, key string, contentType string, gen func() ([]byte, error))
}
