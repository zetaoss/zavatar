// internal/api/api.go
package api

import (
	"context"

	"github.com/zetaoss/zavatar/internal/domain"
)

type API interface {
	Get(ctx context.Context, userID int64) (*domain.UserProfile, error)
}
