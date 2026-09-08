# Development Plan — `higgsfield-go`

A Go port of the [higgsfield-js](https://github.com/higgsfield-ai/higgsfield-js) SDK.
Scope is locked to the **v2 `subscribe` flow + file uploads**, built as an **idiomatic Go**
library.

- **Module path:** `github.com/alephbet-ai/higgsfield-go`
- **Package:** `higgsfield`
- **Upstream:** higgsfield-js (TypeScript), v2 client surface
- **License:** MIT (match upstream)

---

## 1. Goals & non-goals

**Goals**
- Faithful *behavior* parity with the JS v2 client (endpoints, auth, request/response
  shapes, retry/poll semantics).
- Idiomatic Go surface: single `*Client`, `context.Context` on every network call,
  functional options, typed responses/errors.
- **Zero external dependencies** — stdlib `net/http` + `encoding/json` only. (JS uses
  axios + form-data; neither is needed in Go.)
- Fast, deterministic tests via `httptest` and an injectable clock (no real sleeping).

**Non-goals (out of scope for this pass)**
- v1 `generate()` / `JobSet`, `getMotions`, `getSoulStyles`, `createSoulId`/`listSoulIds`.
- v2 schema-loader (`createTypeMapFromSchemas` / `validateInputAgainstSchema`).
- Browser guard (`BrowserNotSupportedError` — meaningless in Go).

---

## 2. What we're porting (JS → Go map)

| JS (higgsfield-js) | Go port | Notes |
|---|---|---|
| `createHiggsfieldClient(cfg)` / `config()` singleton | `New(...Option) (*Client, error)` | No global mutable singleton |
| `subscribe(endpoint, {input, webhook, withPolling})` | `Client.Subscribe(ctx, model, input, ...SubscribeOption)` | `model` is a `Model` typed-string; submit + poll to terminal |
| (submit half of subscribe) | `Client.Submit(ctx, endpoint, input, ...)` | no polling; returns initial `*Response` |
| `pollV2Request` / status endpoint | `Client.Status(ctx, requestID)` + internal poll loop | GET `/requests/{id}/status` |
| `V2Response` | `Response` struct + `Status` enum | `IsCompleted()`, `IsFailed()`, … |
| `upload` / `uploadImage` + `/files/generate-upload-url` | `Upload`, `UploadFile`, `UploadImage` | presigned PUT, no auth headers on PUT |
| `helpers.ts` (`InputImage`, `inputMotion`, `webhook`, `strength`, `seed`) | `helpers.go` builders + validators | typed input refs |
| constant maps (`SoulSize`, `DoPModel`, `BatchSize`, …) | typed constants | for building `input` |
| `errors.ts` classes | `errors.go` typed errors + sentinels | `errors.Is`/`errors.As` friendly |
| `retryWithBackoff` | `retry.go` | net errors + 5xx, exp backoff + jitter |
| `Config` / `ClientConfig` | functional options | defaults preserved |
| `auth.ts` `fetchCredentials` | credential resolution in `New` | env fallback chain |

---

## 3. Idiomatic Go API sketch

```go
// construction
c, err := higgsfield.New(
    higgsfield.WithCredentials("KEY_ID:KEY_SECRET"), // or WithCredentialPair(id, secret)
    higgsfield.WithBaseURL("https://platform.higgsfield.ai"), // default
    higgsfield.WithTimeout(120*time.Second),
    higgsfield.WithMaxRetries(3),
    higgsfield.WithPollInterval(2*time.Second),
    higgsfield.WithMaxPollTime(5*time.Minute),
    higgsfield.WithHTTPClient(myHTTPClient),
)
// creds fallback if no option:
//   HF_CREDENTIALS / HF_KEY ("id:secret") -> HF_API_KEY + HF_API_SECRET -> ErrCredentialsMissing

// generate + wait — model is a required positional arg (a Model typed-string constant)
res, err := c.Subscribe(ctx, higgsfield.ModelFluxKontextMax,
    higgsfield.Input{"prompt": "a sunset", "aspect_ratio": "9:16", "seed": 1234},
    higgsfield.WithWebhook("https://me/cb", "secret"), // optional
    higgsfield.WithOnUpdate(func(r *higgsfield.Response){ log.Println(r.Status) }), // optional progress
)
if res.IsCompleted() {
    fmt.Println(res.Images[0].URL)
}

// any endpoint string still works (open type) — new models need no SDK release
res, _ = c.Subscribe(ctx, "/some/brand-new/model", input)

// fire-and-track manually
res, _ := c.Submit(ctx, model, input)         // no polling
res, _ = c.Status(ctx, res.RequestID)         // one status read

// uploads
url, _ := c.Upload(ctx, data, "image/jpeg")
url, _ = c.UploadFile(ctx, "avatar.png")      // mime guessed from ext
url, _ = c.UploadImage(ctx, buf, "jpeg")      // -> content-type image/jpeg
```

```go
// Model identifies a generation endpoint. It is an OPEN string type: exported constants
// cover the documented catalog (generated from openapi.json), but any endpoint string is
// accepted, so new models work without an SDK release. Not a closed enum.
type Model = string
const (
    ModelSoulStandard   Model = "/higgsfield-ai/soul/standard"
    ModelDoPTurbo       Model = "/higgsfield-ai/dop/turbo"
    ModelFluxKontextMax Model = "/flux-pro/kontext/max/text-to-image"
    ModelVeo31Image2Vid Model = "/veo3.1/image-to-video"
    // … full catalog generated from the OpenAPI spec (see Section 9)
)

type Input map[string]any

type Response struct {
    Status    Status          `json:"status"`
    RequestID string          `json:"request_id"`
    StatusURL string          `json:"status_url"`
    CancelURL string          `json:"cancel_url"`
    Images    []Media         `json:"images,omitempty"`
    Video     *Media          `json:"video,omitempty"`
    Raw       json.RawMessage `json:"-"` // full body kept for forward-compat
}
type Media struct { URL string `json:"url"` }

type Status string
const (
    StatusQueued     Status = "queued"
    StatusInProgress Status = "in_progress"
    StatusCompleted  Status = "completed"
    StatusFailed     Status = "failed"
    StatusNSFW       Status = "nsfw"
    StatusCanceled   Status = "canceled"
)
func (r *Response) IsCompleted() bool // + IsFailed/IsNSFW/IsQueued/IsInProgress
func (r *Response) IsTerminal()  bool // completed|failed|nsfw|canceled
```

*(Optional, Go 1.23+): a `PollStatus(ctx, id) iter.Seq2[*Response, error]` range-over-func
variant that mirrors the JS generator style. The `WithOnUpdate` callback is the baseline so
we can keep a lower Go floor if preferred.)*

