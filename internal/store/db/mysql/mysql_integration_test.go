//go:build integration

// internal/store/db/mysql/mysql_integration_test.go
package mysql

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDB_Get_MySQL_LeftJoin_Defaults(t *testing.T) {
	ctx := context.Background()

	container, dsn := startMySQL(t, ctx)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	// 1) Prepare two databases
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS pdb CHARACTER SET utf8mb4")
	require.NoError(t, err)
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS udb CHARACTER SET utf8mb4")
	require.NoError(t, err)

	// Clean tables (safe)
	_, _ = db.Exec("DROP TABLE IF EXISTS pdb.profiles")
	_, _ = db.Exec("DROP TABLE IF EXISTS udb.`user`")

	// 2) Create user table in udb
	_, err = db.Exec(`
CREATE TABLE udb.` + "`user`" + ` (
  user_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_name VARBINARY(255) NOT NULL DEFAULT '',
  PRIMARY KEY (user_id),
  UNIQUE KEY user_name (user_name)
) ENGINE=InnoDB DEFAULT CHARSET=binary`)
	require.NoError(t, err)

	// 3) Create profiles table in pdb
	_, err = db.Exec(`
CREATE TABLE pdb.profiles (
  user_id INT UNSIGNED NOT NULL,
  t TINYINT UNSIGNED NOT NULL DEFAULT 1,
  ghash VARCHAR(32) DEFAULT NULL,
  PRIMARY KEY (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	require.NoError(t, err)

	// 4) Insert user only (no profile)
	_, err = db.Exec("INSERT INTO udb.`user` (user_id, user_name) VALUES (42, 'Testuser')")
	require.NoError(t, err)

	qpdb, err := quoteIdent("pdb")
	require.NoError(t, err)
	qudb, err := quoteIdent("udb")
	require.NoError(t, err)

	store := &DB{
		db:                 db,
		database:           "pdb",
		userDatabase:       "udb",
		quotedDatabase:     qpdb,
		quotedUserDatabase: qudb,
	}

	p, err := store.Get(ctx, 42)
	require.NoError(t, err)
	require.NotNil(t, p)
	require.Equal(t, "Testuser", p.Name)
	require.Equal(t, "letter", p.Type)
	require.Equal(t, "", p.GHash)

	// 5) Insert user + profile
	_, err = db.Exec("INSERT INTO udb.`user` (user_id, user_name) VALUES (43, 'HasProfile')")
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO pdb.profiles (user_id, t, ghash) VALUES (43, 2, 'abcd1234abcd1234abcd1234abcd1234')")
	require.NoError(t, err)

	p2, err := store.Get(ctx, 43)
	require.NoError(t, err)
	require.NotNil(t, p2)
	require.Equal(t, "HasProfile", p2.Name)
	require.Equal(t, "identicon", p2.Type)
	require.Equal(t, "abcd1234abcd1234abcd1234abcd1234", p2.GHash)
}

func startMySQL(t *testing.T, ctx context.Context) (tc.Container, string) {
	t.Helper()

	req := tc.ContainerRequest{
		Image:        "mysql:8.0",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "root",
			"MYSQL_DATABASE":      "bootstrap",
		},
		WaitingFor: wait.ForListeningPort("3306/tcp").
			WithStartupTimeout(60 * time.Second),
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)

	dsn := formatDSN("root", "root", host, port.Int(), "bootstrap")
	return container, dsn
}
