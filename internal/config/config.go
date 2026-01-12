// internal/config/config.go
package config

import "time"

type Config struct {
	Addr           string
	SiteSalt       string
	ResolveMaxAge  int
	ResolveSMaxAge int
	Storage        StorageConfig
	API            APIConfig
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
	Mode                   string
	Endpoint               string
	SecretKey              string
	CacheEnabled           bool
	CacheDefaultExpiration time.Duration
	CacheCleanupInterval   time.Duration
}
