package higgsfield

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Default configuration values.
const (
	DefaultBaseURL         = "https://api.higgsfield.ai"
	DefaultTimeout         = 120 * time.Second
	DefaultMaxRetries      = 3
	DefaultRetryBackoff    = 1 * time.Second
	DefaultRetryMaxBackoff = 60 * time.Second
	DefaultPollInterval    = 2 * time.Second
	DefaultMaxPollTime     = 5 * time.Minute
)

// Version is the client version, reported in the User-Agent header.
const Version = "0.1.0"

const defaultUserAgent = "higgsfield-go/" + Version

// Client is a Higgsfield API client. It is safe for concurrent use by multiple
// goroutines. Create one with New.
type Client struct {
	baseURL         string
	httpClient      *http.Client
	keyID           string
	keySecret       string
	userAgent       string
	headers         map[string]string
	maxRetries      int
	retryBackoff    time.Duration
	retryMaxBackoff time.Duration
	pollInterval    time.Duration
	maxPollTime     time.Duration

	// configuration captured from options, resolved in New
	credentialsRaw string
	timeout        time.Duration
	clock          clock
}

// New creates a Client. Credentials are taken from options if provided,
// otherwise resolved from the environment (see the package docs). It returns
// ErrCredentialsMissing when none are found, or an error for invalid options.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		baseURL:         DefaultBaseURL,
		userAgent:       defaultUserAgent,
		headers:         map[string]string{},
		maxRetries:      DefaultMaxRetries,
		retryBackoff:    DefaultRetryBackoff,
		retryMaxBackoff: DefaultRetryMaxBackoff,
		pollInterval:    DefaultPollInterval,
		maxPollTime:     DefaultMaxPollTime,
		timeout:         DefaultTimeout,
		clock:           realClock{},
	}

	for _, opt := range opts {
		opt(c)
	}

	if err := c.resolveCredentials(); err != nil {
		return nil, err
	}
	if err := c.validate(); err != nil {
		return nil, err
	}

	c.baseURL = strings.TrimRight(c.baseURL, "/")
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: c.timeout}
	}

	return c, nil
}

func (c *Client) resolveCredentials() error {
	if c.credentialsRaw != "" {
		id, secret, ok := splitCredential(c.credentialsRaw)
		if !ok {
			return fmt.Errorf("higgsfield: WithCredentials expects \"KEY_ID:KEY_SECRET\"")
		}
		c.keyID, c.keySecret = id, secret
	}
	if c.keyID != "" && c.keySecret != "" {
		return nil
	}

	// Environment fallback chain.
	for _, name := range []string{"HF_CREDENTIALS", "HF_KEY"} {
		if v := os.Getenv(name); v != "" {
			if id, secret, ok := splitCredential(v); ok {
				c.keyID, c.keySecret = id, secret
				return nil
			}
		}
	}
	if id, secret := os.Getenv("HF_API_KEY_ID"), os.Getenv("HF_API_KEY_SECRET"); id != "" && secret != "" {
		c.keyID, c.keySecret = id, secret
		return nil
	}
	if id, secret := os.Getenv("HF_API_KEY"), os.Getenv("HF_API_SECRET"); id != "" && secret != "" {
		c.keyID, c.keySecret = id, secret
		return nil
	}
	return ErrCredentialsMissing
}

func (c *Client) validate() error {
	if c.baseURL == "" {
		return fmt.Errorf("higgsfield: base URL must not be empty")
	}
	if c.timeout <= 0 {
		return fmt.Errorf("higgsfield: timeout must be positive")
	}
	if c.maxRetries < 0 {
		return fmt.Errorf("higgsfield: maxRetries must be non-negative")
	}
	if c.pollInterval <= 0 {
		return fmt.Errorf("higgsfield: pollInterval must be positive")
	}
	if c.maxPollTime <= 0 {
		return fmt.Errorf("higgsfield: maxPollTime must be positive")
	}
	return nil
}

// splitCredential splits a "KEY_ID:KEY_SECRET" string. The secret may itself
// contain colons; only the first colon is significant.
func splitCredential(v string) (id, secret string, ok bool) {
	i := strings.IndexByte(v, ':')
	if i <= 0 || i == len(v)-1 {
		return "", "", false
	}
	return v[:i], v[i+1:], true
}
