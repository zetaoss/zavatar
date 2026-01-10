// internal/store/storage/filesystem/filesystem.go
package filesystem

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/zetaoss/zavatar/internal/store/storage"
)

const (
	bucketDir = "./bucket/"
	prefix    = "prefix/"
)

type Storage struct{}

func New() *Storage {
	cwd, _ := os.Getwd()
	slog.Info("filesystem storage", "cwd", cwd, "bucketDir", bucketDir, "prefix", prefix)
	_ = os.MkdirAll(bucketDir, 0755)
	return &Storage{}
}

func (s *Storage) withPrefix(key string) string {
	if prefix == "" {
		return key
	}
	return prefix + key
}

func (s *Storage) path(key string) string {
	clean := filepath.Clean(filepath.FromSlash(key))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return filepath.Join(bucketDir, "__invalid__")
	}
	return filepath.Join(bucketDir, clean)
}

func (s *Storage) Get(_ context.Context, key string) ([]byte, error) {
	k := s.withPrefix(key)
	path := s.path(k)

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

func (s *Storage) Put(_ context.Context, key string, _ string, body []byte) error {
	k := s.withPrefix(key)
	path := s.path(k)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0644)
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	_ = ctx
	p := s.path(key)

	if err := os.Remove(p); err != nil {
		if os.IsNotExist(err) {
			return storage.ErrNotFound
		}
		return err
	}
	return nil
}
