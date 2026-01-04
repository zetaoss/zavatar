// internal/config/load_test.go
package config

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func hashSalt(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:3]
}

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load(nil)
	require.NoError(t, err)

	require.Equal(t, ":8080", cfg.Addr)
	require.Equal(t, "example.com", cfg.SiteSalt)
	require.Equal(t, hashSalt("example.com"), cfg.SiteSaltHash)

	require.Equal(t, "filesystem", cfg.Storage.Driver)
	require.Equal(t, "memory", cfg.DB.Driver)
}

func TestLoad_TrimsAndNormalizes(t *testing.T) {
	cfg, err := Load([]string{
		"-addr", "  :9999  ",
		"-site-salt", "  hello  ",
		"-storage-driver", "  filesystem  ",
	})
	require.NoError(t, err)

	require.Equal(t, ":9999", cfg.Addr)
	require.Equal(t, "hello", cfg.SiteSalt)
	require.Equal(t, hashSalt("hello"), cfg.SiteSaltHash)
}

func TestLoad_R2_NormalizePrefixAndPublicBase(t *testing.T) {
	cfg, err := Load([]string{
		"-storage-driver", "r2",
		"-r2-account-id", "acc",
		"-r2-bucket", "bucket",
		"-r2-access-key", "ak",
		"-r2-secret-key", "sk",
		"-r2-prefix", "dev",
		"-r2-public-base", "https://example.com/base///",
	})
	require.NoError(t, err)

	require.Equal(t, "r2", cfg.Storage.Driver)
	require.Equal(t, "dev/", cfg.Storage.R2.Prefix)
	require.Equal(t, "https://example.com/base", cfg.Storage.R2.PublicBase)
}

func TestLoad_R2_PrefixEmptyStaysEmpty(t *testing.T) {
	cfg, err := Load([]string{
		"-storage-driver", "r2",
		"-r2-account-id", "acc",
		"-r2-bucket", "bucket",
		"-r2-access-key", "ak",
		"-r2-secret-key", "sk",
		"-r2-prefix", "   ",
		"-r2-public-base", "https://example.com",
	})
	require.NoError(t, err)

	require.Equal(t, "", cfg.Storage.R2.Prefix)
}

func TestLoad_R2_ValidateMissingFields(t *testing.T) {
	t.Run("missing account id", func(t *testing.T) {
		_, err := Load([]string{
			"-storage-driver", "r2",
			"-r2-bucket", "bucket",
			"-r2-access-key", "ak",
			"-r2-secret-key", "sk",
			"-r2-public-base", "https://example.com",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "r2: missing R2_ACCOUNT_ID")
	})

	t.Run("missing bucket", func(t *testing.T) {
		_, err := Load([]string{
			"-storage-driver", "r2",
			"-r2-account-id", "acc",
			"-r2-access-key", "ak",
			"-r2-secret-key", "sk",
			"-r2-public-base", "https://example.com",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "r2: missing R2_BUCKET")
	})

	t.Run("missing access key", func(t *testing.T) {
		_, err := Load([]string{
			"-storage-driver", "r2",
			"-r2-account-id", "acc",
			"-r2-bucket", "bucket",
			"-r2-secret-key", "sk",
			"-r2-public-base", "https://example.com",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "r2: missing R2_ACCESS_KEY")
	})

	t.Run("missing secret key", func(t *testing.T) {
		_, err := Load([]string{
			"-storage-driver", "r2",
			"-r2-account-id", "acc",
			"-r2-bucket", "bucket",
			"-r2-access-key", "ak",
			"-r2-public-base", "https://example.com",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "r2: missing R2_SECRET_KEY")
	})

	t.Run("missing public base", func(t *testing.T) {
		_, err := Load([]string{
			"-storage-driver", "r2",
			"-r2-account-id", "acc",
			"-r2-bucket", "bucket",
			"-r2-access-key", "ak",
			"-r2-secret-key", "sk",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "r2: missing R2_PUBLIC_BASE")
	})
}

func TestLoad_InvalidDrivers(t *testing.T) {
	t.Run("invalid storage driver", func(t *testing.T) {
		_, err := Load([]string{"-storage-driver", "nope"})
		require.Error(t, err)
		require.Contains(t, err.Error(), `invalid storage driver: "nope"`)
	})

	t.Run("invalid db driver", func(t *testing.T) {
		_, err := Load([]string{"-db-driver", "nope"})
		require.Error(t, err)
		require.Contains(t, err.Error(), `invalid db driver: "nope"`)
	})
}

func TestLoad_MySQL_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg, err := Load([]string{
			"-db-driver", "mysql",
			"-mysql-host", "127.0.0.1",
			"-mysql-port", "3307",
			"-mysql-username", "u",
			"-mysql-password", "p",
			"-mysql-database", "d",
		})
		require.NoError(t, err)
		require.Equal(t, "mysql", cfg.DB.Driver)
		require.Equal(t, "127.0.0.1", cfg.DB.MySQL.Host)
		require.Equal(t, 3307, cfg.DB.MySQL.Port)
	})

	t.Run("missing required fields", func(t *testing.T) {
		_, err := Load([]string{
			"-db-driver", "mysql",
			"-mysql-host", "127.0.0.1",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "mysql: missing required config")
	})
}

func TestLoad_UsesEnvVars(t *testing.T) {
	t.Setenv("ADDR", "  :1234  ")
	t.Setenv("SITE_SALT", "  envsalt  ")

	cfg, err := Load(nil)
	require.NoError(t, err)

	require.Equal(t, ":1234", cfg.Addr)
	require.Equal(t, "envsalt", cfg.SiteSalt)
	require.Equal(t, hashSalt("envsalt"), cfg.SiteSaltHash)
}
