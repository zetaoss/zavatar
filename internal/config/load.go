// internal/config/load.go
package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func Load(args []string) (Config, error) {
	fs := flag.NewFlagSet("zavatar", flag.ContinueOnError)

	addr := fs.String("addr", "", "listen address (env: ADDR), e.g. :8080")

	store := fs.String("store", "", "store driver override: file|r2 (env: STORE)")
	db := fs.String("db", "", "db driver override: memory|mariadb (env: DB)")

	// R2 overrides
	r2AccountID := fs.String("r2-account-id", "", "env: R2_ACCOUNT_ID")
	r2Bucket := fs.String("r2-bucket", "", "env: R2_BUCKET")
	r2AccessKeyID := fs.String("r2-access-key-id", "", "env: R2_ACCESS_KEY_ID")
	r2SecretAccessKey := fs.String("r2-secret-access-key", "", "env: R2_SECRET_ACCESS_KEY")
	r2Prefix := fs.String("r2-prefix", "", "env: R2_PREFIX")
	r2PublicBase := fs.String("r2-public-base", "", "env: R2_PUBLIC_BASE")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Addr: firstNonEmpty(*addr, os.Getenv("ADDR"), ":8080"),
		Store: StoreConfig{
			Driver: firstNonEmpty(*store, os.Getenv("STORE")),
			R2: R2StoreConfig{
				AccountID:       firstNonEmpty(*r2AccountID, os.Getenv("R2_ACCOUNT_ID")),
				Bucket:          firstNonEmpty(*r2Bucket, os.Getenv("R2_BUCKET")),
				AccessKeyID:     firstNonEmpty(*r2AccessKeyID, os.Getenv("R2_ACCESS_KEY_ID")),
				SecretAccessKey: firstNonEmpty(*r2SecretAccessKey, os.Getenv("R2_SECRET_ACCESS_KEY")),
				Prefix:          firstNonEmpty(*r2Prefix, os.Getenv("R2_PREFIX")),
				PublicBase:      firstNonEmpty(*r2PublicBase, os.Getenv("R2_PUBLIC_BASE")),
			},
		},
		DB: DBConfig{
			Driver: firstNonEmpty(*db, os.Getenv("DB")),
		},
	}

	// defaults (after env/flag)
	if cfg.Store.Driver == "" {
		cfg.Store.Driver = "file"
	}
	if cfg.DB.Driver == "" {
		cfg.DB.Driver = "memory"
	}

	normalize(&cfg)
	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func normalize(cfg *Config) {
	if p := strings.TrimSpace(cfg.Store.R2.Prefix); p != "" && !strings.HasSuffix(p, "/") {
		cfg.Store.R2.Prefix = p + "/"
	}
	cfg.Store.R2.PublicBase = strings.TrimRight(strings.TrimSpace(cfg.Store.R2.PublicBase), "/")
}

func validate(cfg Config) error {
	switch cfg.Store.Driver {
	case "file":
		// ok
	case "r2":
		r2 := cfg.Store.R2
		if r2.AccountID == "" || r2.Bucket == "" || r2.AccessKeyID == "" || r2.SecretAccessKey == "" {
			return fmt.Errorf("r2: missing required config (account/bucket/access/secret)")
		}
		if r2.PublicBase == "" {
			return fmt.Errorf("r2: missing R2_PUBLIC_BASE (needed for redirect)")
		}
	default:
		return fmt.Errorf("invalid store driver: %q", cfg.Store.Driver)
	}

	switch cfg.DB.Driver {
	case "memory":
		return nil
	case "mariadb":
		return fmt.Errorf("db=mariadb not implemented yet")
	default:
		return fmt.Errorf("invalid db driver: %q", cfg.DB.Driver)
	}
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
