// internal/config/load.go
package config

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3"
)

func Load(args []string) (Config, error) {
	fs := flag.NewFlagSet("zavatar", flag.ContinueOnError)

	addr := fs.String("addr", ":8080", "listen address, e.g. :8080 (env: ADDR)")
	siteSalt := fs.String("site-salt", "example.com", "avatar site salt (env: SITE_SALT)")
	resolveMaxAge := fs.Int("resolve-max-age", 60, "env: RESOLVE_MAX_AGE")
	resolveSMaxAge := fs.Int("resolve-s-maxage", 3600, "env: RESOLVE_S_MAXAGE")

	storageDriver := fs.String("storage-driver", "local", "storage driver: local|r2 (env: STORAGE_DRIVER)")
	apiMode := fs.String("api-mode", "fake", "api mode: fake|remote (env: API_MODE)")

	r2Bucket := fs.String("r2-bucket", "", "env: R2_BUCKET")
	r2AccountID := fs.String("r2-account-id", "", "env: R2_ACCOUNT_ID")
	r2AccessKey := fs.String("r2-access-key", "", "env: R2_ACCESS_KEY")
	r2SecretKey := fs.String("r2-secret-key", "", "env: R2_SECRET_KEY")
	r2Directory := fs.String("r2-directory", "", "env: R2_DIRECTORY")
	r2PublicBase := fs.String("r2-public-base", "", "env: R2_PUBLIC_BASE (e.g. https://avatars-cdn.example.com)")

	apiEndpoint := fs.String("api-endpoint", "", "env: API_ENDPOINT")
	apiSecretKey := fs.String("api-secret-key", "", "env: API_SECRET_KEY")
	apiCacheEnabled := fs.Bool("api-cache-enabled", true, "env: API_CACHE_ENABLED")
	apiCacheDefaultExpiration := fs.Duration("api-cache-default-expiration", 5*time.Minute, "env: API_CACHE_DEFAULT_EXPIRATION")
	apiCacheCleanupInterval := fs.Duration("api-cache-cleanup-interval", 10*time.Minute, "env: API_CACHE_CLEANUP_INTERVAL")

	if err := ff.Parse(fs, args, ff.WithEnvVars()); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Addr:           strings.TrimSpace(*addr),
		SiteSalt:       strings.TrimSpace(*siteSalt),
		ResolveMaxAge:  *resolveMaxAge,
		ResolveSMaxAge: *resolveSMaxAge,
		Storage: StorageConfig{
			Driver: strings.TrimSpace(*storageDriver),
			R2: R2Config{
				AccountID:  strings.TrimSpace(*r2AccountID),
				Bucket:     strings.TrimSpace(*r2Bucket),
				AccessKey:  strings.TrimSpace(*r2AccessKey),
				SecretKey:  strings.TrimSpace(*r2SecretKey),
				Directory:  strings.TrimSpace(*r2Directory),
				PublicBase: strings.TrimSpace(*r2PublicBase),
			},
		},
		API: APIConfig{
			Mode:                   strings.TrimSpace(*apiMode),
			Endpoint:               strings.TrimSpace(*apiEndpoint),
			SecretKey:              strings.TrimSpace(*apiSecretKey),
			CacheEnabled:           *apiCacheEnabled,
			CacheDefaultExpiration: *apiCacheDefaultExpiration,
			CacheCleanupInterval:   *apiCacheCleanupInterval,
		},
	}

	normalize(&cfg)
	if err := validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func normalize(cfg *Config) {
	cfg.Addr = strings.TrimSpace(cfg.Addr)

	cfg.SiteSalt = strings.TrimSpace(cfg.SiteSalt)
	if cfg.SiteSalt == "" {
		cfg.SiteSalt = "example.com"
	}

	if p := strings.TrimSpace(cfg.Storage.R2.Directory); p != "" {
		cfg.Storage.R2.Directory = strings.Trim(p, "/")
	}

	cfg.Storage.R2.PublicBase = strings.TrimRight(strings.TrimSpace(cfg.Storage.R2.PublicBase), "/")
}

func validate(cfg Config) error {
	if cfg.ResolveMaxAge <= 0 {
		return fmt.Errorf("invalid resolve max age: %d", cfg.ResolveMaxAge)
	}
	if cfg.ResolveSMaxAge <= 0 {
		return fmt.Errorf("invalid resolve s-maxage: %d", cfg.ResolveSMaxAge)
	}

	switch cfg.Storage.Driver {
	case "local":
		// ok
	case "r2":
		r2 := cfg.Storage.R2
		if r2.AccountID == "" {
			return fmt.Errorf("r2: missing R2_ACCOUNT_ID")
		}
		if r2.Bucket == "" {
			return fmt.Errorf("r2: missing R2_BUCKET")
		}
		if r2.AccessKey == "" {
			return fmt.Errorf("r2: missing R2_ACCESS_KEY")
		}
		if r2.SecretKey == "" {
			return fmt.Errorf("r2: missing R2_SECRET_KEY")
		}
		if r2.PublicBase == "" {
			return fmt.Errorf("r2: missing R2_PUBLIC_BASE")
		}
		if r2.Directory == "" {
			return fmt.Errorf("r2: missing R2_DIRECTORY")
		}
	default:
		return fmt.Errorf("invalid storage driver: %q", cfg.Storage.Driver)
	}

	switch cfg.API.Mode {
	case "fake":
		return nil
	case "remote":
		if strings.TrimSpace(cfg.API.Endpoint) == "" {
			return fmt.Errorf("api: missing API_ENDPOINT")
		}
		if strings.TrimSpace(cfg.API.SecretKey) == "" {
			return fmt.Errorf("api: missing API_SECRET_KEY")
		}
		if cfg.API.CacheEnabled {
			if cfg.API.CacheDefaultExpiration <= 0 {
				return fmt.Errorf("api: invalid API_CACHE_DEFAULT_EXPIRATION")
			}
			if cfg.API.CacheCleanupInterval <= 0 {
				return fmt.Errorf("api: invalid API_CACHE_CLEANUP_INTERVAL")
			}
		}
		return nil
	default:
		return fmt.Errorf("invalid api mode: %q", cfg.API.Mode)
	}
}
