package audio

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestWAVDecoder_16BitStereoSampleRange(t *testing.T) {
	path := filepath.Join("..", "testdata", "wav", "stereo_44100_16bit.wav")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open test fixture: %v", err)
	}
	defer f.Close()

	dec := &WAVDecoder{}
	samples, rate, ch, err := dec.Decode(f)
	if err != nil {
		t.Fatalf("Decode: unexpected error: %v", err)
	}
	if rate != 44100 {
		t.Errorf("expected sampleRate 44100, got %d", rate)
	}
	if ch != 2 {
		t.Errorf("expected channels 2, got %d", ch)
	}
	if len(samples) == 0 {
		t.Fatal("Decode: expected non-empty samples")
	}
	for i, s := range samples {
		if math.IsNaN(s) {
			t.Errorf("sample[%d] is NaN", i)
		}
		if s < -1.0 || s > 1.0 {
			t.Errorf("sample[%d] = %f out of range [-1.0, 1.0]", i, s)
		}
	}
}

func TestWAVDecoder_24BitDecodes(t *testing.T) {
	path := filepath.Join("..", "testdata", "wav", "stereo_44100_24bit.wav")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open test fixture: %v", err)
	}
	defer f.Close()

	dec := &WAVDecoder{}
	samples, rate, ch, err := dec.Decode(f)
	if err != nil {
		t.Fatalf("Decode: unexpected error for 24-bit WAV: %v", err)
	}
	if rate != 44100 {
		t.Errorf("expected sampleRate 44100, got %d", rate)
	}
	if ch != 2 {
		t.Errorf("expected channels 2, got %d", ch)
	}
	if len(samples) == 0 {
		t.Fatal("Decode: expected non-empty samples for 24-bit")
	}
	for i, s := range samples {
		if s < -1.0 || s > 1.0 {
			t.Errorf("sample[%d] = %f out of range [-1.0, 1.0]", i, s)
		}
	}
}

func TestWAVDecoder_32BitDecodes(t *testing.T) {
	path := filepath.Join("..", "testdata", "wav", "stereo_44100_32bit.wav")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open test fixture: %v", err)
	}
	defer f.Close()

	dec := &WAVDecoder{}
	samples, rate, ch, err := dec.Decode(f)
	if err != nil {
		t.Fatalf("Decode: unexpected error for 32-bit WAV: %v", err)
	}
	if rate != 44100 {
		t.Errorf("expected sampleRate 44100, got %d", rate)
	}
	if ch != 2 {
		t.Errorf("expected channels 2, got %d", ch)
	}
	if len(samples) == 0 {
		t.Fatal("Decode: expected non-empty samples for 32-bit")
	}
	for i, s := range samples {
		if s < -1.0 || s > 1.0 {
			t.Errorf("sample[%d] = %f out of range [-1.0, 1.0]", i, s)
		}
	}
}

func TestWAVDecoder_Unsupported8BitReturnsError(t *testing.T) {
	path := filepath.Join("..", "testdata", "wav", "unsupported_8bit.wav")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open test fixture: %v", err)
	}
	defer f.Close()

	dec := &WAVDecoder{}
	_, _, _, err = dec.Decode(f)
	if err == nil {
		t.Fatal("Decode: expected error for 8-bit WAV, got nil")
	}
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("Decode: expected ErrUnsupportedFormat for 8-bit, got %v", err)
	}
}
