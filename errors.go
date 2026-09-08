package higgsfield

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors, usable with errors.Is.
var (
	// ErrCredentialsMissing is returned by New when no credentials are supplied
	// via options or the environment.
	ErrCredentialsMissing = errors.New("higgsfield: API credentials not found; set HF_API_KEY_ID and HF_API_KEY_SECRET, or HF_CREDENTIALS=\"KEY_ID:KEY_SECRET\", or pass WithCredentials/WithCredentialPair")

	// ErrPollTimeout is returned when polling exceeds the configured maximum
	// poll time (see WithMaxPollTime).
	ErrPollTimeout = errors.New("higgsfield: polling exceeded maximum time")

	// The following match APIError values by HTTP semantics via errors.Is.
	ErrAuthentication   = errors.New("higgsfield: authentication failed")
	ErrNotEnoughCredits = errors.New("higgsfield: not enough credits")
	ErrValidation       = errors.New("higgsfield: validation error")
	ErrBadInput         = errors.New("higgsfield: bad input")
)

// ErrorKind classifies an APIError by the HTTP status the server returned.
type ErrorKind int

const (
	KindAPI              ErrorKind = iota // any other non-2xx status
	KindAuthentication                    // 401
	KindNotEnoughCredits                  // 403
	KindValidation                        // 422
	KindBadInput                          // 400
)

// ValidationDetail is a single field error from a 400/422 response.
type ValidationDetail struct {
	Type  string         `json:"type"`
	Loc   []any          `json:"loc"`
	Msg   string         `json:"msg"`
	Input any            `json:"input,omitempty"`
	Ctx   map[string]any `json:"ctx,omitempty"`
}

func (d ValidationDetail) String() string {
	parts := make([]string, 0, len(d.Loc))
	for _, l := range d.Loc {
		parts = append(parts, fmt.Sprint(l))
	}
	if len(parts) == 0 {
		return d.Msg
	}
	return strings.Join(parts, ".") + ": " + d.Msg
}

// APIError is returned for any non-2xx HTTP response from the API.
type APIError struct {
	Kind       ErrorKind
	StatusCode int
	Message    string
	Details    []ValidationDetail
	Body       []byte
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("higgsfield: %s (status %d)", e.Message, e.StatusCode)
	}
	return fmt.Sprintf("higgsfield: request failed with status %d", e.StatusCode)
}

// Is lets callers match APIError values against the exported sentinels, e.g.
// errors.Is(err, higgsfield.ErrAuthentication).
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrAuthentication:
		return e.Kind == KindAuthentication
	case ErrNotEnoughCredits:
		return e.Kind == KindNotEnoughCredits
	case ErrValidation:
		return e.Kind == KindValidation
	case ErrBadInput:
		return e.Kind == KindBadInput
	default:
		return false
	}
}

// newAPIError maps an HTTP status code and response body to a typed APIError.
func newAPIError(statusCode int, body []byte) *APIError {
	e := &APIError{StatusCode: statusCode, Body: body}

	switch statusCode {
	case 401:
		e.Kind = KindAuthentication
		e.Message = "invalid API credentials"
	case 403:
		e.Kind = KindNotEnoughCredits
		e.Message = "not enough credits"
	case 422:
		e.Kind = KindValidation
	case 400:
		e.Kind = KindBadInput
	default:
		e.Kind = KindAPI
	}

	msg, details := extractError(body)
	if len(details) > 0 {
		e.Details = details
	}
	if msg != "" {
		e.Message = msg
	}
	return e
}

// extractError pulls a human-readable message (and structured details, if any)
// out of an error body, mirroring the JS/Python clients: it inspects the
// "detail" field (string or array), then "details"/"message"/"error", and
// finally falls back to the raw body text.
func extractError(body []byte) (string, []ValidationDetail) {
	if len(body) == 0 {
		return "", nil
	}

	var payload struct {
		Detail  json.RawMessage `json:"detail"`
		Details json.RawMessage `json:"details"`
		Message string          `json:"message"`
		Error   string          `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return strings.TrimSpace(string(body)), nil
	}

	for _, raw := range []json.RawMessage{payload.Detail, payload.Details} {
		if len(raw) == 0 {
			continue
		}
		var s string
		if json.Unmarshal(raw, &s) == nil && s != "" {
			return s, nil
		}
		var ds []ValidationDetail
		if json.Unmarshal(raw, &ds) == nil && len(ds) > 0 {
			msgs := make([]string, 0, len(ds))
			for _, d := range ds {
				msgs = append(msgs, d.String())
			}
			return strings.Join(msgs, ", "), ds
		}
	}

	if payload.Message != "" {
		return payload.Message, nil
	}
	if payload.Error != "" {
		return payload.Error, nil
	}
	return strings.TrimSpace(string(body)), nil
}
