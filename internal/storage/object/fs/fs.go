// internal/storage/object/fs/fs.go
package fs

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

type Store struct{}

func New() *Store {
	_ = os.MkdirAll(dataDir, 0755)
	return &Store{}
}

func (s *Store) path(key string) string {
	return filepath.Join(dataDir, key)
}

func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	_, err := os.Stat(s.path(key))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, string, error) {
	f, err := os.Open(s.path(key))
	if err != nil {
		return nil, "", err
	}
	return f, "", nil
}

func (s *Store) Put(ctx context.Context, key string, contentType string, body []byte) error {
	path := s.path(key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0644)
}

func (s *Store) PublicURL(key string) string {
	return publicBase + "/" + prefix + key
}
