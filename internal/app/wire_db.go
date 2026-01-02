package app

import (
	"fmt"

	"github.com/zetaoss/zavatar/internal/config"
	"github.com/zetaoss/zavatar/internal/storage/profile"
	profilemem "github.com/zetaoss/zavatar/internal/storage/profile/memory"
)

func wireDB(cfg config.DBConfig) (profile.Store, error) {
	switch cfg.Driver {
	case "mariadb":
		return nil, fmt.Errorf("db=mariadb not implemented yet")

	default: // memory
		return profilemem.New(), nil
	}
}
