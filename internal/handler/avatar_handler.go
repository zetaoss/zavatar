// internal/handler/avatar_handler.go
package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/zetaoss/zavatar/internal/domain"
	"github.com/zetaoss/zavatar/internal/service"
	"github.com/zetaoss/zavatar/internal/store/storage"
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
		if err != nil {
			http.Error(w, "bad t", http.StatusBadRequest)
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

	sum := sha256.Sum256(out.PNG)
	etag := `"` + hex.EncodeToString(sum[:]) + `"`

	w.Header().Set("Content-Type", "image/png")
	if out.CacheControl != "" {
		w.Header().Set("Cache-Control", out.CacheControl)
	}
	w.Header().Set("ETag", etag)

	if matchIfNoneMatch(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out.PNG)
}

func matchIfNoneMatch(ifNoneMatch string, etag string) bool {
	ifNoneMatch = strings.TrimSpace(ifNoneMatch)
	if ifNoneMatch == "" {
		return false
	}
	if ifNoneMatch == "*" {
		return true
	}

	parts := strings.Split(ifNoneMatch, ",")
	for _, p := range parts {
		tok := strings.TrimSpace(p)
		if tok == etag || (strings.HasPrefix(tok, "W/") && strings.TrimSpace(tok[2:]) == etag) {
			return true
		}
	}
	return false
}
