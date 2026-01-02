// internal/app/wire_db.go
package app

import (
	"github.com/zetaoss/zavatar/internal/config"
	"github.com/zetaoss/zavatar/internal/store/db"
	"github.com/zetaoss/zavatar/internal/store/db/fake"
	"github.com/zetaoss/zavatar/internal/store/db/mysql"
)

func wireDB(cfg config.DBConfig) (db.DB, error) {
	switch cfg.Driver {
	case "mysql":
		return mysql.New(mysql.Config{
			Host:     cfg.MySQL.Host,
			Port:     cfg.MySQL.Port,
			Username: cfg.MySQL.Username,
			Password: cfg.MySQL.Password,
			Database: cfg.MySQL.Database,
		})

	default: // fake
		return fake.New(), nil
	}
}
