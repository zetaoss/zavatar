// internal/app/app.go
package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zetaoss/zavatar/internal/config"
	"github.com/zetaoss/zavatar/internal/handler"
	"github.com/zetaoss/zavatar/internal/infra/cloudflare"
	"github.com/zetaoss/zavatar/internal/service"
	"github.com/zetaoss/zavatar/internal/zlog"
)

type Config struct {
	Args    []string
	Version string
}

func Run(c Config) error {
	zlog.Init()

	slog.Info("zavatar start", "version", c.Version)

	cfg, err := config.Load(c.Args)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := wireStorage(ctx, cfg.Storage)
	if err != nil {
		return err
	}
	db, err := wireDB(cfg.DB)
	if err != nil {
		return err
	}

	cfPurger := cloudflare.NewClient(cfg.Cloudflare.ZoneID, cfg.Cloudflare.APIToken)

	avatarSvc := service.NewAvatarService(st, db, cfg.SiteSalt, cfg.BaseURL, cfPurger)
	defer avatarSvc.Close()

	avatarH := handler.NewAvatarHandler(avatarSvc)
	purgeH := handler.NewPurgeHandler(avatarSvc, cfg.PurgeKey)

	h := router(avatarH, purgeH)

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

	slog.Info("listening", "addr", cfg.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
