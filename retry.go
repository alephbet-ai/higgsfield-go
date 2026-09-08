package higgsfield

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// clock abstracts time so retry backoff and polling can be tested without
// real delays.
type clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// backoffDelay returns the delay before retry number attempt (0-based): an
// exponential backoff with up to one second of jitter, capped at
// retryMaxBackoff.
func (c *Client) backoffDelay(attempt int) time.Duration {
	d := c.retryBackoff * time.Duration(1<<uint(attempt))
	if d <= 0 || d > c.retryMaxBackoff {
		d = c.retryMaxBackoff
	}
	d += time.Duration(rand.Int63n(int64(time.Second)))
	if d > c.retryMaxBackoff {
		d = c.retryMaxBackoff
	}
	return d
}

// isRetryableNetErr reports whether a transport-level error is worth retrying.
// Context cancellation and deadline expiry are never retried.
func isRetryableNetErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	return true
}
