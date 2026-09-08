package higgsfield

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// do performs an authenticated JSON request against the API base URL,
// retrying network errors and 5xx responses per the client's retry policy. It
// returns the raw response body for 2xx responses, or a typed *APIError.
//
// path may include a query string. body, when non-nil, is JSON-encoded.
func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	var reqBody []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("higgsfield: encode request body: %w", err)
		}
		reqBody = b
	}

	url := c.baseURL + path
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-c.clock.After(c.backoffDelay(attempt - 1)):
			}
		}

		var reader io.Reader
		if reqBody != nil {
			reader = bytes.NewReader(reqBody)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, reader)
		if err != nil {
			return nil, fmt.Errorf("higgsfield: build request: %w", err)
		}
		c.setHeaders(req, reqBody != nil)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("higgsfield: %s %s: %w", method, path, err)
			if isRetryableNetErr(err) && attempt < c.maxRetries {
				continue
			}
			return nil, lastErr
		}

		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("higgsfield: read response body: %w", readErr)
			if attempt < c.maxRetries {
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return data, nil
		}

		apiErr := newAPIError(resp.StatusCode, data)
		if resp.StatusCode >= 500 && attempt < c.maxRetries {
			lastErr = apiErr
			continue
		}
		return nil, apiErr
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("higgsfield: request failed")
}

func (c *Client) setHeaders(req *http.Request, hasBody bool) {
	req.Header.Set("Authorization", "Key "+c.keyID+":"+c.keySecret)
	req.Header.Set("User-Agent", c.userAgent)
	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
}
