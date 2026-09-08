package higgsfield

import (
	"math/rand"
	"strings"
)

// ImageRef is an image input reference (type "image_url").
type ImageRef struct {
	Type     string `json:"type"`
	ImageURL string `json:"image_url"`
}

// AudioRef is an audio input reference (type "audio_url").
type AudioRef struct {
	Type     string `json:"type"`
	AudioURL string `json:"audio_url"`
}

// MotionRef is a motion input for image-to-video generation.
type MotionRef struct {
	ID       string  `json:"id"`
	Strength float64 `json:"strength"`
}

func badInput(msg string) *APIError {
	return &APIError{Kind: KindBadInput, Message: msg}
}

// InputImage builds an image reference from a URL.
func InputImage(imageURL string) (ImageRef, error) {
	if strings.TrimSpace(imageURL) == "" {
		return ImageRef{}, badInput("image URL must be a non-empty string")
	}
	return ImageRef{Type: "image_url", ImageURL: imageURL}, nil
}

// InputAudio builds an audio reference from a URL (WAV is expected by the API).
func InputAudio(audioURL string) (AudioRef, error) {
	if strings.TrimSpace(audioURL) == "" {
		return AudioRef{}, badInput("audio URL must be a non-empty string")
	}
	return AudioRef{Type: "audio_url", AudioURL: audioURL}, nil
}

// Motion builds a motion input from a motion id and a strength in [0, 1].
func Motion(motionID string, strength float64) (MotionRef, error) {
	if strings.TrimSpace(motionID) == "" {
		return MotionRef{}, badInput("motion ID must be a non-empty string")
	}
	s, err := Strength(strength)
	if err != nil {
		return MotionRef{}, err
	}
	return MotionRef{ID: motionID, Strength: s}, nil
}

// Strength validates a strength value, which must be in [0, 1].
func Strength(value float64) (float64, error) {
	if value < 0 || value > 1 {
		return 0, badInput("strength must be between 0 and 1")
	}
	return value, nil
}

// MaxSeed is the largest accepted seed value.
const MaxSeed = 1_000_000

// Seed validates a seed value, which must be in [0, MaxSeed].
func Seed(value int) (int, error) {
	if value < 0 || value > MaxSeed {
		return 0, badInput("seed must be an integer between 0 and 1,000,000")
	}
	return value, nil
}

// RandomSeed returns a random valid seed in [0, MaxSeed].
func RandomSeed() int {
	return rand.Intn(MaxSeed + 1)
}

// Soul (text-to-image) quality presets.
const (
	SoulQualitySD = "720p"
	SoulQualityHD = "1080p"
)

// Soul batch sizes.
const (
	BatchSizeSingle = 1
	BatchSizeQuad   = 4
)

// DoP (image-to-video) model presets.
const (
	DoPModelLite     = "dop-lite"
	DoPModelTurbo    = "dop-turbo"
	DoPModelStandard = "dop-standard"
)

// Speak (speech-to-video) quality presets.
const (
	SpeakQualityMid  = "mid"
	SpeakQualityHigh = "high"
)

// Speak durations, in seconds.
const (
	SpeakDurationShort  = 5
	SpeakDurationMedium = 10
	SpeakDurationLong   = 15
)

// Soul (text-to-image) resolutions, as width_and_height values.
const (
	SoulSizeLandscape2048x1152 = "2048x1152"
	SoulSizeLandscape2048x1536 = "2048x1536"
	SoulSizeLandscape2016x1344 = "2016x1344"
	SoulSizeLandscape1696x960  = "1696x960"
	SoulSizeLandscape1632x1088 = "1632x1088"
	SoulSizePortrait1152x2048  = "1152x2048"
	SoulSizePortrait1536x2048  = "1536x2048"
	SoulSizePortrait1344x2016  = "1344x2016"
	SoulSizePortrait960x1696   = "960x1696"
	SoulSizePortrait1088x1632  = "1088x1632"
	SoulSizeSquare1536x1536    = "1536x1536"
	SoulSizeMixed1536x1152     = "1536x1152"
	SoulSizeMixed1152x1536     = "1152x1536"
)
