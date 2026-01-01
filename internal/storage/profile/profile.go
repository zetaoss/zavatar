// internal/storage/profile/profile.go
package profile

import (
	"context"

	"github.com/zetaoss/zavatar/internal/domain"
)

type Store interface {
	Get(ctx context.Context, userID int64) (*domain.UserProfile, error)
}
