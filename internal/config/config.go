// internal/config/config.go
package config

type Config struct {
	Addr string

	Storage StorageConfig
	DB      DBConfig
}

type StorageConfig struct {
	Driver string // filesystem | r2
	R2     R2Config
}

type R2Config struct {
	AccountID       string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Prefix          string
	PublicBase      string
}

type DBConfig struct {
	Driver string // memory | mysql
	MySQL  MySQLConfig
}

type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}
