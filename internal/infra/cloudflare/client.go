// infra/cloudflare/client.go
package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zetaoss/zavatar/internal/zlog"
)

type Purger interface {
	PurgePrefixes(ctx context.Context, prefixes []string) error
}

type Client struct {
	zoneID   string
	apiToken string
	client   *http.Client
}

func NewClient(zoneID, apiToken string) *Client {
	if zoneID == "" || apiToken == "" {
		return nil
	}
	return &Client{
		zoneID:   zoneID,
		apiToken: apiToken,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

type purgeReq struct {
	Prefixes []string `json:"prefixes"`
}

type purgeResp struct {
	Success bool `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *Client) PurgePrefixes(ctx context.Context, prefixes []string) error {
	if c == nil || len(prefixes) == 0 {
		return nil
	}

	payload, _ := json.Marshal(purgeReq{Prefixes: prefixes})
	apiURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/purge_cache", c.zoneID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	zlog.Ctx(ctx).Debug("requesting cloudflare purge by prefix", "count", len(prefixes))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api status: %d", resp.StatusCode)
	}

	var r purgeResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("decode error: %w", err)
	}

	if !r.Success {
		msg := "unknown error"
		if len(r.Errors) > 0 {
			msg = r.Errors[0].Message
		}
		return fmt.Errorf("cloudflare failed: %s", msg)
	}

	return nil
}
