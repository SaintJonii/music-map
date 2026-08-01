package audio

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestMP3Decoder_SampleCountWithinTolerance(t *testing.T) {
	path := filepath.Join("..", "testdata", "mp3", "test_128kbps.mp3")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open test fixture: %v", err)
	}
	defer f.Close()

	dec := &MP3Decoder{}
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

	// 3 seconds at 44100 Hz stereo = 264600 interleaved samples.
	// MP3 encoders (LAME) add inherent encoder delay (~1105-1152 samples).
	// Combined with decoder padding, ±2% tolerance accommodates this.
	expectedSamples := 264600
	got := len(samples)
	delta := float64(got) / float64(expectedSamples)
	if delta < 0.98 || delta > 1.02 {
		t.Errorf("sample count %d is outside ±2%% of expected %d (delta=%.4f)", got, expectedSamples, delta)
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

func TestMP3Decoder_CorruptStreamReturnsError(t *testing.T) {
	path := filepath.Join("..", "testdata", "mp3", "corrupt.mp3")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open test fixture: %v", err)
	}
	defer f.Close()

	dec := &MP3Decoder{}
	_, _, _, err = dec.Decode(f)
	if err == nil {
		t.Fatal("Decode: expected error for corrupt MP3, got nil")
	}
}
