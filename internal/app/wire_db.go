// internal/app/wire_db.go
package app

import (
	"github.com/zetaoss/zavatar/internal/config"
	"github.com/zetaoss/zavatar/internal/store/db"
	"github.com/zetaoss/zavatar/internal/store/db/memory"
	"github.com/zetaoss/zavatar/internal/store/db/mysql"
)

func wireDB(cfg config.DBConfig) (db.DB, error) {
	switch cfg.Driver {
	case "mysql":
		return mysql.New(mysql.Config{
			Host:     cfg.MySQL.Host,
			Port:     cfg.MySQL.Port,
			User:     cfg.MySQL.User,
			Password: cfg.MySQL.Password,
			Database: cfg.MySQL.Database,
		})

	default: // memory
		return memory.New(), nil
	}
}
