// internal/store/db/mysql/mysql_test.go
package mysql

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/zetaoss/zavatar/internal/domain"
)

func TestDB_Get_SQLMock_ProfileExists(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	store := &DB{
		db:                 sqlDB,
		database:           "db",
		userDatabase:       "udb",
		quotedDatabase:     "`db`",
		quotedUserDatabase: "`udb`",
	}

	const q = `
SELECT u.user_name, p.t, p.ghash FROM ` + "`udb`" + `.user u
LEFT JOIN ` + "`db`" + `.profiles p ON p.user_id = u.user_id
WHERE u.user_id = ? LIMIT 1
`

	rows := sqlmock.NewRows([]string{"user_name", "t", "ghash"}).
		AddRow([]byte("Testuser"), int64(2), "abcd1234")

	mock.ExpectQuery(regexp.QuoteMeta(q)).
		WithArgs(int64(42)).
		WillReturnRows(rows)

	p, err := store.Get(context.Background(), 42)
	require.NoError(t, err)
	require.NotNil(t, p)

	require.Equal(t, "Testuser", p.Name)
	require.Equal(t, domain.AvatarTypeLetter, p.Type)
	require.Equal(t, "abcd1234", p.GHash)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Get_SQLMock_ProfileMissing_Defaults(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	store := &DB{
		db:                 sqlDB,
		database:           "db",
		userDatabase:       "udb",
		quotedDatabase:     "`db`",
		quotedUserDatabase: "`udb`",
	}

	const q = `
SELECT u.user_name, p.t, p.ghash FROM ` + "`udb`" + `.user u
LEFT JOIN ` + "`db`" + `.profiles p ON p.user_id = u.user_id
WHERE u.user_id = ? LIMIT 1
`

	rows := sqlmock.NewRows([]string{"user_name", "t", "ghash"}).
		AddRow([]byte("NoProfileUser"), nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(q)).
		WithArgs(int64(43)).
		WillReturnRows(rows)

	p, err := store.Get(context.Background(), 43)
	require.NoError(t, err)
	require.NotNil(t, p)

	require.Equal(t, "NoProfileUser", p.Name)
	require.Equal(t, domain.AvatarTypeIdenticon, p.Type)
	require.Equal(t, "", p.GHash)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Get_SQLMock_UserMissing_ReturnsNil(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	store := &DB{
		db:                 sqlDB,
		database:           "db",
		userDatabase:       "udb",
		quotedDatabase:     "`db`",
		quotedUserDatabase: "`udb`",
	}

	const q = `
SELECT u.user_name, p.t, p.ghash FROM ` + "`udb`" + `.user u
LEFT JOIN ` + "`db`" + `.profiles p ON p.user_id = u.user_id
WHERE u.user_id = ? LIMIT 1
`

	mock.ExpectQuery(regexp.QuoteMeta(q)).
		WithArgs(int64(404)).
		WillReturnError(sql.ErrNoRows)

	p, err := store.Get(context.Background(), 404)
	require.NoError(t, err)
	require.Nil(t, p)

	require.NoError(t, mock.ExpectationsWereMet())
}
