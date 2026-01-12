// internal/service/avatar_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/sync/singleflight"

	"github.com/zetaoss/zavatar/internal/api"
	"github.com/zetaoss/zavatar/internal/domain"
	"github.com/zetaoss/zavatar/internal/render"
	"github.com/zetaoss/zavatar/internal/storage"
	"github.com/zetaoss/zavatar/internal/zlog"
)

type AvatarService struct {
	storage      storage.Storage
	api          api.API
	siteSalt     string
	prefix       string
	cacheControl string
	sf           singleflight.Group
}

func NewAvatarService(st storage.Storage, a api.API, siteSalt string, prefix string, resolveMaxAge, resolveSMaxAge int) *AvatarService {
	return &AvatarService{
		storage:      st,
		api:          a,
		siteSalt:     siteSalt,
		prefix:       prefix,
		cacheControl: fmt.Sprintf("public, max-age=%d, s-maxage=%d", resolveMaxAge, resolveSMaxAge),
	}
}

type ResolveInput struct {
	UserID int64
	Size   int
	T      int
}

type ResolveOutput struct {
	RedirectURL  string
	CacheControl string
}

var ErrBadParams = errors.New("bad params")

func (s *AvatarService) Resolve(ctx context.Context, in ResolveInput) (*ResolveOutput, error) {
	if in.UserID <= 0 {
		return nil, ErrBadParams
	}

	tCode := domain.AvatarTypeCode(in.T)
	if in.T != 0 && (in.T < 1 || in.T > 3) {
		return nil, ErrBadParams
	}

	sizeEff := domain.NormalizeSizeInt(in.Size)

	prof, err := s.api.Get(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("api.Get err: %w", err)
	}
	if prof == nil {
		return nil, storage.ErrNotFound
	}

	sfKey := fmt.Sprintf("uid=%d|s=%d|t=%d", in.UserID, sizeEff, in.T)

	v, err, _ := s.sf.Do(sfKey, func() (any, error) {
		if tCode == 0 {
			tCode = domain.AvatarTypeToCode(prof.Type)
		}

		switch tCode {
		case domain.AvatarTypeCodeGravatar:
			return s.resolveGravatar(ctx, prof, in.UserID, sizeEff)

		case domain.AvatarTypeCodeLetter:
			return s.resolveStored(ctx, prof, in.UserID, sizeEff, domain.AvatarTypeCodeLetter)

		case domain.AvatarTypeCodeIdenticon:
			fallthrough
		default:
			return s.resolveStored(ctx, prof, in.UserID, sizeEff, domain.AvatarTypeCodeIdenticon)
		}
	})
	if err != nil {
		return nil, err
	}

	out, ok := v.(*ResolveOutput)
	if !ok {
		return nil, fmt.Errorf("singleflight: unexpected type %T", v)
	}
	return out, nil
}

func (s *AvatarService) resolveGravatar(ctx context.Context, prof *domain.UserProfile, userID int64, size int) (*ResolveOutput, error) {
	log := zlog.Ctx(ctx)

	gh := strings.ToLower(strings.TrimSpace(prof.GHash))
	if gh != "" {
		return &ResolveOutput{
			RedirectURL:  render.GravatarURL(gh, size),
			CacheControl: s.cacheControl,
		}, nil
	}

	if gh == "" {
		log.Info("missing ghash, fallback to identicon", "uid", userID, "s", size)
	}

	return s.resolveStored(ctx, prof, userID, size, domain.AvatarTypeCodeIdenticon)
}

func (s *AvatarService) resolveStored(ctx context.Context, prof *domain.UserProfile, userID int64, size int, t domain.AvatarTypeCode) (*ResolveOutput, error) {
	log := zlog.Ctx(ctx)

	key := domain.KeyAvatar(s.prefix, t, userID, size)

	exists, err := s.storage.Exists(ctx, key)
	if err != nil {
		log.Warn("storage exists failed", "err", err, "key", key)
		return nil, err
	}
	if !exists {
		png, err := s.renderStable(prof, userID, size, t)
		if err != nil {
			return nil, err
		}
		if err := s.storage.Put(ctx, key, "image/png", png); err != nil {
			log.Warn("storage put failed", "err", err, "key", key)
			return nil, err
		}
	}

	url, err := s.storage.PublicURL(key)
	if err != nil {
		log.Warn("storage public url failed", "err", err, "key", key)
		return nil, err
	}

	return &ResolveOutput{
		RedirectURL:  url,
		CacheControl: s.cacheControl,
	}, nil
}

func (s *AvatarService) renderStable(prof *domain.UserProfile, userID int64, size int, t domain.AvatarTypeCode) ([]byte, error) {
	switch t {
	case domain.AvatarTypeCodeLetter:
		name := safeName(prof.Name, userID)
		return render.InitialsPNG(s.siteSalt, name, size)

	case domain.AvatarTypeCodeIdenticon:
		fallthrough
	default:
		return render.IdenticonPNG(s.siteSalt, userID, size)
	}
}

func safeName(name string, userID int64) string {
	n := strings.TrimSpace(name)
	if n != "" {
		return n
	}
	return fmt.Sprintf("user-%d", userID)
}
