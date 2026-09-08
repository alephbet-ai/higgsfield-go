package higgsfield

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

// Upload stores raw bytes on Higgsfield's CDN and returns the public URL to
// pass into a model input (e.g. as image_url or audio_url). contentType is the
// MIME type of the data (e.g. "image/jpeg").
func (c *Client) Upload(ctx context.Context, data []byte, contentType string) (string, error) {
	publicURL, uploadURL, uploadHeaders, err := c.generateUploadURL(ctx, contentType)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("higgsfield: build upload request: %w", err)
	}
	// The presigned URL must receive exactly the headers the API specifies and
	// must NOT receive Higgsfield credentials.
	for k, v := range uploadHeaders {
		req.Header.Set(k, v)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("higgsfield: upload PUT: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", newAPIError(resp.StatusCode, body)
	}
	return publicURL, nil
}

// UploadFile reads a local file and uploads it, guessing the content type from
// the file extension.
func (c *Client) UploadFile(ctx context.Context, path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("higgsfield: read file: %w", err)
	}
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return c.Upload(ctx, data, contentType)
}

// UploadImage uploads image bytes with content type "image/<format>". format
// defaults to "jpeg" when empty (e.g. "jpeg", "png", "webp").
func (c *Client) UploadImage(ctx context.Context, data []byte, format string) (string, error) {
	if format == "" {
		format = "jpeg"
	}
	return c.Upload(ctx, data, "image/"+format)
}

func (c *Client) generateUploadURL(ctx context.Context, contentType string) (publicURL, uploadURL string, headers map[string]string, err error) {
	data, err := c.do(ctx, http.MethodPost, "/files/generate-upload-url", map[string]string{
		"content_type": contentType,
	})
	if err != nil {
		return "", "", nil, err
	}

	var out struct {
		UploadURL     string            `json:"upload_url"`
		PublicURL     string            `json:"public_url"`
		UploadHeaders map[string]string `json:"upload_headers"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", "", nil, fmt.Errorf("higgsfield: decode upload URL response: %w", err)
	}
	return out.PublicURL, out.UploadURL, out.UploadHeaders, nil
}
