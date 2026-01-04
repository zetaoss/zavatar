// internal/service/avatar_service.go
package service

import (
	"context"
	"fmt"
	"log"

	"github.com/zetaoss/zavatar/internal/domain"
	"github.com/zetaoss/zavatar/internal/render"
	dbstore "github.com/zetaoss/zavatar/internal/store/db"
	storagestore "github.com/zetaoss/zavatar/internal/store/storage"
)

type AvatarService struct {
	storage  storagestore.Storage
	db       dbstore.DB
	siteSalt string
}

func NewAvatarService(storage storagestore.Storage, db dbstore.DB, siteSalt string) *AvatarService {
	return &AvatarService{
		storage:  storage,
		db:       db,
		siteSalt: siteSalt,
	}
}

type ResolveInput struct {
	UserID int64
	Size   int
}

type ResolveOutput struct {
	RedirectURL string
}

func (s *AvatarService) Resolve(ctx context.Context, in ResolveInput) (*ResolveOutput, error) {
	if in.UserID <= 0 {
		return nil, fmt.Errorf("bad user_id")
	}

	p, err := s.db.Get(ctx, in.UserID)
	if err != nil {
		log.Printf("avatar service: resolve: db.Get: %v", err)
		return nil, fmt.Errorf("avatar service: resolve: db.Get: %w", err)
	}
	if p == nil {
		p = &domain.UserProfile{Name: fmt.Sprintf("u%d", in.UserID), Type: "identicon"}
	}

	if p.Type == "gravatar" && p.GHash != "" {
		return &ResolveOutput{RedirectURL: render.GravatarURL(p.GHash, in.Size)}, nil
	}

	if p.Type == "letter" {
		key := domain.KeyLetterSVG(in.UserID)
		s.storage.Ensure(ctx, key, "image/svg+xml; charset=utf-8", func() ([]byte, error) {
			return render.LetterSVG(s.siteSalt, p.Name), nil
		})
		return &ResolveOutput{RedirectURL: s.storage.PublicURL(key)}, nil
	}

	key := domain.KeyIdenticonPNG(in.UserID, in.Size)
	s.storage.Ensure(ctx, key, "image/png", func() ([]byte, error) {
		return render.IdenticonPNG(s.siteSalt, in.UserID, in.Size)
	})

	return &ResolveOutput{RedirectURL: s.storage.PublicURL(key)}, nil
}
