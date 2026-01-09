// internal/handler/purge_handler.go
package handler

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/zetaoss/zavatar/internal/service"
	"github.com/zetaoss/zavatar/internal/zlog"
)

type PurgeHandler struct {
	svc      *service.AvatarService
	purgeKey string
}

func NewPurgeHandler(svc *service.AvatarService, purgeKey string) *PurgeHandler {
	return &PurgeHandler{svc: svc, purgeKey: purgeKey}
}

type purgeResp struct {
	Deleted int `json:"deleted"`
}

func (h *PurgeHandler) PurgeUser(w http.ResponseWriter, r *http.Request) {
	if h.purgeKey == "" {
		http.Error(w, "purge disabled", http.StatusNotFound)
		return
	}

	got := r.Header.Get("X-Purge-Key")
	if subtle.ConstantTimeCompare([]byte(got), []byte(h.purgeKey)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	uidStr := chi.URLParam(r, "user_id")
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil || uid <= 0 {
		http.Error(w, "bad user_id", http.StatusBadRequest)
		return
	}

	log := zlog.Ctx(r.Context()).With("uid", uid)

	n, err := h.svc.Purge(r.Context(), uid)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Info("purge canceled")
			return
		}

		log.Error("purge failed", "err", err)
		http.Error(w, "purge failed", http.StatusInternalServerError)
		return
	}

	log.Info("purge ok", "deleted", n)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(purgeResp{Deleted: n})
}
