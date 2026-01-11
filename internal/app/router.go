// internal/app/router.go
package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/zetaoss/zavatar/internal/handler"
	"github.com/zetaoss/zavatar/internal/zlog"
)

func router(avatarH *handler.AvatarHandler, enableLocalStatic bool) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(zlog.Middleware)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/u/{user_id}", avatarH.GetAvatar)

	if enableLocalStatic {
		fs := http.FileServer(http.Dir("./bucket"))
		r.NotFound(fs.ServeHTTP)
	}

	return r
}
