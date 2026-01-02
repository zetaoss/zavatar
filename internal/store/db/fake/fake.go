// internal/DB/db/fake/fake.go
package fake

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/zetaoss/zavatar/internal/domain"
)

type DB struct{}

func New() *DB { return &DB{} }

func (d *DB) Get(ctx context.Context, userID int64) (*domain.UserProfile, error) {
	switch userID % 3 {
	case 0:
		return &domain.UserProfile{
			Name:  "Gravatar User",
			Type:  "gravatar",
			GHash: fake32(fmt.Sprintf("user%d@example.com", userID)),
		}, nil
	case 1:
		return &domain.UserProfile{
			Name: "Testuser",
			Type: "letter",
		}, nil
	default:
		return &domain.UserProfile{
			Name: fmt.Sprintf("user%d", userID),
			Type: "identicon",
		}, nil
	}
}

func (d *DB) Close() error {
	return nil
}

func fake32(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:32]
}
