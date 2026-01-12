// internal/api/remote/remote.go
package remote

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/zetaoss/zavatar/internal/domain"
)

type API struct {
	endpoint  string
	secretKey string
	client    *http.Client
	cache     *cache.Cache
}

func (a *API) cacheGet(key string) (any, bool) {
	if a.cache == nil {
		return nil, false
	}
	return a.cache.Get(key)
}

func (a *API) cacheSet(key string, value any, expiration time.Duration) {
	if a.cache == nil {
		return
	}
	a.cache.Set(key, value, expiration)
}

func New(endpoint, secretKey string, cacheEnabled bool, cacheDefaultExpiration, cacheCleanupInterval time.Duration) *API {
	var apiCache *cache.Cache
	if cacheEnabled {
		apiCache = cache.New(cacheDefaultExpiration, cacheCleanupInterval)
	}

	return &API{
		endpoint:  strings.TrimRight(endpoint, "/"),
		secretKey: secretKey,
		client:    &http.Client{Timeout: 5 * time.Second},
		cache:     apiCache,
	}
}

func (a *API) Get(ctx context.Context, userID int64) (*domain.UserProfile, error) {
	if a.endpoint == "" {
		return nil, errors.New("remote: missing endpoint")
	}

	cacheKey := fmt.Sprintf("%d", userID)
	if v, ok := a.cacheGet(cacheKey); ok {
		if v == nil {
			return nil, nil
		}
		if cached, ok := v.(*domain.UserProfile); ok && cached != nil {
			cp := *cached
			return &cp, nil
		}
	}

	url := fmt.Sprintf("%s/api/internal/profiles/%d", a.endpoint, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	ts := fmt.Sprintf("%d", time.Now().UTC().Unix())
	sig := a.sign(req.Method, req.URL.Path, req.URL.RawQuery, ts)

	req.Header.Set("X-Api-Timestamp", ts)
	req.Header.Set("X-Api-Signature", sig)

	resp, err := a.client.Do(req)
	if err != nil {
		slog.Warn("GET failed", "userID", userID, "err", err)
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		slog.Debug("remote profile not found", "userID", userID, "status", resp.StatusCode)
		a.cacheSet(cacheKey, nil, cache.DefaultExpiration)
		return nil, nil
	default:
		slog.Warn("remote unexpected status", "userID", userID, "status", resp.StatusCode)
		return nil, fmt.Errorf("remote: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("remote read body failed", "userID", userID, "status", resp.StatusCode, "err", err)
		return nil, err
	}

	var payload struct {
		Name  string `json:"name"`
		T     int    `json:"t"`
		GHash string `json:"ghash"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		slog.Warn("remote unmarshal failed", "userID", userID, "status", resp.StatusCode, "body", string(body), "err", err)
		return nil, err
	}

	prof := domain.UserProfile{
		Name:  payload.Name,
		Type:  domain.AvatarTypeFromCode(domain.AvatarTypeCode(payload.T)),
		GHash: payload.GHash,
	}

	slog.Debug("remote profile received", "userID", userID, "status", resp.StatusCode, "profile", payload)

	a.cacheSet(cacheKey, &prof, cache.DefaultExpiration)

	return &prof, nil
}

func (a *API) sign(method, path, query, timestamp string) string {
	msg := strings.Join([]string{method, path, query, timestamp}, "\n")
	h := hmac.New(sha256.New, []byte(a.secretKey))
	_, _ = h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}
