// internal/service/avatar_service.go
package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zetaoss/zavatar/internal/domain"
	"github.com/zetaoss/zavatar/internal/render"
	"github.com/zetaoss/zavatar/internal/storage/object"
	"github.com/zetaoss/zavatar/internal/storage/profile"
)

type AvatarService struct {
	obj   object.Store
	users profile.Store
}

func NewAvatarService(obj object.Store, users profile.Store) *AvatarService {
	return &AvatarService{obj: obj, users: users}
}

type ResolveInput struct {
	UserID int64
	Size   int
	V      int
}

type ResolveOutput struct {
	RedirectURL string
}

func (s *AvatarService) Resolve(ctx context.Context, in ResolveInput) (*ResolveOutput, error) {
	if in.UserID <= 0 {
		return nil, fmt.Errorf("bad user_id")
	}
	if in.V <= 0 {
		in.V = 1
	}

	p, err := s.users.Get(ctx, in.UserID)
	if err != nil || p == nil {
		p = &domain.UserProfile{Name: fmt.Sprintf("u%d", in.UserID), Type: "identicon"}
	}

	// gravatar: redirect
	if p.Type == "gravatar" && p.GHash != "" {
		u := render.GravatarURL(p.GHash, in.Size)
		return &ResolveOutput{RedirectURL: u}, nil
	}

	// letter: SVG
	if p.Type == "letter" {
		key := domain.KeyLetterSVG(in.V, in.UserID)
		exists, _ := s.obj.Exists(ctx, key)
		if !exists {
			body := render.LetterSVG(p.Name)
			_ = s.obj.Put(ctx, key, "image/svg+xml; charset=utf-8", body)
		}
		return &ResolveOutput{RedirectURL: s.obj.PublicURL(key)}, nil
	}

	// identicon: PNG
	key := domain.KeyPNG(in.V, in.UserID, in.Size)
	exists, _ := s.obj.Exists(ctx, key)
	if exists {
		return &ResolveOutput{RedirectURL: s.obj.PublicURL(key)}, nil
	}

	seed := "u:" + strconv.FormatInt(in.UserID, 10) + ":" + p.Name
	pngBytes, err := render.IdenticonPNG(seed, in.Size)
	if err != nil {
		return nil, err
	}
	_ = s.obj.Put(ctx, key, "image/png", pngBytes)

	return &ResolveOutput{RedirectURL: s.obj.PublicURL(key)}, nil
}
