// internal/config/config.go
package config

type Config struct {
	Addr     string // env: ADDR
	SiteSalt string // env: SITE_SALT
	PurgeKey string // env: PURGE_KEY
	BaseURL  string // env: BASE_URL

	Cloudflare CloudflareConfig
	Storage    StorageConfig
	DB         DBConfig
}

type CloudflareConfig struct {
	ZoneID   string // env: CF_ZONE_ID
	APIToken string // env: CF_API_TOKEN
}

type StorageConfig struct {
	Driver string // filesystem | r2
	R2     R2Config
}

type R2Config struct {
	AccountID string
	Bucket    string
	AccessKey string
	SecretKey string
	Prefix    string
}

type DBConfig struct {
	Driver string // memory | mysql
	MySQL  MySQLConfig
}

type MySQLConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	Database     string
	UserDatabase string
}
