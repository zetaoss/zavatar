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

	st, err := wireStorage(cfg.Storage)
	if err != nil {
		return err
	}
	apiClient, err := wireAPI(cfg.API)
	if err != nil {
		return err
	}

	avatarSvc := service.NewAvatarService(
		st.St,
		apiClient,
		cfg.SiteSalt,
		st.Prefix,
		cfg.ResolveMaxAge,
		cfg.ResolveSMaxAge,
	)
	avatarH := handler.NewAvatarHandler(avatarSvc)
	h := router(avatarH, st.EnableLocalStatic)

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
