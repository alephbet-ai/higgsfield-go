# higgsfield-go

An idiomatic Go client for the [Higgsfield AI](https://higgsfield.ai) generation API.
It is a Go port of the official [higgsfield-js](https://github.com/higgsfield-ai/higgsfield-js)
SDK, covering the v2 `subscribe` flow and file uploads.

- Single `*Client`; every network call takes a `context.Context`
- Functional options for configuration
- Automatic polling to a terminal state, or fire-and-track manually
- Typed errors (`errors.Is` / `errors.As` friendly)
- **Zero dependencies** — standard library only

## Install

```bash
go get github.com/alephbet-ai/higgsfield-go
```

Requires Go 1.23+.

## Authentication

Credentials are `KEY_ID:KEY_SECRET`. Provide them via options or the environment
(resolved in this order): `HF_CREDENTIALS` or `HF_KEY` (`"KEY_ID:KEY_SECRET"`),
then `HF_API_KEY_ID` + `HF_API_KEY_SECRET`, then legacy `HF_API_KEY` + `HF_API_SECRET`.

```bash
export HF_CREDENTIALS="YOUR_KEY_ID:YOUR_KEY_SECRET"
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	higgsfield "github.com/alephbet-ai/higgsfield-go"
)

func main() {
	c, err := higgsfield.New() // reads credentials from the environment
	if err != nil {
		log.Fatal(err)
	}

	res, err := c.Subscribe(context.Background(), higgsfield.ModelSoulStandard, higgsfield.Input{
		"prompt": "A quiet alpine lake at sunrise, editorial photography",
	})
	if err != nil {
		log.Fatal(err)
	}

	if res.IsCompleted() && len(res.Images) > 0 {
		fmt.Println("image:", res.Images[0].URL)
	}
}
```

## Usage

### Configure the client

```go
c, err := higgsfield.New(
	higgsfield.WithCredentials("KEY_ID:KEY_SECRET"),
	higgsfield.WithTimeout(120*time.Second),
	higgsfield.WithMaxRetries(3),
	higgsfield.WithPollInterval(2*time.Second),
	higgsfield.WithMaxPollTime(5*time.Minute),
)
```

### Track progress with a callback

```go
res, err := c.Subscribe(ctx, higgsfield.ModelFluxKontextMaxText2Image,
	higgsfield.Input{"prompt": "a sunset", "aspect_ratio": "9:16"},
	higgsfield.WithOnUpdate(func(r *higgsfield.Response) {
		log.Println("status:", r.Status)
	}),
)
```

### Submit and poll manually

```go
res, err := c.Submit(ctx, higgsfield.ModelSoulStandard, input) // no polling
// ... later:
res, err = c.Status(ctx, res.RequestID)
err = c.Cancel(ctx, res.RequestID) // only queued requests can be canceled
```

Any endpoint string works, so newly released models need no SDK update:

```go
res, err := c.Subscribe(ctx, "/some/brand-new/model", input)
```

### Webhooks

```go
res, err := c.Subscribe(ctx, model, input,
	higgsfield.WithWebhook("https://example.com/callback", "secret"),
)
```

### File uploads

Upload assets to Higgsfield's CDN and pass the returned URL into a model input:

```go
url, err := c.UploadFile(ctx, "avatar.png")
img, _ := higgsfield.InputImage(url)

res, err := c.Subscribe(ctx, higgsfield.ModelDoPTurbo, higgsfield.Input{
	"model":        higgsfield.DoPModelTurbo,
	"prompt":       "cinematic camera movement",
	"input_images": []higgsfield.ImageRef{img},
})
```

Other upload helpers: `Upload(ctx, data, contentType)` and
`UploadImage(ctx, data, "png")`.

### Error handling

```go
_, err := c.Subscribe(ctx, model, input)
switch {
case errors.Is(err, higgsfield.ErrAuthentication):
	// invalid credentials
case errors.Is(err, higgsfield.ErrNotEnoughCredits):
	// top up account
case errors.Is(err, higgsfield.ErrValidation), errors.Is(err, higgsfield.ErrBadInput):
	var apiErr *higgsfield.APIError
	errors.As(err, &apiErr)
	fmt.Println(apiErr.Details) // field-level details, when provided
case errors.Is(err, higgsfield.ErrPollTimeout):
	// exceeded WithMaxPollTime
}
```

## Models

There is no runtime model-discovery endpoint; `Model` is an open string type with
exported constants for the documented catalog (`ModelSoulStandard`, `ModelDoPTurbo`,
`ModelFluxKontextMaxText2Image`, `ModelVeo31Image2Video`, `ModelKling*`, `ModelSeedance*`, …).
Because the type is open, any endpoint string is also accepted.

## Status

Initial version. Scope: v2 `subscribe` + uploads + helpers. Not yet ported from
higgsfield-js: the deprecated v1 `generate`/`JobSet`, motions/styles listing, and
SoulId endpoints.

## License

[MIT](LICENSE)
