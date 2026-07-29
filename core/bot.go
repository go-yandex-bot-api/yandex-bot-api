// Package core provides the core logic and bot instance.
package core

import (
	"net/http"
)

const (
	// BaseURL is the default base URL for all requests to the Yandex Messenger API.
	BaseURL = "https://botapi.messenger.yandex.net/bot/v1/"
)

// Client represents the core HTTP client for working with the Yandex Messenger API.
type Client struct {
	token      string
	httpClient HTTPClient
	debug      bool // Set to true to print raw HTTP requests and responses
	apiURL     string
	maxRetries int // Number of automatic retries for network errors and 429 status
	errConfig  *ErrorHandlingConfig
}

// ErrorHandlingConfig defines rules for retrying failed requests.
type ErrorHandlingConfig struct {
	// RetryableStatuses contains a list of HTTP status codes that should trigger a retry.
	RetryableStatuses []int
}

// Option configures the Client.
type Option func(*Client)

// WithClient allows the library user to pass a custom HTTP client.
func WithClient(c HTTPClient) Option {
	return func(cl *Client) { cl.httpClient = c }
}

// WithAPIURL allows overriding the base API URL.
func WithAPIURL(apiURL string) Option {
	return func(cl *Client) {
		if apiURL != "" && apiURL[len(apiURL)-1] != '/' {
			apiURL += "/"
		}
		cl.apiURL = apiURL
	}
}

// WithMaxRetries allows customizing the number of automatic retries.
func WithMaxRetries(retries int) Option {
	return func(cl *Client) { cl.maxRetries = retries }
}

// WithDebug enables or disables debug logging for raw HTTP requests.
func WithDebug(debug bool) Option {
	return func(cl *Client) { cl.debug = debug }
}

// WithErrorHandlingConfig allows configuring which HTTP status codes should trigger a retry.
// By default, if no custom configuration is provided, the client retries on 429 and >= 500 status codes.
func WithErrorHandlingConfig(config *ErrorHandlingConfig) Option {
	return func(cl *Client) { cl.errConfig = config }
}

// NewClient creates a new Client instance with default settings and applies options.
//
// It requires a token obtained from Yandex.
func NewClient(token string, opts ...Option) *Client {
	cl := &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 0}, // Rely on Context for timeouts
		apiURL:     BaseURL,
		maxRetries: 3, //nolint:mnd // Default safe number of retries
	}
	for _, opt := range opts {
		opt(cl)
	}

	if cl.debug {
		cl.httpClient = &loggingClient{next: cl.httpClient}
	}

	return cl
}

// Debug returns whether debug logging is enabled.
func (c *Client) Debug() bool {
	return c.debug
}

// isRetryableStatus checks if the given HTTP status code should trigger a retry based on the client configuration.
func (c *Client) isRetryableStatus(code int) bool {
	if c.errConfig != nil {
		for _, status := range c.errConfig.RetryableStatuses {
			if code == status {
				return true
			}
		}
		return false
	}
	// Default behavior
	return code == http.StatusTooManyRequests || code >= http.StatusInternalServerError
}
