// internal/config/config.go
package config

type Config struct {
	Dev  bool
	Addr string

	Store StoreConfig
	DB    DBConfig
}

type StoreConfig struct {
	Driver string // "file" | "r2"
	R2     R2StoreConfig
}

type R2StoreConfig struct {
	AccountID       string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Prefix          string
	PublicBase      string // required for redirect
}

type DBConfig struct {
	Driver string // "memory" | "mariadb"
	// MariaDB settings later...
}
