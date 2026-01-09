// internal/store/db/mysql/mysql.go
package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"unicode"

	"github.com/zetaoss/zavatar/internal/domain"
)

type DB struct {
	db           *sql.DB
	database     string
	userDatabase string
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
	if cfg.Database == "" {
		return nil, fmt.Errorf("mysql: Database is required")
	}
	if cfg.UserDatabase == "" {
		return nil, fmt.Errorf("mysql: UserDatabase is required")
	}

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

	d := &DB{
		db:           db,
		database:     cfg.Database,
		userDatabase: cfg.UserDatabase,
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
	udb, err := quoteIdent(d.userDatabase)
	if err != nil {
		return fmt.Errorf("mysql: invalid UserDatabase: %w", err)
	}
	pdb, err := quoteIdent(d.database)
	if err != nil {
		return fmt.Errorf("mysql: invalid Database: %w", err)
	}

	q := fmt.Sprintf(`
SELECT u.user_name, p.t, p.ghash FROM %s.user u
LEFT JOIN %s.profiles p ON p.user_id = u.user_id LIMIT 0
`, udb, pdb)

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
	udb, err := quoteIdent(d.userDatabase)
	if err != nil {
		return nil, fmt.Errorf("mysql: invalid UserDatabase: %w", err)
	}
	pdb, err := quoteIdent(d.database)
	if err != nil {
		return nil, fmt.Errorf("mysql: invalid Database: %w", err)
	}

	q := fmt.Sprintf(`
SELECT u.user_name, p.t, p.ghash FROM %s.user u
LEFT JOIN %s.profiles p ON p.user_id = u.user_id
WHERE u.user_id = ? LIMIT 1
`, udb, pdb)

	var (
		userName []byte
		t        sql.NullInt64
		ghash    sql.NullString
	)

	err = d.db.QueryRowContext(ctx, q, userID).Scan(&userName, &t, &ghash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("mysql: get profile user_id=%d: %w", userID, err)
	}

	tt := int64(1)
	if t.Valid {
		tt = t.Int64
	}

	p := &domain.UserProfile{
		Name: string(userName),
		Type: mapProfileType(tt),
	}

	if ghash.Valid {
		p.GHash = ghash.String
	}

	return p, nil
}

func mapProfileType(t int64) string {
	switch t {
	case 1:
		return "letter"
	case 2:
		return "identicon"
	case 3:
		return "gravatar"
	default:
		return "letter"
	}
}

func quoteIdent(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("empty identifier")
	}
	for _, r := range s {
		if !isIdentChar(r) {
			return "", fmt.Errorf("invalid char in identifier: %q", r)
		}
	}
	return "`" + s + "`", nil
}

func isIdentChar(r rune) bool {
	if r > unicode.MaxASCII {
		return false
	}
	return r == '_' ||
		r == '$' ||
		(r >= '0' && r <= '9') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z')
}
