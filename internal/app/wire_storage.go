// internal/app/wire_store.go
package app

import (
	"context"

	"github.com/zetaoss/zavatar/internal/config"
	storagestore "github.com/zetaoss/zavatar/internal/store/storage"
	filesystemstorage "github.com/zetaoss/zavatar/internal/store/storage/filesystem"
	r2storage "github.com/zetaoss/zavatar/internal/store/storage/r2"
)

func wireStorage(ctx context.Context, sc config.StorageConfig) (storagestore.Storage, error) {
	switch sc.Driver {
	case "r2":
		return r2storage.New(ctx, r2storage.Config{
			AccountID:  sc.R2.AccountID,
			Bucket:     sc.R2.Bucket,
			AccessKey:  sc.R2.AccessKey,
			SecretKey:  sc.R2.SecretKey,
			Prefix:     sc.R2.Prefix,
			PublicBase: sc.R2.PublicBase,
		})

	default: // file
		return filesystemstorage.New(), nil
	}
}
