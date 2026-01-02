// internal/store/db/mysql/mysql_test.go
package mysql

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestMapProfileType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   int
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

func TestDB_Get_SQLMock(t *testing.T) {
	t.Parallel()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	store := &DB{db: sqlDB}

	const q = `SELECT name, t, ghash FROM profiles WHERE user_id = ? LIMIT 1`

	rows := sqlmock.NewRows([]string{"name", "t", "ghash"}).
		AddRow("Testuser", 2, "abcd1234")

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
