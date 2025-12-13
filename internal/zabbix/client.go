// Package zabbix provides a client for the Zabbix JSON-RPC API.
package zabbix

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Client is a Zabbix API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	tokenMu    sync.RWMutex
	token      string // API token or session token
	requestID  int64
}

// Request represents a JSON-RPC request to the Zabbix API.
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int64       `json:"id"`
}

// Response represents a JSON-RPC response from the Zabbix API.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *APIError       `json:"error,omitempty"`
	ID      int64           `json:"id"`
}

// APIError represents an error returned by the Zabbix API.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (e *APIError) Error() string {
	if e.Data != "" {
		return fmt.Sprintf("zabbix API error %d: %s - %s", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("zabbix API error %d: %s", e.Code, e.Message)
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithInsecureSkipVerify disables TLS certificate verification.
func WithInsecureSkipVerify() ClientOption {
	return func(c *Client) {
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			if transport.TLSClientConfig == nil {
				transport.TLSClientConfig = &tls.Config{}
			}
			transport.TLSClientConfig.InsecureSkipVerify = true
		}
	}
}

// NewClient creates a new Zabbix API client.
func NewClient(baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: baseURL + "/api_jsonrpc.php",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{},
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// SetToken sets the API token for authentication.
func (c *Client) SetToken(token string) {
	c.tokenMu.Lock()
	c.token = token
	c.tokenMu.Unlock()
}

// getToken returns the current token (thread-safe).
func (c *Client) getToken() string {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.token
}

// nextID returns the next request ID.
func (c *Client) nextID() int64 {
	return atomic.AddInt64(&c.requestID, 1)
}

// call makes a JSON-RPC call to the Zabbix API.
func (c *Client) call(ctx context.Context, method string, params interface{}, result interface{}) error {
	return c.callWithAuth(ctx, method, params, result, true)
}

// callNoAuth makes a JSON-RPC call without authentication (for apiinfo.version, etc.)
func (c *Client) callNoAuth(ctx context.Context, method string, params interface{}, result interface{}) error {
	return c.callWithAuth(ctx, method, params, result, false)
}

// callWithAuth makes a JSON-RPC call to the Zabbix API with optional authentication.
func (c *Client) callWithAuth(ctx context.Context, method string, params interface{}, result interface{}, useAuth bool) error {
	req := Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      c.nextID(),
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json-rpc")
	if useAuth {
		if token := c.getToken(); token != "" {
			httpReq.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResp Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResp.Error != nil {
		return apiResp.Error
	}

	if result != nil {
		if err := json.Unmarshal(apiResp.Result, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// Login authenticates with username and password and stores the session token.
func (c *Client) Login(ctx context.Context, username, password string) error {
	params := map[string]string{
		"username": username,
		"password": password,
	}

	var token string
	if err := c.call(ctx, "user.login", params, &token); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	c.SetToken(token)
	return nil
}

// Logout invalidates the current session token.
func (c *Client) Logout(ctx context.Context) error {
	if c.getToken() == "" {
		return nil
	}

	var result bool
	if err := c.call(ctx, "user.logout", []string{}, &result); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	c.SetToken("")
	return nil
}

// Version returns the Zabbix API version.
// Note: apiinfo.version must be called without authorization header.
func (c *Client) Version(ctx context.Context) (string, error) {
	var version string
	if err := c.callNoAuth(ctx, "apiinfo.version", []string{}, &version); err != nil {
		return "", err
	}
	return version, nil
}

// IsConnected checks if the client can communicate with Zabbix.
func (c *Client) IsConnected(ctx context.Context) bool {
	_, err := c.Version(ctx)
	return err == nil
}
