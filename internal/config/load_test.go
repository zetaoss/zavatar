// internal/config/load_test.go
package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load(nil)
	require.NoError(t, err)

	require.Equal(t, ":8080", cfg.Addr)
	require.Equal(t, "example.com", cfg.SiteSalt)

	require.Equal(t, "local", cfg.Storage.Driver)
	require.Equal(t, "fake", cfg.API.Mode)
}

func TestLoad_TrimsAndNormalizes(t *testing.T) {
	cfg, err := Load([]string{
		"-addr", "  :9999  ",
		"-site-salt", "  hello  ",
		"-storage-driver", "  local  ",
	})
	require.NoError(t, err)

	require.Equal(t, ":9999", cfg.Addr)
	require.Equal(t, "hello", cfg.SiteSalt)
}

func TestLoad_R2_Directory(t *testing.T) {
	cfg, err := Load([]string{
		"-storage-driver", "r2",
		"-r2-bucket", "bucket",
		"-r2-account-id", "acc",
		"-r2-access-key", "ak",
		"-r2-secret-key", "sk",
		"-r2-directory", "/dev/",
		"-r2-public-base", "https://cdn.example.com",
	})
	require.NoError(t, err)

	require.Equal(t, "r2", cfg.Storage.Driver)
	require.Equal(t, "dev", cfg.Storage.R2.Directory)
}

func TestLoad_R2_DirectoryEmptyStaysEmpty(t *testing.T) {
	_, err := Load([]string{
		"-storage-driver", "r2",
		"-r2-bucket", "bucket",
		"-r2-account-id", "acc",
		"-r2-access-key", "ak",
		"-r2-secret-key", "sk",
		"-r2-directory", "   ",
		"-r2-public-base", "https://cdn.example.com",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "r2: missing R2_DIRECTORY")
}

func TestLoad_R2_ValidateMissingFields(t *testing.T) {
	t.Run("missing account id", func(t *testing.T) {
		_, err := Load([]string{
			"-storage-driver", "r2",
			"-r2-bucket", "bucket",
			"-r2-access-key", "ak",
			"-r2-secret-key", "sk",
			"-r2-directory", "dev",
			"-r2-public-base", "https://cdn.example.com",
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
			"-r2-directory", "dev",
			"-r2-public-base", "https://cdn.example.com",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "r2: missing R2_BUCKET")
	})

	t.Run("missing access key", func(t *testing.T) {
		_, err := Load([]string{
			"-storage-driver", "r2",
			"-r2-bucket", "bucket",
			"-r2-account-id", "acc",
			"-r2-secret-key", "sk",
			"-r2-directory", "dev",
			"-r2-public-base", "https://cdn.example.com",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "r2: missing R2_ACCESS_KEY")
	})

	t.Run("missing secret key", func(t *testing.T) {
		_, err := Load([]string{
			"-storage-driver", "r2",
			"-r2-bucket", "bucket",
			"-r2-account-id", "acc",
			"-r2-access-key", "ak",
			"-r2-directory", "dev",
			"-r2-public-base", "https://cdn.example.com",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "r2: missing R2_SECRET_KEY")
	})

	t.Run("missing public base", func(t *testing.T) {
		_, err := Load([]string{
			"-storage-driver", "r2",
			"-r2-bucket", "bucket",
			"-r2-account-id", "acc",
			"-r2-access-key", "ak",
			"-r2-secret-key", "sk",
			"-r2-directory", "dev",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "r2: missing R2_PUBLIC_BASE")
	})

	t.Run("missing directory", func(t *testing.T) {
		_, err := Load([]string{
			"-storage-driver", "r2",
			"-r2-bucket", "bucket",
			"-r2-account-id", "acc",
			"-r2-access-key", "ak",
			"-r2-secret-key", "sk",
			"-r2-public-base", "https://cdn.example.com",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "r2: missing R2_DIRECTORY")
	})
}

func TestLoad_InvalidDrivers(t *testing.T) {
	t.Run("invalid storage driver", func(t *testing.T) {
		_, err := Load([]string{"-storage-driver", "nope"})
		require.Error(t, err)
		require.Contains(t, err.Error(), `invalid storage driver: "nope"`)
	})

	t.Run("invalid api mode", func(t *testing.T) {
		_, err := Load([]string{"-api-mode", "nope"})
		require.Error(t, err)
		require.Contains(t, err.Error(), `invalid api mode: "nope"`)
	})
}

func TestLoad_API_RemoteRequiresKeys(t *testing.T) {
	_, err := Load([]string{
		"-api-mode", "remote",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "api: missing API_ENDPOINT")

	_, err = Load([]string{
		"-api-mode", "remote",
		"-api-endpoint", "https://api.example.com",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "api: missing API_SECRET_KEY")
}

func TestLoad_UsesEnvVars(t *testing.T) {
	t.Setenv("ADDR", "  :1234  ")
	t.Setenv("SITE_SALT", "  envsalt  ")
	t.Setenv("API_MODE", "fake")

	cfg, err := Load(nil)
	require.NoError(t, err)

	require.Equal(t, ":1234", cfg.Addr)
	require.Equal(t, "envsalt", cfg.SiteSalt)
}
