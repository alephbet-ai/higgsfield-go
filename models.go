package higgsfield

// Model identifies a generation endpoint. It is an OPEN string type: the
// exported constants below cover the documented catalog, but any endpoint
// string is accepted, so newly released models work without an SDK update.
//
// There is no runtime "list models" endpoint; this catalog is derived from the
// API's OpenAPI specification (https://docs.higgsfield.ai/docs/openapi.json) and
// should be regenerated from it when models change. A leading slash is optional;
// the client normalizes it.
type Model = string

// Higgsfield Soul (text-to-image).
const (
	ModelSoulStandard  Model = "/higgsfield-ai/soul/standard"
	ModelSoulCharacter Model = "/higgsfield-ai/soul/character"
	ModelSoulReference Model = "/higgsfield-ai/soul/reference"
)

// Higgsfield DoP (image-to-video).
const (
	ModelDoPTurbo    Model = "/higgsfield-ai/dop/turbo"
	ModelDoPLite     Model = "/higgsfield-ai/dop/lite"
	ModelDoPStandard Model = "/higgsfield-ai/dop/standard"
)

// Higgsfield Popcorn.
const ModelPopcornAuto Model = "/higgsfield-ai/popcorn/auto"

// Google Nano Banana (Gemini image model).
const ModelNanoBanana Model = "/nano-banana"

// FLUX Pro Kontext Max (text-to-image).
const ModelFluxKontextMaxText2Image Model = "/flux-pro/kontext/max/text-to-image"

// Reve (text-to-image, edit, remix).
const (
	ModelReveText2Image Model = "/reve/text-to-image"
	ModelReveEdit       Model = "/reve/edit"
	ModelReveFastEdit   Model = "/reve/fast/edit"
	ModelReveRemix      Model = "/reve/remix"
	ModelReveFastRemix  Model = "/reve/fast/remix"
)

// Google Veo 3.1.
const (
	ModelVeo31                   Model = "/veo3.1"
	ModelVeo31Image2Video        Model = "/veo3.1/image-to-video"
	ModelVeo31FirstLastFrame     Model = "/veo3.1/first-last-frame-to-video"
	ModelVeo31Fast               Model = "/veo3.1/fast"
	ModelVeo31FastImage2Video    Model = "/veo3.1/fast/image-to-video"
	ModelVeo31FastFirstLastFrame Model = "/veo3.1/fast/first-last-frame-to-video"
	ModelVeo31Reference2Video    Model = "/veo3.1/reference-to-video"
)

// OpenAI Sora 2.
const (
	ModelSora2Text2Video     Model = "/sora-2/text-to-video"
	ModelSora2Text2VideoPro  Model = "/sora-2/text-to-video/pro"
	ModelSora2Image2Video    Model = "/sora-2/image-to-video"
	ModelSora2Image2VideoPro Model = "/sora-2/image-to-video/pro"
)

// ByteDance Seedance v1.
const (
	ModelSeedanceLiteImage2Video    Model = "/bytedance/seedance/v1/lite/image-to-video"
	ModelSeedanceLiteText2Video     Model = "/bytedance/seedance/v1/lite/text-to-video"
	ModelSeedanceProFastImage2Video Model = "/bytedance/seedance/v1/pro/fast/image-to-video"
	ModelSeedanceProFastText2Video  Model = "/bytedance/seedance/v1/pro/fast/text-to-video"
)

// Kling Video v2.1 / v2.5-turbo.
const (
	ModelKlingV21MasterImage2Video        Model = "/kling-video/v2.1/master/image-to-video"
	ModelKlingV21MasterText2Video         Model = "/kling-video/v2.1/master/text-to-video"
	ModelKlingV21ProImage2Video           Model = "/kling-video/v2.1/pro/image-to-video"
	ModelKlingV21StandardImage2Video      Model = "/kling-video/v2.1/standard/image-to-video"
	ModelKlingV25TurboProImage2Video      Model = "/kling-video/v2.5-turbo/pro/image-to-video"
	ModelKlingV25TurboProText2Video       Model = "/kling-video/v2.5-turbo/pro/text-to-video"
	ModelKlingV25TurboStandardImage2Video Model = "/kling-video/v2.5-turbo/standard/image-to-video"
)

