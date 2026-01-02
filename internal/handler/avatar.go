// internal/handler/avatar.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/zetaoss/zavatar/internal/domain"
	"github.com/zetaoss/zavatar/internal/service"
)

type AvatarHandler struct {
	svc *service.AvatarService
}

func NewAvatarHandler(svc *service.AvatarService) *AvatarHandler {
	return &AvatarHandler{svc: svc}
}

func (h *AvatarHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	uidStr := chi.URLParam(r, "user_id")
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil || uid <= 0 {
		http.Error(w, "bad user_id", http.StatusBadRequest)
		return
	}

	size := domain.NormalizeSize(r.URL.Query().Get("s"))

	out, err := h.svc.Resolve(r.Context(), service.ResolveInput{
		UserID: uid,
		Size:   size,
	})
	if err != nil {
		http.Error(w, "resolve failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, out.RedirectURL, http.StatusFound)
}
