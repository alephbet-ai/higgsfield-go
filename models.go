package higgsfield

// Model identifies a generation endpoint. It is an OPEN string type: the
// exported constants below cover the documented catalog, but any endpoint
// string is accepted, so newly released models work without an SDK update.
//
// There is no runtime "list models" endpoint; this catalog is derived from the
// API's OpenAPI specification (https://docs.higgsfield.ai/docs/openapi.json) and
// should be regenerated from it when models change.
type Model = string

// Known model endpoints, grouped by family. A leading slash is optional; the
// client normalizes it.
const (
	// Higgsfield Soul (text-to-image).
	ModelSoulStandard  Model = "/higgsfield-ai/soul/standard"
	ModelSoulCharacter Model = "/higgsfield-ai/soul/character"
	ModelSoulReference Model = "/higgsfield-ai/soul/reference"

	// Higgsfield DoP (image-to-video).
	ModelDoPTurbo    Model = "/higgsfield-ai/dop/turbo"
	ModelDoPLite     Model = "/higgsfield-ai/dop/lite"
	ModelDoPStandard Model = "/higgsfield-ai/dop/standard"

	// Higgsfield Popcorn.
	ModelPopcornAuto Model = "/higgsfield-ai/popcorn/auto"

	// Google Veo 3.1.
	ModelVeo31                   Model = "/veo3.1"
	ModelVeo31Image2Video        Model = "/veo3.1/image-to-video"
	ModelVeo31FirstLastFrame     Model = "/veo3.1/first-last-frame-to-video"
	ModelVeo31Fast               Model = "/veo3.1/fast"
	ModelVeo31FastImage2Video    Model = "/veo3.1/fast/image-to-video"
	ModelVeo31FastFirstLastFrame Model = "/veo3.1/fast/first-last-frame-to-video"
	ModelVeo31Reference2Video    Model = "/veo3.1/reference-to-video"

	// ByteDance Seedance v1.
	ModelSeedanceLiteImage2Video    Model = "/bytedance/seedance/v1/lite/image-to-video"
	ModelSeedanceLiteText2Video     Model = "/bytedance/seedance/v1/lite/text-to-video"
	ModelSeedanceProFastImage2Video Model = "/bytedance/seedance/v1/pro/fast/image-to-video"
	ModelSeedanceProFastText2Video  Model = "/bytedance/seedance/v1/pro/fast/text-to-video"

	// FLUX Pro Kontext Max (text-to-image).
	ModelFluxKontextMaxText2Image Model = "/flux-pro/kontext/max/text-to-image"

	// Kling Video v2.1 / v2.5-turbo.
	ModelKlingV21MasterImage2Video        Model = "/kling-video/v2.1/master/image-to-video"
	ModelKlingV21MasterText2Video         Model = "/kling-video/v2.1/master/text-to-video"
	ModelKlingV21ProImage2Video           Model = "/kling-video/v2.1/pro/image-to-video"
	ModelKlingV21StandardImage2Video      Model = "/kling-video/v2.1/standard/image-to-video"
	ModelKlingV25TurboProImage2Video      Model = "/kling-video/v2.5-turbo/pro/image-to-video"
	ModelKlingV25TurboProText2Video       Model = "/kling-video/v2.5-turbo/pro/text-to-video"
	ModelKlingV25TurboStandardImage2Video Model = "/kling-video/v2.5-turbo/standard/image-to-video"
)
