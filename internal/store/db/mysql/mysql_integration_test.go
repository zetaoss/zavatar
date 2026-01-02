// internal/store/db/mysql/mysql_integration_test.go
package mysql

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDB_Get_MariaDB(t *testing.T) {
	ctx := context.Background()

	container, dsn := startMariaDB(t, ctx)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`
CREATE TABLE profiles (
  user_id BIGINT PRIMARY KEY,
  name VARCHAR(64) NOT NULL,
  t INT NOT NULL,
  ghash VARCHAR(64)
)`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO profiles (user_id, name, t, ghash) VALUES (42, 'Testuser', 2, 'abcd1234')`)
	require.NoError(t, err)

	store := &DB{db: db}
	p, err := store.Get(ctx, 42)
	require.NoError(t, err)

	require.Equal(t, "Testuser", p.Name)
	require.Equal(t, "identicon", p.Type)
	require.Equal(t, "abcd1234", p.GHash)
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

	dsn := "root:root@tcp(" + host + ":" + port.Port() + ")/testdb?parseTime=true&charset=utf8mb4"

	return container, dsn
}
