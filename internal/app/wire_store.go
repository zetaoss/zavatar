// internal/app/wire_store.go
package app

import (
	"context"
	"fmt"

	"github.com/zetaoss/zavatar/internal/config"
	"github.com/zetaoss/zavatar/internal/storage/object"
	objectfs "github.com/zetaoss/zavatar/internal/storage/object/fs"
	objectr2 "github.com/zetaoss/zavatar/internal/storage/object/r2"
)

func wireStore(ctx context.Context, sc config.StoreConfig) (object.Store, error) {
	switch sc.Driver {
	case "file":
		return objectfs.New(), nil

	case "r2":
		return objectr2.New(ctx, objectr2.Config{
			AccountID:       sc.R2.AccountID,
			Bucket:          sc.R2.Bucket,
			AccessKeyID:     sc.R2.AccessKeyID,
			SecretAccessKey: sc.R2.SecretAccessKey,
			Prefix:          sc.R2.Prefix,
			PublicBase:      sc.R2.PublicBase,
		})

	default:
		return nil, fmt.Errorf("invalid store driver: %q", sc.Driver)
	}
}
