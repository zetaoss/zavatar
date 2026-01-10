// cmd/zavatar/main.go
package main

import (
	"log/slog"
	"os"

	"github.com/zetaoss/zavatar/internal/app"
)

var Version = "dev"

func main() {
	if err := app.Run(app.Config{
		Args:    os.Args[1:],
		Version: Version,
	}); err != nil {
		slog.Error("application exited with error", "err", err)
		os.Exit(1)
	}
}
