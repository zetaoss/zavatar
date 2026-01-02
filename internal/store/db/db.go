// internal/store/db/db.go
package db

import (
	"context"

	"github.com/zetaoss/zavatar/internal/domain"
)

type DB interface {
	Get(ctx context.Context, userID int64) (*domain.UserProfile, error)
	Close() error
}
