// internal/storage/storage.go
package storage

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("storage: not found")
)

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

type Storage interface {
	Put(ctx context.Context, key string, contentType string, body []byte) error
	Exists(ctx context.Context, key string) (bool, error)
	PublicURL(key string) (string, error)
}
