// internal/app/router.go
package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zetaoss/zavatar/internal/handler"
)

func router(avatarH *handler.AvatarHandler) http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/u/{user_id}", avatarH.GetAvatar)

	return r
}
