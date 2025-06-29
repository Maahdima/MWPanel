package httphelper

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	BaseURL            string        // The base URL for all requests, e.g., "http://192.168.88.1/rest"
	Username           string        // Username for Basic Authentication.
	Password           string        // Password for Basic Authentication.
	InsecureSkipVerify bool          // If true, the client will skip TLS certificate verification. Equivalent to 'curl -k'.
	Timeout            time.Duration // Request timeout. Defaults to 30 seconds if not set.
}

type Client struct {
	httpClient *http.Client
	config     Config
	logger     *zap.Logger
}

// NewClient creates and configures a new helper Client.
func NewClient(config Config) (*Client, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("BaseURL is a required configuration field")
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()

	if config.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: config.InsecureSkipVerify}
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &Client{
		httpClient: httpClient,
		config:     config,
		logger:     zap.L().With(zap.String("baseURL", config.BaseURL)),
	}, nil
}

func (c *Client) Do(req *http.Request, respBody interface{}) error {
	if c.config.Username != "" || c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	logger := c.logger.With(zap.String("method", req.Method), zap.String("url", req.URL.String()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Request failed", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Reading response body failed", zap.Int("statusCode", resp.StatusCode), zap.Error(err))
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		logger.Warn("Request returned non-2xx status",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("responseBody", string(body)),
		)
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	if respBody == nil || len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, respBody); err != nil {
		logger.Error("Unmarshaling response body failed", zap.String("body", string(body)), zap.Error(err))
		return err
	}

	return nil
}

func (c *Client) newRequestWithBody(ctx context.Context, method, path string, reqBody interface{}) (*http.Request, error) {
	fullURL := c.config.BaseURL + path

	var bodyReader io.Reader
	if reqBody != nil {
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			c.logger.Error("Failed to marshal request body", zap.String("method", method), zap.String("path", path), zap.Error(err))
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		c.logger.Error("Failed to create request", zap.String("method", method), zap.String("path", path), zap.Error(err))
		return nil, err
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) Get(ctx context.Context, path string, respBody interface{}) error {
	fullURL := c.config.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		c.logger.Error("Failed to create GET request", zap.String("path", path), zap.Error(err))
		return err
	}

	return c.Do(req, respBody)
}

func (c *Client) Post(ctx context.Context, path string, reqBody, respBody interface{}) error {
	req, err := c.newRequestWithBody(ctx, http.MethodPost, path, reqBody)
	if err != nil {
		return err
	}
	return c.Do(req, respBody)
}

func (c *Client) Put(ctx context.Context, path string, reqBody, respBody interface{}) error {
	req, err := c.newRequestWithBody(ctx, http.MethodPut, path, reqBody)
	if err != nil {
		return err
	}
	return c.Do(req, respBody)
}

func (c *Client) Patch(ctx context.Context, path string, reqBody, respBody interface{}) error {
	req, err := c.newRequestWithBody(ctx, http.MethodPatch, path, reqBody)
	if err != nil {
		return err
	}
	return c.Do(req, respBody)
}

func (c *Client) Delete(ctx context.Context, path string, respBody interface{}) error {
	fullURL := c.config.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fullURL, nil)
	if err != nil {
		c.logger.Error("Failed to create DELETE request", zap.String("path", path), zap.Error(err))
		return err
	}

	return c.Do(req, respBody)
}
