package httphelper

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	defaultTimeout      = 30 * time.Second
	defaultMaxRetries   = 3
	defaultRetryBackoff = 500 * time.Millisecond
)

type Config struct {
	BaseURL            string        // The base URL for all requests, e.g., "http://127.0.0.1/rest"
	Username           string        // Username for Basic Authentication.
	Password           string        // Password for Basic Authentication.
	InsecureSkipVerify bool          // If true, the client will skip TLS certificate verification. Equivalent to 'curl -k'.
	Timeout            time.Duration // Request timeout. Defaults to 30 seconds if not set.
	MaxRetries         int           // Number of retries after the initial attempt. Defaults to 3.
	RetryBackoff       time.Duration // Base delay between retries. Defaults to 500ms.
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
		config.Timeout = defaultTimeout
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = defaultMaxRetries
	}
	if config.RetryBackoff == 0 {
		config.RetryBackoff = defaultRetryBackoff
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
	maxAttempts := c.config.MaxRetries + 1
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			if !isRetryableError(lastErr) {
				return lastErr
			}
			if err := resetRequestBody(req); err != nil {
				return lastErr
			}
			if err := waitBeforeRetry(req.Context(), c.config.RetryBackoff, attempt-1); err != nil {
				return err
			}
			c.logger.Warn("Retrying request",
				zap.String("method", req.Method),
				zap.String("url", req.URL.String()),
				zap.Int("attempt", attempt),
				zap.Error(lastErr),
			)
		}

		lastErr = c.doOnce(req, respBody)
		if lastErr == nil {
			return nil
		}
	}

	return lastErr
}

func (c *Client) doOnce(req *http.Request, respBody interface{}) error {
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
		err = fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(body))
		if isRetryableStatusCode(resp.StatusCode) {
			return &retryableError{err: err}
		}
		return err
	}

	if respBody == nil || len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, respBody); err != nil {
		logger.Error("Unmarshalling response body failed", zap.String("body", string(body)), zap.Error(err))
		return err
	}

	return nil
}

type retryableError struct {
	err error
}

func (e *retryableError) Error() string {
	return e.err.Error()
}

func (e *retryableError) Unwrap() error {
	return e.err
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var retryable *retryableError
	if errors.As(err, &retryable) {
		return true
	}

	if errors.Is(err, context.Canceled) {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "broken pipe") ||
		strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "no route to host") ||
		strings.Contains(errMsg, "i/o timeout") ||
		strings.Contains(errMsg, "eof")
}

func isRetryableStatusCode(statusCode int) bool {
	return statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout
}

func resetRequestBody(req *http.Request) error {
	if req.Body == nil {
		return nil
	}
	if req.GetBody == nil {
		return fmt.Errorf("cannot retry request with non-resettable body")
	}

	body, err := req.GetBody()
	if err != nil {
		return err
	}

	req.Body = body
	return nil
}

func waitBeforeRetry(ctx context.Context, baseBackoff time.Duration, attempt int) error {
	delay := baseBackoff * time.Duration(attempt)
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *Client) newRequestWithBody(ctx context.Context, method, path string, reqBody interface{}) (*http.Request, error) {
	fullURL := c.config.BaseURL + path

	var bodyReader io.Reader
	var jsonBody []byte
	if reqBody != nil {
		var err error
		jsonBody, err = json.Marshal(reqBody)
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
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(jsonBody)), nil
		}
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
