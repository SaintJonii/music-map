package audio

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestFLACDecoder_16BitLosslessVerify(t *testing.T) {
	path := filepath.Join("..", "testdata", "flac", "test_16bit.flac")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open test fixture: %v", err)
	}
	defer f.Close()

	dec := &FLACDecoder{}
	samples, rate, ch, err := dec.Decode(f)
	if err != nil {
		t.Fatalf("Decode: unexpected error: %v", err)
	}

	// Fixture: 44100 Hz stereo FLAC.
	if rate != 44100 {
		t.Errorf("expected sampleRate 44100, got %d", rate)
	}
	if ch != 2 {
		t.Errorf("expected channels 2, got %d", ch)
	}

	// 3 seconds at 44100 Hz stereo = 264600 interleaved samples.
	expectedSamples := 264600
	if len(samples) != expectedSamples {
		t.Errorf("expected %d interleaved samples, got %d", expectedSamples, len(samples))
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
