// internal/store/db/mysql/mysql_test.go
// internal/store/db/mysql/mysql_test.go
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestMapProfileType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   int64
		want string
	}{
		{1, "letter"},
		{2, "identicon"},
		{3, "gravatar"},
		{0, "letter"},
		{-1, "letter"},
		{999, "letter"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("t=%d", tc.in), func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, mapProfileType(tc.in))
		})
	}
}

func TestDB_Get_SQLMock_ProfileExists(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	store := &DB{db: sqlDB}

	const q = `
SELECT u.user_name, p.t, p.ghash FROM user u
LEFT JOIN profiles p ON p.user_id = u.user_id
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
	require.Equal(t, "identicon", p.Type)
	require.Equal(t, "abcd1234", p.GHash)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Get_SQLMock_ProfileMissing_Defaults(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	store := &DB{db: sqlDB}

	const q = `
SELECT u.user_name, p.t, p.ghash FROM user u
LEFT JOIN profiles p ON p.user_id = u.user_id
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
	require.Equal(t, "letter", p.Type)
	require.Equal(t, "", p.GHash)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Get_SQLMock_UserMissing_ReturnsNil(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	store := &DB{db: sqlDB}

	const q = `
SELECT u.user_name, p.t, p.ghash FROM user u
LEFT JOIN profiles p ON p.user_id = u.user_id
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
