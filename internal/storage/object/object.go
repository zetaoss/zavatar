// internal/storage/object/object.go
package object

import (
	"context"
	"io"
)

type Store interface {
	Exists(ctx context.Context, key string) (bool, error)
	Get(ctx context.Context, key string) (io.ReadCloser, string, error)
	Put(ctx context.Context, key string, contentType string, body []byte) error
	PublicURL(key string) string
}
