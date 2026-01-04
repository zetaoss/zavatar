// internal/app/app.go
package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zetaoss/zavatar/internal/config"
	"github.com/zetaoss/zavatar/internal/handler"
	"github.com/zetaoss/zavatar/internal/service"
)

type Config struct {
	Args    []string
	Version string
}

func Run(c Config) error {
	cfg, err := config.Load(c.Args)
	if err != nil {
		return err
	}

	log.Printf("zavatar %s", c.Version)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	storage, err := wireStorage(ctx, cfg.Storage)
	if err != nil {
		return err
	}
	db, err := wireDB(cfg.DB)
	if err != nil {
		return err
	}

	avatarSvc := service.NewAvatarService(storage, db, cfg.SiteSalt)
	avatarH := handler.NewAvatarHandler(avatarSvc)
	h := router(avatarH)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: h,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Println("listening on", cfg.Addr)
	if err := srv.ListenAndServe(); err == http.ErrServerClosed {
		return nil
	}
	return err
}