// MiniMax Hailuo.
const (
	ModelHailuo02ProImage2Video          Model = "/minimax/hailuo-02/pro/image-to-video"
	ModelHailuo02ProText2Video           Model = "/minimax/hailuo-02/pro/text-to-video"
	ModelHailuo02StandardImage2Video     Model = "/minimax/hailuo-02/standard/image-to-video"
	ModelHailuo02StandardText2Video      Model = "/minimax/hailuo-02/standard/text-to-video"
	ModelHailuo23ProImage2Video          Model = "/minimax/hailuo-2.3/pro/image-to-video"
	ModelHailuo23ProText2Video           Model = "/minimax/hailuo-2.3/pro/text-to-video"
	ModelHailuo23StandardImage2Video     Model = "/minimax/hailuo-2.3/standard/image-to-video"
	ModelHailuo23StandardText2Video      Model = "/minimax/hailuo-2.3/standard/text-to-video"
	ModelHailuo23FastProImage2Video      Model = "/minimax/hailuo-2.3-fast/pro/image-to-video"
	ModelHailuo23FastStandardImage2Video Model = "/minimax/hailuo-2.3-fast/standard/image-to-video"
)

// Wan 2.5 preview.
const (
	ModelWan25PreviewImage2Video Model = "/wan-25-preview/image-to-video"
	ModelWan25PreviewText2Video  Model = "/wan-25-preview/text-to-video"
)

// Models lists every known model endpoint. Because Model is an open type, this
// is a convenience for discovery/enumeration, not an exhaustive constraint.
var Models = []Model{
	ModelSoulStandard,
	ModelSoulCharacter,
	ModelSoulReference,
	ModelDoPTurbo,
	ModelDoPLite,
	ModelDoPStandard,
	ModelPopcornAuto,
	ModelNanoBanana,
	ModelFluxKontextMaxText2Image,
	ModelReveText2Image,
	ModelReveEdit,
	ModelReveFastEdit,
	ModelReveRemix,
	ModelReveFastRemix,
	ModelVeo31,
	ModelVeo31Image2Video,
	ModelVeo31FirstLastFrame,
	ModelVeo31Fast,
	ModelVeo31FastImage2Video,
	ModelVeo31FastFirstLastFrame,
	ModelVeo31Reference2Video,
	ModelSora2Text2Video,
	ModelSora2Text2VideoPro,
	ModelSora2Image2Video,
	ModelSora2Image2VideoPro,
	ModelSeedanceLiteImage2Video,
	ModelSeedanceLiteText2Video,
	ModelSeedanceProFastImage2Video,
	ModelSeedanceProFastText2Video,
	ModelKlingV21MasterImage2Video,
	ModelKlingV21MasterText2Video,
	ModelKlingV21ProImage2Video,
	ModelKlingV21StandardImage2Video,
	ModelKlingV25TurboProImage2Video,
	ModelKlingV25TurboProText2Video,
	ModelKlingV25TurboStandardImage2Video,
	ModelHailuo02ProImage2Video,
	ModelHailuo02ProText2Video,
	ModelHailuo02StandardImage2Video,
	ModelHailuo02StandardText2Video,
	ModelHailuo23ProImage2Video,
	ModelHailuo23ProText2Video,
	ModelHailuo23StandardImage2Video,
	ModelHailuo23StandardText2Video,
	ModelHailuo23FastProImage2Video,
	ModelHailuo23FastStandardImage2Video,
	ModelWan25PreviewImage2Video,
	ModelWan25PreviewText2Video,
}
