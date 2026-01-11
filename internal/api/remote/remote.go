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
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/zetaoss/zavatar/internal/domain"
)

type API struct {
	endpoint  string
	accessKey string
	secretKey string
	client    *http.Client
}

func New(endpoint, accessKey, secretKey string) *API {
	return &API{
		endpoint:  strings.TrimRight(endpoint, "/"),
		accessKey: accessKey,
		secretKey: secretKey,
		client:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (a *API) Get(ctx context.Context, userID int64) (*domain.UserProfile, error) {
	if a.endpoint == "" {
		return nil, errors.New("remote: missing endpoint")
	}

	url := fmt.Sprintf("%s/api/internal/profiles/%d", a.endpoint, userID)
	slog.Debug("remote api get", "userID", userID, "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	ts := fmt.Sprintf("%d", time.Now().UTC().Unix())
	sig := a.sign(req.Method, req.URL.Path, req.URL.RawQuery, ts)

	req.Header.Set("X-Api-Access-Key", a.accessKey)
	req.Header.Set("X-Api-Timestamp", ts)
	req.Header.Set("X-Api-Signature", sig)

	resp, err := a.client.Do(req)
	if err != nil {
		slog.Warn("remote api request failed", "err", err, "user_id", userID)
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		slog.Debug("remote api response", "status", resp.StatusCode, "user_id", userID)
	case http.StatusNotFound:
		slog.Debug("remote api not found", "status", resp.StatusCode, "user_id", userID)
		return nil, nil
	default:
		slog.Warn("remote api unexpected status", "status", resp.StatusCode, "user_id", userID)
		return nil, fmt.Errorf("remote: unexpected status %d", resp.StatusCode)
	}

	var payload struct {
		Name  string `json:"name"`
		Type  int    `json:"type"`
		GHash string `json:"ghash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return &domain.UserProfile{
		Name:  payload.Name,
		Type:  domain.AvatarTypeFromCode(domain.AvatarTypeCode(payload.Type)),
		GHash: payload.GHash,
	}, nil
}

func (a *API) sign(method, path, query, timestamp string) string {
	msg := strings.Join([]string{method, path, query, timestamp}, "\n")
	h := hmac.New(sha256.New, []byte(a.secretKey))
	_, _ = h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}
