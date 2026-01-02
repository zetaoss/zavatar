// internal/config/load.go
package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3"
)

func Load(args []string) (Config, error) {
	fs := flag.NewFlagSet("zavatar", flag.ContinueOnError)

	// Server
	addr := fs.String("addr", ":8080", "listen address, e.g. :8080 (env: ADDR)")

	// Drivers
	storageDriver := fs.String(
		"storage-driver",
		"filesystem",
		"storage driver: filesystem|r2 (env: STORAGE_DRIVER)",
	)
	dbDriver := fs.String(
		"db-driver",
		"memory",
		"db driver: memory|mysql (env: DB_DRIVER)",
	)

	// R2
	r2AccountID := fs.String("r2-account-id", "", "env: R2_ACCOUNT_ID")
	r2Bucket := fs.String("r2-bucket", "", "env: R2_BUCKET")
	r2AccessKey := fs.String("r2-access-key", "", "env: R2_ACCESS_KEY")
	r2SecretKey := fs.String("r2-secret-key", "", "env: R2_SECRET_KEY")
	r2Prefix := fs.String("r2-prefix", "", "env: R2_PREFIX")
	r2PublicBase := fs.String("r2-public-base", "", "env: R2_PUBLIC_BASE")

	// MySQL
	mysqlHost := fs.String("mysql-host", "", "env: MYSQL_HOST")
	mysqlPort := fs.Int("mysql-port", 3306, "env: MYSQL_PORT")
	mysqlUsername := fs.String("mysql-username", "", "env: MYSQL_USERNAME")
	mysqlPassword := fs.String("mysql-password", "", "env: MYSQL_PASSWORD")
	mysqlDatabase := fs.String("mysql-database", "", "env: MYSQL_DATABASE")
	if err := ff.Parse(
		fs,
		args,
		ff.WithEnvVarNoPrefix(),
		ff.WithEnvVarPrefix(""),
	); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Addr: strings.TrimSpace(*addr),

		Storage: StorageConfig{
			Driver: strings.TrimSpace(*storageDriver),
			R2: R2Config{
				AccountID:  strings.TrimSpace(*r2AccountID),
				Bucket:     strings.TrimSpace(*r2Bucket),
				AccessKey:  strings.TrimSpace(*r2AccessKey),
				SecretKey:  strings.TrimSpace(*r2SecretKey),
				Prefix:     strings.TrimSpace(*r2Prefix),
				PublicBase: strings.TrimSpace(*r2PublicBase),
			},
		},

		DB: DBConfig{
			Driver: strings.TrimSpace(*dbDriver),
			MySQL: MySQLConfig{
				Host:     strings.TrimSpace(*mysqlHost),
				Port:     *mysqlPort,
				Username: strings.TrimSpace(*mysqlUsername),
				Password: strings.TrimSpace(*mysqlPassword),
				Database: strings.TrimSpace(*mysqlDatabase),
			},
		},
	}

	normalize(&cfg)
	if err := validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func normalize(cfg *Config) {
	if p := strings.TrimSpace(cfg.Storage.R2.Prefix); p != "" && !strings.HasSuffix(p, "/") {
		cfg.Storage.R2.Prefix = p + "/"
	}
	cfg.Storage.R2.PublicBase =
		strings.TrimRight(strings.TrimSpace(cfg.Storage.R2.PublicBase), "/")
}

func validate(cfg Config) error {
	switch cfg.Storage.Driver {
	case "filesystem":
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
		// r2.Prefix can be empty
		if r2.PublicBase == "" {
			return fmt.Errorf("r2: missing R2_PUBLIC_BASE")
		}
	default:
		return fmt.Errorf("invalid storage driver: %q", cfg.Storage.Driver)
	}

	switch cfg.DB.Driver {
	case "memory":
		return nil
	case "mysql":
		m := cfg.DB.MySQL
		if m.Host == "" {
			return fmt.Errorf("mysql: missing required config host")
		}
		if m.Port == 0 {
			return fmt.Errorf("mysql: missing required config port")
		}
		if m.Username == "" {
			return fmt.Errorf("mysql: missing required config username")
		}
		if m.Password == "" {
			return fmt.Errorf("mysql: missing required config password")
		}
		if m.Database == "" {
			return fmt.Errorf("mysql: missing required config database")
		}
		return nil
	default:
		return fmt.Errorf("invalid db driver: %q", cfg.DB.Driver)
	}
}
