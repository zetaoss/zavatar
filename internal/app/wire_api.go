// internal/app/wire_api.go
package app

import (
	"github.com/zetaoss/zavatar/internal/api"
	"github.com/zetaoss/zavatar/internal/api/fake"
	"github.com/zetaoss/zavatar/internal/api/remote"
	"github.com/zetaoss/zavatar/internal/config"
)

func wireAPI(cfg config.APIConfig) (api.API, error) {
	switch cfg.Mode {
	case "remote":
		return remote.New(
			cfg.Endpoint,
			cfg.SecretKey,
			cfg.CacheEnabled,
			cfg.CacheDefaultExpiration,
			cfg.CacheCleanupInterval,
		), nil

	default: // fake
		return fake.New(), nil
	}
}
