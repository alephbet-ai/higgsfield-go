package higgsfield

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// Input holds the parameters for a generation request. Its fields depend on
// the model; see the Higgsfield API docs for each endpoint. It is sent as the
// JSON request body.
type Input map[string]any

type subscribeConfig struct {
	webhookURL string
	poll       bool
	onUpdate   func(*Response)
}

// SubscribeOption configures a single Subscribe or Submit call.
type SubscribeOption func(*subscribeConfig)

// WithWebhook registers a webhook to be called on completion. The URL is added
// as the hf_webhook query parameter. The secret is accepted for forward
// compatibility but is not currently sent by the v2 API path.
func WithWebhook(webhookURL, secret string) SubscribeOption {
	return func(sc *subscribeConfig) { sc.webhookURL = webhookURL }
}

// WithoutPolling makes Subscribe return immediately after submission, without
// waiting for the request to reach a terminal state.
func WithoutPolling() SubscribeOption {
	return func(sc *subscribeConfig) { sc.poll = false }
}

// WithOnUpdate registers a callback invoked with each status update observed
// while polling (and once for the initial submission response).
func WithOnUpdate(fn func(*Response)) SubscribeOption {
	return func(sc *subscribeConfig) { sc.onUpdate = fn }
}

func (c *Client) newSubscribeConfig(opts ...SubscribeOption) subscribeConfig {
	sc := subscribeConfig{poll: true}
	for _, opt := range opts {
		opt(&sc)
	}
	return sc
}

// Submit sends a generation request and returns the initial response without
// waiting for completion. Use the returned RequestID with Status or Cancel.
func (c *Client) Submit(ctx context.Context, model Model, input any, opts ...SubscribeOption) (*Response, error) {
	sc := c.newSubscribeConfig(opts...)
	return c.submit(ctx, model, input, sc)
}

// Subscribe sends a generation request and, unless WithoutPolling is set, polls
// until the request reaches a terminal state, returning the final response.
func (c *Client) Subscribe(ctx context.Context, model Model, input any, opts ...SubscribeOption) (*Response, error) {
	sc := c.newSubscribeConfig(opts...)

	res, err := c.submit(ctx, model, input, sc)
	if err != nil {
		return nil, err
	}
	if sc.onUpdate != nil {
		sc.onUpdate(res)
	}
	if !sc.poll || res.IsTerminal() || res.RequestID == "" {
		return res, nil
	}
	return c.poll(ctx, res.RequestID, sc)
}

// Status returns the current state of a request by id.
func (c *Client) Status(ctx context.Context, requestID string) (*Response, error) {
	data, err := c.do(ctx, http.MethodGet, "/requests/"+url.PathEscape(requestID)+"/status", nil)
	if err != nil {
		return nil, err
	}
	return decodeResponse(data)
}

// Cancel cancels a queued request. Requests already in progress cannot be
// canceled and return an error.
func (c *Client) Cancel(ctx context.Context, requestID string) error {
	_, err := c.do(ctx, http.MethodPost, "/requests/"+url.PathEscape(requestID)+"/cancel", nil)
	return err
}

func (c *Client) submit(ctx context.Context, model Model, input any, sc subscribeConfig) (*Response, error) {
	path := normalizeEndpoint(model)
	if sc.webhookURL != "" {
		sep := "?"
		if strings.Contains(path, "?") {
			sep = "&"
		}
		path += sep + "hf_webhook=" + url.QueryEscape(sc.webhookURL)
	}

	data, err := c.do(ctx, http.MethodPost, path, input)
	if err != nil {
		return nil, err
	}
	return decodeResponse(data)
}

func (c *Client) poll(ctx context.Context, requestID string, sc subscribeConfig) (*Response, error) {
	start := c.clock.Now()
	for {
		if c.clock.Now().Sub(start) > c.maxPollTime {
			return nil, ErrPollTimeout
		}

		res, err := c.Status(ctx, requestID)
		if err != nil {
			// Transient server errors during polling are ignored; keep going.
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode < 500 {
				return nil, err
			}
		} else {
			if sc.onUpdate != nil {
				sc.onUpdate(res)
			}
			if res.IsTerminal() {
				return res, nil
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-c.clock.After(c.pollInterval):
		}
	}
}

func normalizeEndpoint(model Model) string {
	s := string(model)
	if s == "" || strings.HasPrefix(s, "/") {
		return s
	}
	return "/" + s
}
