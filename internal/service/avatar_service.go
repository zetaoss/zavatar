// internal/service/avatar_service.go
package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/zetaoss/zavatar/internal/domain"
	"github.com/zetaoss/zavatar/internal/render"
	"github.com/zetaoss/zavatar/internal/store/db"
	"github.com/zetaoss/zavatar/internal/store/storage"
)

type AvatarService struct {
	storage  storage.Storage
	db       db.DB
	siteSalt string

	sf singleflight.Group

	httpClient *http.Client
}

func NewAvatarService(st storage.Storage, d db.DB, siteSalt string) *AvatarService {
	return &AvatarService{
		storage:  st,
		db:       d,
		siteSalt: siteSalt,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

type ResolveInput struct {
	UserID int64
	Size   int
	T      int
}

type ResolveOutput struct {
	PNG          []byte
	CacheControl string
}

func (s *AvatarService) Resolve(ctx context.Context, in ResolveInput) (*ResolveOutput, error) {
	if in.UserID <= 0 {
		return nil, fmt.Errorf("bad user_id")
	}

	sizeEff := domain.NormalizeSizeInt(in.Size)
	isPreview := in.T != 0

	prof, err := s.db.Get(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	typ := strings.TrimSpace(prof.Type)
	if isPreview {
		typ = mapProfileType(in.T)
	}
	if typ == "" {
		typ = "letter"
	}

	if isPreview {
		sfKey := fmt.Sprintf("preview|t=%d|type=%s|uid=%d|s=%d", in.T, typ, in.UserID, sizeEff)
		b, err := s.sfBytes(sfKey, func() ([]byte, error) {
			return s.renderAt(ctx, prof, typ, in.UserID, sizeEff)
		})
		if err != nil {
			return nil, err
		}
		return &ResolveOutput{
			PNG:          b,
			CacheControl: "public, max-age=60",
		}, nil
	}

	officialKey := domain.KeyAvatar(in.UserID, sizeEff)
	if b, err := s.storage.Get(ctx, officialKey); err == nil {
		return &ResolveOutput{
			PNG:          b,
			CacheControl: "public, max-age=31536000, immutable",
		}, nil
	} else {
		if !storage.IsNotFound(err) {
			return nil, err
		}
	}

	sfKey := fmt.Sprintf("official|type=%s|uid=%d|s=%d", typ, in.UserID, sizeEff)
	b, err := s.sfBytes(sfKey, func() ([]byte, error) {
		if bb, e := s.storage.Get(ctx, officialKey); e == nil {
			return bb, nil
		} else {
			if !storage.IsNotFound(e) {
				return nil, e
			}
		}

		gen, e := s.renderAt(ctx, prof, typ, in.UserID, sizeEff)
		if e != nil {
			return nil, e
		}
		if e := s.storage.Put(ctx, officialKey, "image/png", gen); e != nil {
			return nil, e
		}
		return gen, nil
	})
	if err != nil {
		return nil, err
	}

	return &ResolveOutput{
		PNG:          b,
		CacheControl: "public, max-age=31536000, immutable",
	}, nil
}

func (s *AvatarService) sfBytes(key string, fn func() ([]byte, error)) ([]byte, error) {
	v, err, _ := s.sf.Do(key, func() (any, error) { return fn() })
	if err != nil {
		return nil, err
	}
	return v.([]byte), nil
}

func mapProfileType(t int) string {
	switch t {
	case 2:
		return "identicon"
	case 3:
		return "gravatar"
	case 1:
		fallthrough
	default:
		return "letter"
	}
}

func (s *AvatarService) renderAt(ctx context.Context, prof *domain.UserProfile, typ string, userID int64, size int) ([]byte, error) {
	switch typ {
	case "identicon":
		return render.IdenticonPNG(s.siteSalt, userID, size)

	case "gravatar":
		gh := strings.TrimSpace(prof.GHash)
		if gh == "" {
			name := safeName(prof.Name, userID)
			return render.LetterPNG(s.siteSalt, name, size)
		}
		url := render.GravatarURL(gh, size)
		return s.fetchPNG(ctx, url)

	case "letter":
		fallthrough
	default:
		name := safeName(prof.Name, userID)
		return render.LetterPNG(s.siteSalt, name, size)
	}
}

func safeName(name string, userID int64) string {
	n := strings.TrimSpace(name)
	if n != "" {
		return n
	}
	return fmt.Sprintf("user-%d", userID)
}

func (s *AvatarService) fetchPNG(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, storage.ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch failed: status=%d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
