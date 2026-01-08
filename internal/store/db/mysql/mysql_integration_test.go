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

	container, dsn := startMariaDB(t, ctx)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`DROP TABLE IF EXISTS profiles`)
	require.NoError(t, err)
	_, err = db.Exec("DROP TABLE IF EXISTS `user`")
	require.NoError(t, err)

	_, err = db.Exec(`
CREATE TABLE user (
  user_id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  user_name VARBINARY(255) NOT NULL DEFAULT '',
  PRIMARY KEY (user_id),
  UNIQUE KEY user_name (user_name)
) ENGINE=InnoDB DEFAULT CHARSET=binary`)
	require.NoError(t, err)

	_, err = db.Exec(`
CREATE TABLE profiles (
  user_id INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  t TINYINT(3) UNSIGNED NOT NULL DEFAULT 1,
  ghash VARCHAR(32) DEFAULT NULL,
  PRIMARY KEY (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO user (user_id, user_name) VALUES (42, 'Testuser')`)
	require.NoError(t, err)

	store := &DB{db: db}

	p, err := store.Get(ctx, 42)
	require.NoError(t, err)
	require.NotNil(t, p)
	require.Equal(t, "Testuser", p.Name)
	require.Equal(t, "letter", p.Type)
	require.Equal(t, "", p.GHash)

	_, err = db.Exec(`INSERT INTO user (user_id, user_name) VALUES (43, 'HasProfile')`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO profiles (user_id, t, ghash) VALUES (43, 2, 'abcd1234abcd1234abcd1234abcd1234')`)
	require.NoError(t, err)

	p2, err := store.Get(ctx, 43)
	require.NoError(t, err)
	require.NotNil(t, p2)
	require.Equal(t, "HasProfile", p2.Name)
	require.Equal(t, "identicon", p2.Type)
	require.Equal(t, "abcd1234abcd1234abcd1234abcd1234", p2.GHash)
}

func startMariaDB(t *testing.T, ctx context.Context) (tc.Container, string) {
	t.Helper()

	req := tc.ContainerRequest{
		Image:        "mariadb:11",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MARIADB_ROOT_PASSWORD": "root",
			"MARIADB_DATABASE":      "testdb",
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

	dsn := formatDSN("root", "root", host, port.Int(), "testdb")
	return container, dsn
}
