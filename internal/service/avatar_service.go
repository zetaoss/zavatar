// internal/service/avatar_service.go
package service

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/zetaoss/zavatar/internal/domain"
	"github.com/zetaoss/zavatar/internal/render"
	dbstore "github.com/zetaoss/zavatar/internal/store/db"
	storagestore "github.com/zetaoss/zavatar/internal/store/storage"
)

type AvatarService struct {
	storage      storagestore.Storage
	db           dbstore.DB
	siteSalt     string
	siteSaltHash string
}

func NewAvatarService(storage storagestore.Storage, db dbstore.DB, siteSalt, siteSaltHash string) *AvatarService {
	return &AvatarService{
		storage:      storage,
		db:           db,
		siteSalt:     siteSalt,
		siteSaltHash: siteSaltHash,
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
		key := domain.KeyLetterSVG(s.siteSaltHash, in.UserID)

		exists, err := s.storage.Exists(ctx, key)
		if err != nil {
			log.Printf("avatar_service: storage.Exists error: %v", err)
		}
		if !exists {
			body := render.LetterSVG(s.siteSalt, p.Name)
			if err := s.storage.Put(ctx, key, "image/svg+xml; charset=utf-8", body); err != nil {
				log.Printf("avatar_service: storage.Put error: %v", err)
			}
		}
		return &ResolveOutput{RedirectURL: s.storage.PublicURL(key)}, nil
	}

	// identicon: PNG
	key := domain.KeyIdenticonPNG(s.siteSaltHash, in.UserID, in.Size)

	exists, err := s.storage.Exists(ctx, key)
	if err != nil {
		log.Printf("avatar_service: storage.Exists error: %v", err)
	}
	if exists {
		return &ResolveOutput{RedirectURL: s.storage.PublicURL(key)}, nil
	}

	userSalt := s.siteSalt + "|" + strconv.FormatInt(in.UserID, 10)

	pngBytes, err := render.IdenticonPNG(userSalt, in.Size)
	if err != nil {
		return nil, err
	}
	if err := s.storage.Put(ctx, key, "image/png", pngBytes); err != nil {
		log.Printf("avatar_service: storage.Put error: %v", err)
	}

	return &ResolveOutput{RedirectURL: s.storage.PublicURL(key)}, nil
}
