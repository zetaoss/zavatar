// internal/app/router.go
package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/zetaoss/zavatar/internal/handler"
	"github.com/zetaoss/zavatar/internal/zlog"
)

func router(avatarH *handler.AvatarHandler, purgeH *handler.PurgeHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(zlog.Middleware)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/u/{user_id}", avatarH.GetAvatar)

	r.Route("/internal", func(r chi.Router) {
		r.Post("/purge/u/{user_id}", purgeH.PurgeUser)
	})

	return r
}
