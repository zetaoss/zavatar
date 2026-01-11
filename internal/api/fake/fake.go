// internal/api/fake/fake.go
package fake

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"

	"github.com/zetaoss/zavatar/internal/domain"
)

type API struct{}

func New() *API { return &API{} }

func (a *API) Get(ctx context.Context, userID int64) (*domain.UserProfile, error) {
	slog.Info("fake api get", "user_id", userID)

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

func fake32(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:32]
}
