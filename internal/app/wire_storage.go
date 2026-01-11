// internal/app/wire_storage.go
package app

import (
	"fmt"

	"github.com/zetaoss/zavatar/internal/config"
	"github.com/zetaoss/zavatar/internal/storage"
	"github.com/zetaoss/zavatar/internal/storage/local"
	"github.com/zetaoss/zavatar/internal/storage/r2"
)

type WiredStorage struct {
	St                storage.Storage
	EnableLocalStatic bool
	Prefix            string
}

func wireStorage(cfg config.StorageConfig) (WiredStorage, error) {
	switch cfg.Driver {
	case "local":
		st := local.New()
		return WiredStorage{
			St:                st,
			EnableLocalStatic: true,
			Prefix:            "v1",
		}, nil

	case "r2":
		st, err := r2.New(
			cfg.R2.AccountID,
			cfg.R2.Bucket,
			cfg.R2.AccessKey,
			cfg.R2.SecretKey,
			cfg.R2.PublicBase,
		)
		if err != nil {
			return WiredStorage{}, err
		}
		return WiredStorage{
			St:                st,
			EnableLocalStatic: false,
			Prefix:            cfg.R2.Directory,
		}, nil

	default:
		return WiredStorage{}, fmt.Errorf("invalid storage driver: %q", cfg.Driver)
	}
}
