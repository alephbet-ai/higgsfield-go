// Command text_to_image demonstrates a basic text-to-image generation with the
// higgsfield-go client.
//
// Create a key at https://cloud.higgsfield.ai/api-keys, then set credentials
// first:
//
//	export HF_CREDENTIALS="YOUR_KEY_ID:YOUR_KEY_SECRET"
//	go run ./examples/text_to_image
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	higgsfield "github.com/alephbet-ai/higgsfield-go"
)

func main() {
	c, err := higgsfield.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	res, err := c.Subscribe(ctx, higgsfield.ModelSoulStandard,
		higgsfield.Input{
			"prompt": "A quiet alpine lake at sunrise, editorial photography",
		},
		higgsfield.WithOnUpdate(func(r *higgsfield.Response) {
			log.Println("status:", r.Status)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if !res.IsCompleted() {
		log.Fatalf("finished with status %q", res.Status)
	}
	for _, img := range res.Images {
		fmt.Println(img.URL)
	}
}
