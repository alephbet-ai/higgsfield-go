package higgsfield

import (
	"net/http"
	"time"
)

// Option configures a Client in New.
type Option func(*Client)

// WithCredentials sets credentials from a single "KEY_ID:KEY_SECRET" string.
func WithCredentials(credentials string) Option {
	return func(c *Client) { c.credentialsRaw = credentials }
}

// WithCredentialPair sets credentials from separate key id and secret values.
func WithCredentialPair(keyID, keySecret string) Option {
	return func(c *Client) {
		c.keyID = keyID
		c.keySecret = keySecret
	}
}

// WithBaseURL overrides the API base URL (default DefaultBaseURL).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = baseURL }
}

// WithTimeout sets the per-request HTTP timeout. Ignored if WithHTTPClient is
// used.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// WithHTTPClient supplies a custom *http.Client. When set, WithTimeout has no
// effect.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

// WithMaxRetries sets how many times a failed request is retried (default
// DefaultMaxRetries). Only network errors and 5xx responses are retried.
func WithMaxRetries(n int) Option {
	return func(c *Client) { c.maxRetries = n }
}

// WithRetryBackoff sets the base delay for exponential backoff between retries.
func WithRetryBackoff(d time.Duration) Option {
	return func(c *Client) { c.retryBackoff = d }
}

// WithRetryMaxBackoff caps the delay between retries.
func WithRetryMaxBackoff(d time.Duration) Option {
	return func(c *Client) { c.retryMaxBackoff = d }
}

// WithPollInterval sets how often Subscribe polls for status (default
// DefaultPollInterval).
func WithPollInterval(d time.Duration) Option {
	return func(c *Client) { c.pollInterval = d }
}

// WithMaxPollTime caps how long Subscribe polls before returning ErrPollTimeout.
func WithMaxPollTime(d time.Duration) Option {
	return func(c *Client) { c.maxPollTime = d }
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// WithHeader adds a custom header sent on every API request.
func WithHeader(key, value string) Option {
	return func(c *Client) { c.headers[key] = value }
}

// WithHeaders adds multiple custom headers sent on every API request.
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) {
		for k, v := range headers {
			c.headers[k] = v
		}
	}
}
