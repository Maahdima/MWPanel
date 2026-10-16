package shopclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/maahdima/mwp/api/dataservice/model"
)

type Link struct {
	UUID          string  `json:"uuid"`
	Linked        bool    `json:"linked"`
	Username      *string `json:"username,omitempty"`
	NotifyEnabled bool    `json:"notify_enabled"`
}

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		token:   strings.TrimSpace(token),
		http:    &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.baseURL != "" && c.token != ""
}

func (c *Client) Token() string {
	if c == nil {
		return ""
	}
	return c.token
}

func (c *Client) Forward(req *http.Request, body []byte) (*http.Response, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("shop service is not configured")
	}
	target := c.baseURL + req.URL.RequestURI()
	out, err := http.NewRequestWithContext(req.Context(), req.Method, target, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if contentType := req.Header.Get("Content-Type"); contentType != "" {
		out.Header.Set("Content-Type", contentType)
	}
	out.Header.Set("Authorization", "Bearer "+c.token)
	return c.http.Do(out)
}

func (c *Client) Links(ctx context.Context, uuids []string) (map[string]Link, error) {
	var rows []Link
	if err := c.do(ctx, http.MethodPost, "/api/service/links", map[string]any{"uuids": uuids}, &rows); err != nil {
		return nil, err
	}
	out := make(map[string]Link, len(rows))
	for _, row := range rows {
		out[row.UUID] = row
	}
	return out, nil
}

func (c *Client) Unlink(ctx context.Context, peerUUID string) error {
	return c.do(ctx, http.MethodDelete, "/api/service/peers/"+url.PathEscape(peerUUID)+"/links", nil, nil)
}

type Notifier struct {
	client *Client
}

func NewNotifier(client *Client) *Notifier {
	return &Notifier{client: client}
}

func (n *Notifier) Enabled() bool {
	return n != nil && n.client.Enabled()
}

func (n *Notifier) NotifyPeerUsage(ctx context.Context, peer *model.Peer, percent, totalUsage, limit int64) error {
	if peer == nil {
		return fmt.Errorf("peer is required")
	}
	return n.client.do(ctx, http.MethodPost, "/api/service/notify/usage", map[string]any{
		"uuid":        peer.UUID,
		"name":        peer.Name,
		"percent":     percent,
		"total_usage": totalUsage,
		"limit":       limit,
	}, nil)
}

func (n *Notifier) NotifyPeerExpiry(ctx context.Context, peer *model.Peer, daysLeft int) error {
	if peer == nil {
		return fmt.Errorf("peer is required")
	}
	return n.client.do(ctx, http.MethodPost, "/api/service/notify/expiry", map[string]any{
		"uuid":        peer.UUID,
		"name":        peer.Name,
		"days_left":   daysLeft,
		"expire_time": peer.ExpireTime,
	}, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body any, dest any) error {
	if !c.Enabled() {
		return fmt.Errorf("shop service is not configured")
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("shop %s %s: %s", method, path, strings.TrimSpace(string(payload)))
	}
	if dest == nil || len(payload) == 0 {
		return nil
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return err
	}
	if len(envelope.Data) == 0 {
		return nil
	}
	return json.Unmarshal(envelope.Data, dest)
}
