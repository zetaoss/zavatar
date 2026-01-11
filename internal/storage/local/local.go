// internal/storage/local/local.go
package local

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/zetaoss/zavatar/internal/storage"
)

const (
	bucketDir  = "./bucket"
	publicBase = "http://localhost:8080"
)

type Storage struct{}

func New() *Storage {
	cwd, _ := os.Getwd()
	slog.Info("local storage", "cwd", cwd, "bucketDir", bucketDir)
	_ = os.MkdirAll(bucketDir, 0o755)
	return &Storage{}
}

func (s *Storage) path(key string) string {
	key = strings.TrimLeft(key, "/")

	clean := filepath.Clean(filepath.FromSlash(key))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return filepath.Join(bucketDir, "__invalid__")
	}
	return filepath.Join(bucketDir, clean)
}

func (s *Storage) Put(ctx context.Context, key string, _ string, body []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	path := s.path(key)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, body, 0o644); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

func (s *Storage) Exists(ctx context.Context, key string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	path := s.path(key)

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *Storage) PublicURL(key string) (string, error) {
	key = strings.TrimLeft(key, "/")
	return publicBase + "/" + key, nil
}

var _ storage.Storage = (*Storage)(nil)
