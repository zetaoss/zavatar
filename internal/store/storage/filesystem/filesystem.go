// internal/storage/filesystem/filesystem.go
package filesystem

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

const (
	dataDir    = "./data"
	publicBase = "http://localhost:8080"
	prefix     = "data/"
)

type Storage struct{}

func New() *Storage {
	_ = os.MkdirAll(dataDir, 0755)
	return &Storage{}
}

func (s *Storage) path(key string) string {
	return filepath.Join(dataDir, key)
}

func (s *Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := os.Stat(s.path(key))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *Storage) Get(ctx context.Context, key string) (io.ReadCloser, string, error) {
	f, err := os.Open(s.path(key))
	if err != nil {
		return nil, "", err
	}
	return f, "", nil
}

func (s *Storage) Put(ctx context.Context, key string, contentType string, body []byte) error {
	path := s.path(key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0644)
}

func (s *Storage) PublicURL(key string) string {
	return publicBase + "/" + prefix + key
}
