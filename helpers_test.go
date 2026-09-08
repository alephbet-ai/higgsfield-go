package higgsfield

import (
	"errors"
	"testing"
)

func TestStrength(t *testing.T) {
	if _, err := Strength(0.5); err != nil {
		t.Errorf("Strength(0.5): %v", err)
	}
	for _, v := range []float64{-0.1, 1.1} {
		if _, err := Strength(v); !errors.Is(err, ErrBadInput) {
			t.Errorf("Strength(%v) err = %v, want ErrBadInput", v, err)
		}
	}
}

func TestSeed(t *testing.T) {
	if _, err := Seed(0); err != nil {
		t.Errorf("Seed(0): %v", err)
	}
	if _, err := Seed(MaxSeed); err != nil {
		t.Errorf("Seed(MaxSeed): %v", err)
	}
	for _, v := range []int{-1, MaxSeed + 1} {
		if _, err := Seed(v); !errors.Is(err, ErrBadInput) {
			t.Errorf("Seed(%d) err = %v, want ErrBadInput", v, err)
		}
	}
	if s := RandomSeed(); s < 0 || s > MaxSeed {
		t.Errorf("RandomSeed() = %d out of range", s)
	}
}

func TestInputBuilders(t *testing.T) {
	img, err := InputImage("https://x/y.jpg")
	if err != nil {
		t.Fatalf("InputImage: %v", err)
	}
	if img.Type != "image_url" || img.ImageURL != "https://x/y.jpg" {
		t.Errorf("img = %+v", img)
	}
	if _, err := InputImage("  "); !errors.Is(err, ErrBadInput) {
		t.Errorf("InputImage(blank) err = %v, want ErrBadInput", err)
	}

	au, err := InputAudio("https://x/a.wav")
	if err != nil {
		t.Fatalf("InputAudio: %v", err)
	}
	if au.Type != "audio_url" || au.AudioURL != "https://x/a.wav" {
		t.Errorf("audio = %+v", au)
	}

	m, err := Motion("motion-1", 0.8)
	if err != nil {
		t.Fatalf("Motion: %v", err)
	}
	if m.ID != "motion-1" || m.Strength != 0.8 {
		t.Errorf("motion = %+v", m)
	}
	if _, err := Motion("", 0.5); !errors.Is(err, ErrBadInput) {
		t.Errorf("Motion(blank) err = %v, want ErrBadInput", err)
	}
	if _, err := Motion("id", 2); !errors.Is(err, ErrBadInput) {
		t.Errorf("Motion(bad strength) err = %v, want ErrBadInput", err)
	}
}
