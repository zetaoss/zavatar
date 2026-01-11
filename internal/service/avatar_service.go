// internal/service/avatar_service.go
package service

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/zetaoss/zavatar/internal/domain"
	"github.com/zetaoss/zavatar/internal/infra/cloudflare"
	"github.com/zetaoss/zavatar/internal/render"
	"github.com/zetaoss/zavatar/internal/store/db"
	"github.com/zetaoss/zavatar/internal/store/storage"
	"github.com/zetaoss/zavatar/internal/zlog"
)

const (
	cachePreview  = "public, max-age=60"
	cacheOfficial = "public, max-age=31536000, immutable"
	cacheUnstable = "public, max-age=300"

	maxRemoteBytes = 2 << 20
)

type AvatarService struct {
	storage  storage.Storage
	db       db.DB
	siteSalt string

	baseURL     string
	batchPurger *cloudflare.BatchPurger

	sf         singleflight.Group
	httpClient *http.Client
}

func NewAvatarService(st storage.Storage, d db.DB, siteSalt string, baseURL string, purger cloudflare.Purger) *AvatarService {
	baseURL = strings.TrimRight(baseURL, "/")

	bp := cloudflare.NewBatchPurger(purger)
	return &AvatarService{
		storage:     st,
		db:          d,
		siteSalt:    siteSalt,
		baseURL:     baseURL,
		batchPurger: bp,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (s *AvatarService) Close() {
	if s.batchPurger != nil {
		s.batchPurger.Close()
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

	log := zlog.Ctx(ctx)

	sizeEff := domain.NormalizeSizeInt(in.Size)
	isPreview := in.T != 0

	prof, err := s.db.Get(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("db.Get err: %w", err)
	}
	if prof == nil {
		return nil, storage.ErrNotFound
	}

	typ := domain.NormalizeAvatarType(prof.Type)

	if isPreview {
		typ = domain.AvatarTypeFromCode(domain.AvatarTypeCode(in.T))
	}

	if isPreview {
		sfKey := fmt.Sprintf(
			"preview|type=%s|uid=%d|s=%d",
			typ, in.UserID, sizeEff,
		)

		b, err := s.sfBytes(sfKey, func() ([]byte, error) {
			png, _, e := s.renderAt(ctx, prof, typ, in.UserID, sizeEff)
			return png, e
		})
		if err != nil {
			return nil, err
		}

		return &ResolveOutput{
			PNG:          b,
			CacheControl: cachePreview,
		}, nil
	}

	officialKey := domain.KeyAvatar(in.UserID, sizeEff)

	if b, err := s.storage.Get(ctx, officialKey); err == nil {
		log.Debug("storage hit")
		return &ResolveOutput{
			PNG:          b,
			CacheControl: cacheOfficial,
		}, nil
	} else if !storage.IsNotFound(err) {
		return nil, err
	}

	log.Debug("storage miss, generating")

	type genRes struct {
		png    []byte
		stable bool
	}

	sfKey := fmt.Sprintf(
		"official|type=%s|uid=%d|s=%d",
		typ, in.UserID, sizeEff,
	)

	v, err, _ := s.sf.Do(sfKey, func() (any, error) {
		if bb, e := s.storage.Get(ctx, officialKey); e == nil {
			return genRes{png: bb, stable: true}, nil
		} else if !storage.IsNotFound(e) {
			return nil, e
		}

		gen, stable, e := s.renderAt(ctx, prof, typ, in.UserID, sizeEff)
		if e != nil {
			return nil, e
		}

		if stable {
			if e := s.storage.Put(ctx, officialKey, "image/png", gen); e != nil {
				return nil, e
			}
		}

		return genRes{png: gen, stable: stable}, nil
	})
	if err != nil {
		return nil, err
	}

	r := v.(genRes)
	cc := cacheOfficial
	if !r.stable {
		cc = cacheUnstable
	}

	return &ResolveOutput{
		PNG:          r.png,
		CacheControl: cc,
	}, nil
}

func (s *AvatarService) sfBytes(key string, fn func() ([]byte, error)) ([]byte, error) {
	v, err, _ := s.sf.Do(key, func() (any, error) { return fn() })
	if err != nil {
		return nil, err
	}
	b, ok := v.([]byte)
	if !ok {
		return nil, fmt.Errorf("singleflight: unexpected type %T", v)
	}
	return b, nil
}

func (s *AvatarService) renderAt(ctx context.Context, prof *domain.UserProfile, typ domain.AvatarType, userID int64, size int) ([]byte, bool, error) {
	log := zlog.Ctx(ctx)

	switch typ {
	case domain.AvatarTypeLetter:
		name := safeName(prof.Name, userID)
		b, e := render.LetterPNG(s.siteSalt, name, size)
		return b, true, e

	case domain.AvatarTypeGravatar:
		gh := strings.TrimSpace(prof.GHash)
		if gh == "" {
			fb, fe := render.IdenticonPNG(s.siteSalt, userID, size)
			return fb, false, fe
		}

		url := render.GravatarURL(gh, size)
		b, e := s.fetchPNG(ctx, url)
		if e == nil {
			return b, true, nil
		}

		if storage.IsNotFound(e) {
			log.Debug("gravatar not found, fallback to identicon")
		} else {
			log.Warn("gravatar fetch error, fallback to identicon", "err", e)
		}

		fb, fe := render.IdenticonPNG(s.siteSalt, userID, size)
		return fb, false, fe

	case domain.AvatarTypeIdenticon:
		fallthrough
	default:
		b, e := render.IdenticonPNG(s.siteSalt, userID, size)
		return b, true, e
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
	req.Header.Set("Accept", "image/png,image/*;q=0.9,*/*;q=0.1")
	req.Header.Set("User-Agent", "zavatar/1.0")

	zlog.Ctx(ctx).Debug("fetching upstream", "url", url)

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

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		mt, _, perr := mime.ParseMediaType(ct)
		if perr == nil && !strings.HasPrefix(mt, "image/") {
			return nil, fmt.Errorf("fetch failed: content-type=%q", mt)
		}
	}

	lr := io.LimitReader(resp.Body, maxRemoteBytes+1)
	b, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > maxRemoteBytes {
		return nil, fmt.Errorf("fetch failed: body too large")
	}

	return b, nil
}

func (s *AvatarService) Purge(ctx context.Context, userID int64) (int, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("bad user_id")
	}

	prof, err := s.db.Get(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("db.Get err: %w", err)
	}
	if prof == nil {
		return 0, storage.ErrNotFound
	}

	deleted := 0
	for _, sz := range domain.PresetSizes {
		key := domain.KeyAvatar(userID, sz)
		err := s.storage.Delete(ctx, key)
		if err == nil {
			deleted++
			continue
		}
		if storage.IsNotFound(err) {
			continue
		}
		return deleted, err
	}

	if s.baseURL != "" && s.batchPurger != nil {
		prefix := fmt.Sprintf("%s/u/%d", s.baseURL, userID)
		s.batchPurger.Add(prefix)
	}

	return deleted, nil
}
