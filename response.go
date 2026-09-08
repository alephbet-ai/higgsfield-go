package higgsfield

import (
	"encoding/json"
	"fmt"
)

// Status is the lifecycle state of a generation request.
type Status string

const (
	StatusQueued     Status = "queued"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusNSFW       Status = "nsfw"
	StatusCanceled   Status = "canceled"
)

// Media is a generated asset (image or video) returned by the API.
type Media struct {
	URL string `json:"url"`
}

// Response is the result of a generation request, as returned by the submit
// and status endpoints.
type Response struct {
	Status    Status  `json:"status"`
	RequestID string  `json:"request_id"`
	StatusURL string  `json:"status_url"`
	CancelURL string  `json:"cancel_url"`
	Images    []Media `json:"images,omitempty"`
	Video     *Media  `json:"video,omitempty"`

	// Raw holds the complete, undecoded response body so callers can read
	// fields not modeled above (the API adds model-specific output over time).
	Raw json.RawMessage `json:"-"`
}

// IsQueued reports whether the request is queued and waiting to start.
func (r *Response) IsQueued() bool { return r.Status == StatusQueued }

// IsInProgress reports whether the request is currently being processed.
func (r *Response) IsInProgress() bool { return r.Status == StatusInProgress }

// IsCompleted reports whether the request finished successfully.
func (r *Response) IsCompleted() bool { return r.Status == StatusCompleted }

// IsFailed reports whether the request failed.
func (r *Response) IsFailed() bool { return r.Status == StatusFailed }

// IsNSFW reports whether the request was rejected by content moderation.
func (r *Response) IsNSFW() bool { return r.Status == StatusNSFW }

// IsCanceled reports whether the request was canceled.
func (r *Response) IsCanceled() bool { return r.Status == StatusCanceled }

// IsTerminal reports whether the request has reached a final state and will
// not change further (completed, failed, nsfw, or canceled).
func (r *Response) IsTerminal() bool {
	switch r.Status {
	case StatusCompleted, StatusFailed, StatusNSFW, StatusCanceled:
		return true
	default:
		return false
	}
}

func decodeResponse(data []byte) (*Response, error) {
	var r Response
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("higgsfield: decode response: %w", err)
	}
	r.Raw = append(json.RawMessage(nil), data...)
	return &r, nil
}
