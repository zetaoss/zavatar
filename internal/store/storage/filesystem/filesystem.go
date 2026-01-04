// internal/store/storage/filesystem/filesystem.go
package filesystem

import (
	"context"
	"io"
	"log"
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

func (s *Storage) Ensure(ctx context.Context, key string, contentType string, gen func() ([]byte, error)) {
	if _, err := os.Stat(s.path(key)); err == nil {
		return
	} else if !os.IsNotExist(err) {
		log.Printf("storage.ensure: stat error: %v (key=%s)", err, key)
	}

	body, err := gen()
	if err != nil {
		log.Printf("storage.ensure: gen error: %v (key=%s)", err, key)
		return
	}

	if err := s.Put(ctx, key, contentType, body); err != nil {
		log.Printf("storage.ensure: put error: %v (key=%s)", err, key)
	}
}
