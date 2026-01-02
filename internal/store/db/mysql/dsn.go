package mysql

import (
	"fmt"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func formatDSN(user, pass, host string, port int, dbname string) string {
	cfg := mysqlDriver.Config{
		User:      user,
		Passwd:    pass,
		Net:       "tcp",
		Addr:      fmt.Sprintf("%s:%d", host, port),
		DBName:    dbname,
		ParseTime: true,
		Params: map[string]string{
			"allowNativePasswords": "true",
			"charset":              "utf8mb4",
		},
	}
	return cfg.FormatDSN()
}