---

## 4. Proposed package layout

Single flat package (small library — no premature sub-packages):

```
higgsfield-go/
├── go.mod                 // module path; go 1.23; no require block
├── LICENSE                // MIT (match upstream)
├── README.md
├── doc.go                 // package overview / godoc
├── client.go             // Client, New, config struct, credential resolution
├── options.go            // Option + With* constructors + defaults + validate()
├── subscribe.go          // Subscribe, Submit, Status, poll loop
├── response.go           // Response, Media, Status, helper methods
├── upload.go             // Upload, UploadFile, UploadImage, mime guessing
├── helpers.go            // InputImage/InputAudio/Motion/Webhook/Strength/Seed + constants
├── errors.go             // typed errors + HTTP status -> error mapping
├── transport.go          // internal doer: header injection, decode, error mapping
├── retry.go              // retryWithBackoff (+ clock interface)
├── *_test.go             // per-file httptest-based tests
└── examples/
    └── text_to_image/main.go
```

---

## 5. Behavioral parity details (the tricky bits)

Each of these becomes a test:

1. **Auth header (v2):** `Authorization: Key {id}:{secret}` + `Content-Type: application/json`
   + our own `User-Agent: higgsfield-go/<version>` (we do **not** replicate JS's
   obfuscated UA).
2. **Request body:** `input` is serialized **directly** as the JSON body (not wrapped in
   `{params}` — that was v1).
3. **Webhook:** appended as `?hf_webhook=<url-encoded url>` query param; the secret is
   currently unused on the v2 path (documented).
