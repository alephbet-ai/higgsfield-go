package higgsfield

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeClock advances virtual time on each After call and never really sleeps,
// so retry/poll loops run instantly and deterministically.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock { return &fakeClock{now: time.Unix(0, 0)} }

func (f *fakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *fakeClock) After(d time.Duration) <-chan time.Time {
	f.mu.Lock()
	f.now = f.now.Add(d)
	now := f.now
	f.mu.Unlock()
	ch := make(chan time.Time, 1)
	ch <- now
	return ch
}

func testClient(t *testing.T, baseURL string, opts ...Option) *Client {
	t.Helper()
	all := append([]Option{WithBaseURL(baseURL), WithCredentialPair("id", "secret")}, opts...)
	c, err := New(all...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	c.clock = newFakeClock()
	return c
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func TestSubscribePollsToCompletion(t *testing.T) {
	var statusCalls int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /higgsfield-ai/soul/standard", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Key id:secret" {
			t.Errorf("Authorization = %q, want %q", got, "Key id:secret")
		}
		if got := r.Header.Get("User-Agent"); got != defaultUserAgent {
			t.Errorf("User-Agent = %q, want %q", got, defaultUserAgent)
		}
		writeJSON(w, 200, `{"status":"queued","request_id":"req-123","status_url":"/requests/req-123/status"}`)
	})
	mux.HandleFunc("GET /requests/req-123/status", func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&statusCalls, 1) < 3 {
			writeJSON(w, 200, `{"status":"in_progress","request_id":"req-123"}`)
			return
		}
		writeJSON(w, 200, `{"status":"completed","request_id":"req-123","images":[{"url":"https://cdn/x.jpg"}]}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := testClient(t, srv.URL)

	var updates []Status
	res, err := c.Subscribe(context.Background(), ModelSoulStandard,
		Input{"prompt": "a lake"},
		WithOnUpdate(func(r *Response) { updates = append(updates, r.Status) }),
	)
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if !res.IsCompleted() {
		t.Fatalf("status = %q, want completed", res.Status)
	}
	if len(res.Images) != 1 || res.Images[0].URL != "https://cdn/x.jpg" {
		t.Fatalf("images = %+v, want one image with url", res.Images)
	}
	if len(res.Raw) == 0 {
		t.Error("Raw not populated")
	}
	// initial (queued) + two in_progress + completed
	if len(updates) < 2 || updates[0] != StatusQueued || updates[len(updates)-1] != StatusCompleted {
		t.Errorf("updates = %v", updates)
	}
}

func TestSubmitNoPolling(t *testing.T) {
	var statusHit bool
	mux := http.NewServeMux()
	mux.HandleFunc("POST /higgsfield-ai/soul/standard", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"status":"queued","request_id":"req-1"}`)
	})
	mux.HandleFunc("GET /requests/req-1/status", func(w http.ResponseWriter, r *http.Request) {
		statusHit = true
		writeJSON(w, 200, `{"status":"completed"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := testClient(t, srv.URL)
	res, err := c.Submit(context.Background(), ModelSoulStandard, Input{"prompt": "x"})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if res.RequestID != "req-1" || !res.IsQueued() {
		t.Fatalf("res = %+v", res)
	}
	if statusHit {
		t.Error("Submit must not poll the status endpoint")
	}
}

func TestStatusAndCancel(t *testing.T) {
	var canceled bool
	mux := http.NewServeMux()
	mux.HandleFunc("GET /requests/abc/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"status":"in_progress","request_id":"abc"}`)
	})
	mux.HandleFunc("POST /requests/abc/cancel", func(w http.ResponseWriter, r *http.Request) {
		canceled = true
		writeJSON(w, 200, `{"status":"canceled"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := testClient(t, srv.URL)
	res, err := c.Status(context.Background(), "abc")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !res.IsInProgress() {
		t.Fatalf("status = %q", res.Status)
	}
	if err := c.Cancel(context.Background(), "abc"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if !canceled {
		t.Error("cancel endpoint not called")
	}
}

func TestErrorMapping(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		body     string
		wantKind ErrorKind
		wantIs   error
		wantMsg  string
	}{
		{"auth", 401, `{"detail":"nope"}`, KindAuthentication, ErrAuthentication, "nope"},
		{"credits", 403, ``, KindNotEnoughCredits, ErrNotEnoughCredits, "not enough credits"},
		{"validation", 422, `{"detail":[{"type":"missing","loc":["body","prompt"],"msg":"field required"}]}`, KindValidation, ErrValidation, "body.prompt: field required"},
		{"badinput", 400, `{"detail":"check params"}`, KindBadInput, ErrBadInput, "check params"},
		{"other", 418, `{"message":"teapot"}`, KindAPI, nil, "teapot"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("GET /requests/x/status", func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, tc.status, tc.body)
			})
			srv := httptest.NewServer(mux)
			defer srv.Close()

			c := testClient(t, srv.URL, WithMaxRetries(0))
			_, err := c.Status(context.Background(), "x")
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("err = %v, want *APIError", err)
			}
			if apiErr.Kind != tc.wantKind {
				t.Errorf("Kind = %d, want %d", apiErr.Kind, tc.wantKind)
			}
			if apiErr.StatusCode != tc.status {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tc.status)
			}
			if tc.wantMsg != "" && !strings.Contains(apiErr.Message, tc.wantMsg) {
				t.Errorf("Message = %q, want contains %q", apiErr.Message, tc.wantMsg)
			}
			if tc.wantIs != nil && !errors.Is(err, tc.wantIs) {
				t.Errorf("errors.Is(%v) = false", tc.wantIs)
			}
			if tc.name == "validation" && len(apiErr.Details) != 1 {
				t.Errorf("Details = %+v, want 1", apiErr.Details)
			}
		})
	}
}

func TestRetryOn500ThenSuccess(t *testing.T) {
	var calls int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /higgsfield-ai/soul/standard", func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) <= 2 {
			writeJSON(w, 500, `{"detail":"server error"}`)
			return
		}
		writeJSON(w, 200, `{"status":"queued","request_id":"ok"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := testClient(t, srv.URL, WithMaxRetries(3))
	res, err := c.Submit(context.Background(), ModelSoulStandard, Input{"p": 1})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if res.RequestID != "ok" {
		t.Fatalf("res = %+v", res)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("calls = %d, want 3", got)
	}
}

func TestRetryExhausted(t *testing.T) {
	var calls int32
	mux := http.NewServeMux()
	mux.HandleFunc("GET /requests/x/status", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		writeJSON(w, 503, `{"detail":"down"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := testClient(t, srv.URL, WithMaxRetries(2))
	_, err := c.Status(context.Background(), "x")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 503 {
		t.Fatalf("err = %v, want 503 APIError", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 { // initial + 2 retries
		t.Errorf("calls = %d, want 3", got)
	}
}

func TestUpload(t *testing.T) {
	var putBody []byte
	var putAuth string
	var putCT string
	mux := http.NewServeMux()
	mux.HandleFunc("POST /files/generate-upload-url", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"upload_url":"`+uploadURLBase(r)+`/put","public_url":"https://cdn/x.png","upload_headers":{"Content-Type":"image/png","x-amz-acl":"private"}}`)
	})
	mux.HandleFunc("PUT /put", func(w http.ResponseWriter, r *http.Request) {
		putAuth = r.Header.Get("Authorization")
		putCT = r.Header.Get("Content-Type")
		putBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(200)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := testClient(t, srv.URL)
	url, err := c.Upload(context.Background(), []byte("PNGDATA"), "image/png")
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if url != "https://cdn/x.png" {
		t.Errorf("url = %q", url)
	}
	if string(putBody) != "PNGDATA" {
		t.Errorf("put body = %q", putBody)
	}
	if putAuth != "" {
		t.Errorf("presigned PUT must not receive Authorization, got %q", putAuth)
	}
	if putCT != "image/png" {
		t.Errorf("put Content-Type = %q", putCT)
	}
}

func TestWebhookQueryParam(t *testing.T) {
	var gotHook string
	mux := http.NewServeMux()
	mux.HandleFunc("POST /higgsfield-ai/soul/standard", func(w http.ResponseWriter, r *http.Request) {
		gotHook = r.URL.Query().Get("hf_webhook")
		writeJSON(w, 200, `{"status":"completed","request_id":"w"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := testClient(t, srv.URL)
	_, err := c.Subscribe(context.Background(), ModelSoulStandard, Input{"p": 1},
		WithWebhook("https://example.com/cb?a=1", "sekret"))
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if gotHook != "https://example.com/cb?a=1" {
		t.Errorf("hf_webhook = %q", gotHook)
	}
}

func TestPollTimeout(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /higgsfield-ai/soul/standard", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"status":"queued","request_id":"slow"}`)
	})
	mux.HandleFunc("GET /requests/slow/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"status":"in_progress"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := testClient(t, srv.URL, WithMaxPollTime(10*time.Second), WithPollInterval(2*time.Second))
	_, err := c.Subscribe(context.Background(), ModelSoulStandard, Input{"p": 1})
	if !errors.Is(err, ErrPollTimeout) {
		t.Fatalf("err = %v, want ErrPollTimeout", err)
	}
}

func TestContextCancellation(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /higgsfield-ai/soul/standard", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"status":"queued"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := testClient(t, srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.Submit(ctx, ModelSoulStandard, Input{"p": 1})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestCredentialResolution(t *testing.T) {
	clearCredEnv(t)

	t.Run("missing", func(t *testing.T) {
		clearCredEnv(t)
		_, err := New()
		if !errors.Is(err, ErrCredentialsMissing) {
			t.Fatalf("err = %v, want ErrCredentialsMissing", err)
		}
	})

	t.Run("combined_env", func(t *testing.T) {
		clearCredEnv(t)
		t.Setenv("HF_CREDENTIALS", "envid:envsecret")
		c, err := New()
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if c.keyID != "envid" || c.keySecret != "envsecret" {
			t.Errorf("creds = %q:%q", c.keyID, c.keySecret)
		}
	})

	t.Run("pair_env", func(t *testing.T) {
		clearCredEnv(t)
		t.Setenv("HF_API_KEY_ID", "id2")
		t.Setenv("HF_API_KEY_SECRET", "sec2")
		c, err := New()
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if c.keyID != "id2" || c.keySecret != "sec2" {
			t.Errorf("creds = %q:%q", c.keyID, c.keySecret)
		}
	})

	t.Run("bad_format", func(t *testing.T) {
		clearCredEnv(t)
		_, err := New(WithCredentials("no-colon-here"))
		if err == nil || errors.Is(err, ErrCredentialsMissing) {
			t.Fatalf("err = %v, want a format error", err)
		}
	})

	t.Run("secret_with_colon", func(t *testing.T) {
		clearCredEnv(t)
		c, err := New(WithCredentials("id:sec:ret"))
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if c.keyID != "id" || c.keySecret != "sec:ret" {
			t.Errorf("creds = %q:%q", c.keyID, c.keySecret)
		}
	})
}

func clearCredEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"HF_CREDENTIALS", "HF_KEY", "HF_API_KEY_ID", "HF_API_KEY_SECRET", "HF_API_KEY", "HF_API_SECRET"} {
		t.Setenv(k, "")
	}
}

func TestNormalizeEndpoint(t *testing.T) {
	cases := map[string]string{
		"/a/b":  "/a/b",
		"a/b":   "/a/b",
		"":      "",
		"model": "/model",
	}
	for in, want := range cases {
		if got := normalizeEndpoint(in); got != want {
			t.Errorf("normalizeEndpoint(%q) = %q, want %q", in, got, want)
		}
	}
}

func uploadURLBase(r *http.Request) string {
	return "http://" + r.Host
}
