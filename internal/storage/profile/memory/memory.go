// internal/storage/profile/memory/memory.go
package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/zetaoss/zavatar/internal/domain"
)

type Store struct{}

func New() *Store { return &Store{} }

func (r *Store) Get(ctx context.Context, userID int64) (*domain.UserProfile, error) {
	switch userID % 3 {
	case 0:
		return &domain.UserProfile{
			Name:  "Gravatar User",
			Type:  "gravatar",
			GHash: fake32(fmt.Sprintf("user%d@example.com", userID)),
		}, nil
	case 1:
		return &domain.UserProfile{
			Name: "Jmnote",
			Type: "letter",
		}, nil
	default:
		return &domain.UserProfile{
			Name: fmt.Sprintf("user%d", userID),
			Type: "identicon",
		}, nil
	}
}

func fake32(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:32]
}