4. **Endpoint normalization:** prepend `/` if missing.
5. **Poll semantics:** GET `/requests/{id}/status` every `pollInterval`; stop on terminal
   status; abort after `maxPollTime` (→ `ErrPollTimeout`) or `ctx` cancellation; **swallow
   5xx during polling and keep going** (matches JS). Deviation: we treat `canceled` as
   terminal too (JS's v2 loop omits it and could spin) — documented.
6. **Retry:** retry on transport/network errors + HTTP >= 500 only;
   `delay = min(backoff·2^attempt + rand(0..1s), maxBackoff)`; caps at `maxRetries`.
7. **Uploads:** POST `/files/generate-upload-url {content_type}` → `{upload_url, public_url}`;
   then **PUT the bytes to the presigned URL with no SDK auth headers** (only `Content-Type`);
   return `public_url`.
8. **Credentials chain:** explicit option → `HF_CREDENTIALS`/`HF_KEY` (split on first `:`) →
   `HF_API_KEY`+`HF_API_SECRET` → `ErrCredentialsMissing`.
9. **Defaults:** timeout 120s, pollInterval 2s, maxPollTime 5m, maxRetries 3,
   retryBackoff 1s, retryMaxBackoff 60s, baseURL `https://platform.higgsfield.ai`.

---

## 6. Open questions / risks to resolve during implementation

- 🚧 **Base URL & endpoint namespace (top blocker).** The official OpenAPI spec
  (`docs.higgsfield.ai/docs/openapi.json`) uses base URL `https://api.higgsfield.ai` with
  paths like `/higgsfield-ai/soul/standard`, `/flux-pro/kontext/max/text-to-image`,
  `/veo3.1/image-to-video`. The JS SDK (and this plan's current defaults) use
  `https://platform.higgsfield.ai` with legacy `/v1/...` paths. The v2 SDK's example
  endpoints match the new `api.higgsfield.ai` catalog; the `/v1/...` paths look legacy. Pin
  down the correct base URL + path style (and whether `platform` redirects to `api`) before
  wiring `Subscribe`/`Status`; this drives both the default `baseURL` and the `Model`
  constant catalog (Section 9).
- ⚠️ **Upload endpoint auth (highest risk).** Uploads existed only on the v1 JS client,
  which authenticates with `hf-api-key`/`hf-secret` — *not* the v2 `Authorization: Key`
  header. Confirm which header `/files/generate-upload-url` accepts (check live API / docs,
  or send both). Blocks finalizing `upload.go`.
- Webhook secret on the v2 path — verify whether the API actually wants it (header/body) or
  if query-only `hf_webhook` is complete.
- Whether to fix the JS `canceled`-not-terminal quirk (planned: yes, documented deviation).
- Go floor: `go 1.23` (enables the optional `iter.Seq2` poller) vs `go 1.21`/`1.22`
  (callback only). Recommend **1.23** (local toolchain is 1.23.1).

---

## 7. Testing strategy

- `httptest.Server` for the platform API; a **second** server (or unauthenticated handler)
  for the presigned PUT to assert no auth headers leak.
- **Injectable clock** (a tiny `sleeper`/`now` interface) so retry backoff and poll intervals
  run instantly and deterministically in tests.
- Coverage targets: credential resolution chain; header injection; `Subscribe` happy path
  (`queued`→`in_progress`→`completed`); `Submit` without polling; each error mapping
  (401/403/422/400/5xx-retry-then-fail); retry backoff math; poll timeout via tiny
  `maxPollTime`; `ctx` cancellation mid-poll; upload two-step flow; webhook query param;
  endpoint normalization; helper validators (`Strength` bounds, `Seed` range/random).
- `go test -race`, `go vet`, and `golangci-lint` in CI.

---

## 8. Milestones (phased)

| # | Milestone | Contents |
|---|---|---|
| M0 | Scaffold | `go.mod`, MIT `LICENSE`, README skeleton, `doc.go`, GitHub Actions (build/vet/test-race/lint, Go 1.23) |
| M1 | Config + auth + errors | `options.go`, credential chain in `New`, `errors.go` + status mapping (+ tests) |
| M2 | Transport + retry | `transport.go`, `retry.go` with injectable clock (+ tests) |
| M3 | Subscribe/Submit/Status + polling | `subscribe.go`, `response.go` (+ tests) — **core value** |
| M4 | Uploads | `upload.go` (+ tests) — gated on the auth-header question above |
| M5 | Helpers + constants | `helpers.go` (+ tests) |
| M6 | Docs & release | godoc pass, README with examples, `examples/`, tag `v0.1.0` |

This is a small library; the core (M0–M3) is the bulk of the value and is achievable quickly.
M4–M6 are incremental.

---

## 9. Request model specification (design note)

There is **no formal model registry** in higgsfield-js — a "model" is just an **endpoint
string** passed to `subscribe(endpoint, { input })`, and `input` is typed `any`. The client
does not validate the endpoint name or the input at call time. The upstream SDK expresses the
request "spec" in three overlapping ways, **none of which is enforced by `subscribe`**:

1. **Static TypeScript interfaces (`src/v2/types.ts`)** for three known endpoints, wired via
   `EndpointInputMap`. Compile-time only (erased at runtime; not actually connected to
   `subscribe`, which uses `SubscribeOptions<any>`). Useful as field documentation:
   - `/v1/text2image/soul` → `SoulText2ImageInput`
     (`prompt`, `width_and_height`, `quality` `720p|1080p`, `batch_size` `1|4`, plus optional
     `style_id`/`style_strength`, `custom_reference_id`/`custom_reference_strength`,
     `image_reference`, `enhance_prompt`, `seed`).
   - `/v1/image2video/dop` → `DoPImage2VideoInput`
     (`model` `dop-lite|dop-turbo|dop-standard`, `prompt`, `input_images[]`, optional
     `motions[]`, `seed`, `enhance_prompt`).
   - `/v1/speak/higgsfield` → `SpeakVideoInput`
     (`input_image`, `input_audio` (WAV only), `prompt`, `quality` `mid|high`,
     `duration` `5|10|15`, optional `seed`).
2. **Dynamic JSON-Schema `ModelSchema` (`src/v2/schema-loader.ts`)** — the intended generic
   mechanism: backend-served `{ endpoint, name, inputSchema:{ properties, required } }` plus
   `createTypeMapFromSchemas()` and `validateInputAgainstSchema()` (checks required fields,
   types, enums). **Dead code upstream** — never called; the old `autoLoadSchemas` /
   `loadSchemasOnInit` options are deprecated no-ops.
3. **Constants + validating builders (`src/helpers.ts`)** — the only client-side validation
   that actually runs. Enumerations (`SoulSize`, `SoulQuality`, `BatchSize`, `DoPModel`,
   `SpeakVideoQuality`, `SpeakDuration`) and builders (`InputImage.fromUrl`,
   `InputAudio.fromUrl`, `inputMotion`, `webhook`, `strength` 0–1, `seed` 0–1,000,000) that
   throw `BadInputError` on invalid values.

Body wrapping is part of the spec too: **v2 `subscribe` sends `input` spread directly** as the
JSON body (webhook → `?hf_webhook=` query param), whereas v1 `generate` wraps it as
`{ params, webhook }`.

**Decision for the Go port:** mirror v2's open contract — `Input map[string]any`, endpoint as a
string, no schema enforcement (faithful and forward-compatible with new models). Port the
helper constants/builders from `helpers.ts` to guide correctness at the call site. Two additive
follow-ups remain available if desired later:
- **Typed input structs** for the three known endpoints (generic `Subscribe[T]` or per-endpoint
  helpers) for better ergonomics — at the cost of coupling to schemas that drift.
- **Runtime JSON-Schema validation** (port of `validateInputAgainstSchema`) — note this would be
  *new* behavior, not parity, since it is unused upstream.

### Model catalog

The API has **no runtime models-discovery endpoint** — the OpenAPI spec exposes only
`GET /requests/{id}/status` and `POST /requests/{id}/cancel` besides the 48 per-model POST
endpoints (`/higgsfield-ai/soul/standard`, `/nano-banana`, `/higgsfield-ai/dop/{turbo,lite,standard}`,
`/flux-pro/kontext/max/text-to-image`, `/veo3.1/*`, `/sora-2/*`, `/kling-video/*`, `/minimax/hailuo-*`,
`/reve/*`, `/wan-25-preview/*`, `/bytedance/seedance/*`, `/higgsfield-ai/popcorn/auto`). The dangling `ModelSchemasResponse` type in higgsfield-js is
not backed by any live endpoint.

Because the catalog is finite and versioned (it *is* the OpenAPI spec), models are exposed as a
`Model` typed-string with exported constants (see Section 3), passed positionally to
`Subscribe`/`Submit` — not as an optional `.WithModel()` (the model is required, not config).
The type stays **open** (users can pass any endpoint string) so new models need no release. To
keep constants honest, a tiny build-time script regenerates them from `openapi.json` using stdlib
`encoding/json` — no runtime dependency and no external client-generator involved. Exact paths and
base URL depend on resolving the namespace question in Section 6.
