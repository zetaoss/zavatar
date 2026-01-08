// internal/store/storage/storage.go
package storage

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("storage: not found")

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

type Storage interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, contentType string, body []byte) error
}
