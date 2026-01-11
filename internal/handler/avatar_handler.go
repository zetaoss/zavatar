// internal/handler/avatar_handler.go
package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/zetaoss/zavatar/internal/domain"
	"github.com/zetaoss/zavatar/internal/service"
	"github.com/zetaoss/zavatar/internal/storage"
	"github.com/zetaoss/zavatar/internal/zlog"
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

	q := r.URL.Query()
	sizeEff := domain.NormalizeSizeQuery(q.Get("s"))

	tStr := q.Get("t")
	t := 0
	if tStr != "" {
		t, err = strconv.Atoi(tStr)
		if err != nil || t < 1 || t > 3 {
			http.Error(w, "bad params", http.StatusBadRequest)
			return
		}
	}

	log := zlog.Ctx(r.Context()).With("uid", uid, "s", sizeEff, "t", t)

	out, err := h.svc.Resolve(r.Context(), service.ResolveInput{
		UserID: uid,
		Size:   sizeEff,
		T:      t,
	})
	if err != nil {
		if storage.IsNotFound(err) {
			http.NotFound(w, r)
			return
		}

		if errors.Is(err, service.ErrBadParams) {
			http.Error(w, "bad params", http.StatusBadRequest)
			return
		}

		if errors.Is(err, context.Canceled) {
			log.Debug("canceled")
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			log.Warn("timeout")
			http.Error(w, "timeout", http.StatusGatewayTimeout)
			return
		}

		log.Error("resolve failed", "err", err)
		http.Error(w, "resolve failed", http.StatusInternalServerError)
		return
	}

	if out.CacheControl != "" {
		w.Header().Set("Cache-Control", out.CacheControl)
	}

	if out.RedirectURL == "" {
		http.Error(w, "missing redirect", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, out.RedirectURL, http.StatusFound)
}
