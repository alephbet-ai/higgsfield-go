package higgsfield

import (
	"sort"
	"strings"
	"testing"
)

// specModelEndpoints is the authoritative catalog of generation endpoints from
// the API's OpenAPI spec (https://docs.higgsfield.ai/docs/openapi.json), i.e.
// every POST path except the /requests/{id}/{status,cancel} lifecycle routes.
// Keep it in sync with the spec; the tests below verify Models matches it
// exactly, so a typo or a missing model fails the build.
var specModelEndpoints = []string{
	"/bytedance/seedance/v1/lite/image-to-video",
	"/bytedance/seedance/v1/lite/text-to-video",
	"/bytedance/seedance/v1/pro/fast/image-to-video",
	"/bytedance/seedance/v1/pro/fast/text-to-video",
	"/flux-pro/kontext/max/text-to-image",
	"/higgsfield-ai/dop/lite",
	"/higgsfield-ai/dop/standard",
	"/higgsfield-ai/dop/turbo",
	"/higgsfield-ai/popcorn/auto",
	"/higgsfield-ai/soul/character",
	"/higgsfield-ai/soul/reference",
	"/higgsfield-ai/soul/standard",
	"/kling-video/v2.1/master/image-to-video",
	"/kling-video/v2.1/master/text-to-video",
	"/kling-video/v2.1/pro/image-to-video",
	"/kling-video/v2.1/standard/image-to-video",
	"/kling-video/v2.5-turbo/pro/image-to-video",
	"/kling-video/v2.5-turbo/pro/text-to-video",
	"/kling-video/v2.5-turbo/standard/image-to-video",
	"/minimax/hailuo-02/pro/image-to-video",
	"/minimax/hailuo-02/pro/text-to-video",
	"/minimax/hailuo-02/standard/image-to-video",
	"/minimax/hailuo-02/standard/text-to-video",
	"/minimax/hailuo-2.3-fast/pro/image-to-video",
	"/minimax/hailuo-2.3-fast/standard/image-to-video",
	"/minimax/hailuo-2.3/pro/image-to-video",
	"/minimax/hailuo-2.3/pro/text-to-video",
	"/minimax/hailuo-2.3/standard/image-to-video",
	"/minimax/hailuo-2.3/standard/text-to-video",
	"/nano-banana",
	"/reve/edit",
	"/reve/fast/edit",
	"/reve/fast/remix",
	"/reve/remix",
	"/reve/text-to-image",
	"/sora-2/image-to-video",
	"/sora-2/image-to-video/pro",
	"/sora-2/text-to-video",
	"/sora-2/text-to-video/pro",
	"/veo3.1",
	"/veo3.1/fast",
	"/veo3.1/fast/first-last-frame-to-video",
	"/veo3.1/fast/image-to-video",
	"/veo3.1/first-last-frame-to-video",
	"/veo3.1/image-to-video",
	"/veo3.1/reference-to-video",
	"/wan-25-preview/image-to-video",
	"/wan-25-preview/text-to-video",
}

func TestModelsMatchSpec(t *testing.T) {
	got := append([]Model(nil), Models...)
	want := append([]string(nil), specModelEndpoints...)
	sort.Strings(got)
	sort.Strings(want)

	gotSet := map[string]bool{}
	for _, m := range got {
		if gotSet[m] {
			t.Errorf("duplicate model constant: %q", m)
		}
		gotSet[m] = true
		if !strings.HasPrefix(m, "/") {
			t.Errorf("model %q should start with /", m)
		}
	}

	wantSet := map[string]bool{}
	for _, m := range want {
		wantSet[m] = true
	}
	for _, m := range want {
		if !gotSet[m] {
			t.Errorf("missing model constant for spec endpoint %q", m)
		}
	}
	for _, m := range got {
		if !wantSet[m] {
			t.Errorf("Models has %q which is not in the OpenAPI spec", m)
		}
	}
	if len(Models) != len(specModelEndpoints) {
		t.Errorf("len(Models) = %d, want %d", len(Models), len(specModelEndpoints))
	}
}
