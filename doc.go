// Package higgsfield is a Go client for the Higgsfield AI generation API
// (https://higgsfield.ai). It is an idiomatic Go port of the official
// higgsfield-js SDK, covering the v2 "subscribe" flow and file uploads.
//
// A single [Client] is created with [New] and configured via functional
// options. Every network call takes a [context.Context]; run calls in your own
// goroutines for concurrency.
//
// Basic usage:
//
//	c, err := higgsfield.New(
//		higgsfield.WithCredentials("KEY_ID:KEY_SECRET"),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	res, err := c.Subscribe(ctx, higgsfield.ModelSoulStandard, higgsfield.Input{
//		"prompt": "A quiet alpine lake at sunrise, editorial photography",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	if res.IsCompleted() {
//		fmt.Println(res.Images[0].URL)
//	}
//
// Credentials are resolved from options first, then from the environment:
// HF_CREDENTIALS or HF_KEY ("KEY_ID:KEY_SECRET"), then HF_API_KEY_ID +
// HF_API_KEY_SECRET, then the legacy HF_API_KEY + HF_API_SECRET.
package higgsfield
