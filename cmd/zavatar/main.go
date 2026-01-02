// cmd/zavatar/main.go
package main

import (
	"log"
	"os"

	"github.com/zetaoss/zavatar/internal/app"
)

var Version = "dev"

func main() {
	if err := app.Run(app.Config{
		Args:    os.Args[1:],
		Version: Version,
	}); err != nil {
		log.Fatal(err)
	}
}
