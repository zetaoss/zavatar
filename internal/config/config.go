// internal/config/config.go
package config

type Config struct {
	Addr     string
	SiteSalt string
	Storage  StorageConfig
	API      APIConfig
}

type StorageConfig struct {
	Driver string
	R2     R2Config
}

type R2Config struct {
	Bucket     string
	AccountID  string
	AccessKey  string
	SecretKey  string
	Directory  string
	PublicBase string
}

type APIConfig struct {
	Mode      string
	Endpoint  string
	SecretKey string
}
