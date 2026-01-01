// internal/handler/object_serve.go
package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/zetaoss/zavatar/internal/storage/object"
)

type R2ServeHandler struct {
	obj object.Store
}

func NewR2ServeHandler(obj object.Store) *R2ServeHandler {
	return &R2ServeHandler{obj: obj}
}

func (h *R2ServeHandler) Serve(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/r2/")
	if key == "" {
		http.NotFound(w, r)
		return
	}

	rc, ct, err := h.obj.Get(r.Context(), key)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer rc.Close()

	if ct == "" {
		ct = "application/octet-stream"
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Content-Type", ct)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}
