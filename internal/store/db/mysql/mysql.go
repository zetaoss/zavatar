// internal/store/db/mysql/mysql.go
package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zetaoss/zavatar/internal/domain"
)

type DB struct {
	db *sql.DB

	// raw
	database     string
	userDatabase string

	// quoted & validated (computed once in New)
	quotedDatabase     string
	quotedUserDatabase string
}

type Config struct {
	Host         string
	Port         int
	Username     string
	Password     string
	Database     string
	UserDatabase string
	Params       string
}

func New(cfg Config) (*DB, error) {
	dsn := formatDSN(cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql: ping failed: %w", err)
	}

	qdb, err := quoteIdent(cfg.Database)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql: invalid Database: %w", err)
	}
	qudb, err := quoteIdent(cfg.UserDatabase)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql: invalid UserDatabase: %w", err)
	}

	d := &DB{
		db:                 db,
		database:           cfg.Database,
		userDatabase:       cfg.UserDatabase,
		quotedDatabase:     qdb,
		quotedUserDatabase: qudb,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := d.validateSchema(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql: schema validation failed: %w", err)
	}

	return d, nil
}

func (d *DB) validateSchema(ctx context.Context) error {
	q := fmt.Sprintf(`
SELECT u.user_name, p.t, p.ghash FROM %s.user u
LEFT JOIN %s.profiles p ON p.user_id = u.user_id LIMIT 0
`, d.quotedUserDatabase, d.quotedDatabase)

	rows, err := d.db.QueryContext(ctx, q)
	if err != nil {
		return fmt.Errorf("mysql: invalid schema (user/profiles): %w", err)
	}
	defer func() { _ = rows.Close() }()

	if err := rows.Err(); err != nil {
		return fmt.Errorf("mysql: invalid schema (user/profiles): %w", err)
	}
	return nil
}

func (d *DB) Close() error { return d.db.Close() }

func (d *DB) Get(ctx context.Context, userID int64) (*domain.UserProfile, error) {
	q := fmt.Sprintf(`
SELECT u.user_name, p.t, p.ghash FROM %s.user u
LEFT JOIN %s.profiles p ON p.user_id = u.user_id
WHERE u.user_id = ? LIMIT 1
`, d.quotedUserDatabase, d.quotedDatabase)

	var (
		userName []byte
		t        sql.NullInt64
		ghash    sql.NullString
	)

	err := d.db.QueryRowContext(ctx, q, userID).Scan(&userName, &t, &ghash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("mysql: get profile user_id=%d: %w", userID, err)
	}

	tt := domain.AvatarTypeCodeIdenticon
	if t.Valid {
		tt = domain.AvatarTypeCode(t.Int64)
	}

	p := &domain.UserProfile{
		Name: string(userName),
		Type: domain.AvatarTypeFromCode(tt),
	}
	if ghash.Valid {
		p.GHash = ghash.String
	}
	return p, nil
}

func quoteIdent(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("empty identifier")
	}
	for _, r := range s {
		if r == 0 {
			return "", fmt.Errorf("identifier contains NUL")
		}
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("identifier contains control char: %q", r)
		}
	}
	escaped := strings.ReplaceAll(s, "`", "``")
	return "`" + escaped + "`", nil
}
