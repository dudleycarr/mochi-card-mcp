// Package mochi is a thin client for the Mochi Cards REST API
// (https://mochi.cards/docs/api/). It covers the card and deck operations
// needed by the MCP server.
package mochi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// DefaultBaseURL is the base URL of the Mochi Cards REST API.
const DefaultBaseURL = "https://app.mochi.cards/api"

// Client talks to the Mochi Cards REST API. It is safe for concurrent use.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Option customizes a Client.
type Option func(*Client)

// WithBaseURL overrides the API base URL. It is primarily useful in tests.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithHTTPClient overrides the underlying *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// NewClient creates a Client authenticating with the given API key. The key is
// sent using HTTP Basic auth as the username with an empty password, as Mochi
// requires.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		baseURL:    DefaultBaseURL,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// APIError represents a non-2xx response from the Mochi API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("mochi: unexpected status %d", e.StatusCode)
	}
	return fmt.Sprintf("mochi: unexpected status %d: %s", e.StatusCode, e.Body)
}

// do performs an HTTP request against the API. If body is non-nil it is encoded
// as JSON. If out is non-nil a successful response body is decoded into it.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("mochi: encoding request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return fmt.Errorf("mochi: building request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.doRequest(req, out)
}

// doRequest applies authentication, executes req, checks the status, and decodes
// a successful JSON response body into out (when out is non-nil and a body is
// present). Callers that need a non-JSON request body (e.g. multipart uploads)
// build the request themselves and call this directly.
func (c *Client) doRequest(req *http.Request, out any) error {
	req.SetBasicAuth(c.apiKey, "")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mochi: request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("mochi: reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(data))}
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("mochi: decoding response: %w", err)
		}
	}
	return nil
}
