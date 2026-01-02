// internal/store/db/mysql/mysql.go
package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/zetaoss/zavatar/internal/domain"
)

type DB struct {
	db *sql.DB
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Params   string
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
		return nil, err
	}

	return &DB{db: db}, nil
}
func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Get(ctx context.Context, userID int64) (*domain.UserProfile, error) {
	const q = `SELECT name, t, ghash FROM profiles WHERE user_id = ? LIMIT 1`
	var (
		name  string
		t     int
		ghash sql.NullString
	)

	err := d.db.QueryRowContext(ctx, q, userID).Scan(&name, &t, &ghash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("mysql: get profile user_id=%d: %w", userID, err)
	}

	p := &domain.UserProfile{
		Name: name,
		Type: mapProfileType(t),
	}

	if ghash.Valid {
		p.GHash = ghash.String
	}

	return p, nil
}

func mapProfileType(t int) string {
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
