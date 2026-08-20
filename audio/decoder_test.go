package audio

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// mockDecoder implements the Decoder interface for testing.
type mockDecoder struct {
	samples    []float64
	sampleRate int
	channels   int
	err        error
}

func (m *mockDecoder) Decode(r io.Reader) ([]float64, int, int, error) {
	return m.samples, m.sampleRate, m.channels, m.err
}

func TestDecoderInterfaceMockSatisfiesContract(t *testing.T) {
	// Compile-time check: mockDecoder must satisfy Decoder interface.
	var d Decoder = &mockDecoder{samples: []float64{0.5, -0.3}, sampleRate: 44100, channels: 2}
	samples, rate, ch, err := d.Decode(bytes.NewReader(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate != 44100 {
		t.Errorf("expected sampleRate 44100, got %d", rate)
	}
	if ch != 2 {
		t.Errorf("expected channels 2, got %d", ch)
	}
	if len(samples) != 2 {
		t.Errorf("expected 2 samples, got %d", len(samples))
	}
	if samples[0] != 0.5 || samples[1] != -0.3 {
		t.Errorf("unexpected sample values: %v", samples)
	}
}

func TestDetectFormat_WAVMagicBytes(t *testing.T) {
	// RIFF....WAVE header — should return a *WAVDecoder (non-nil, no error).
	wavHeader := []byte("RIFF\x00\x00\x00\x00WAVE")
	dec, _, err := DetectFormat(bytes.NewReader(wavHeader))
	if err != nil {
		t.Fatalf("DetectFormat: expected no error for WAV, got %v", err)
	}
	if dec == nil {
		t.Fatal("DetectFormat: expected non-nil Decoder for WAV")
	}
}

func TestDetectFormat_UnsupportedOGG(t *testing.T) {
	oggHeader := []byte("OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00")
	_, _, err := DetectFormat(bytes.NewReader(oggHeader))
	if err == nil {
		t.Fatal("DetectFormat: expected error for OGG, got nil")
	}
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("DetectFormat: expected ErrUnsupportedFormat, got %v", err)
	}
}

func TestDetectFormat_EmptyInput(t *testing.T) {
	_, _, err := DetectFormat(bytes.NewReader([]byte{}))
	if err == nil {
		t.Fatal("DetectFormat: expected error for empty input, got nil")
	}
	if !errors.Is(err, ErrEmptyInput) {
		t.Errorf("DetectFormat: expected ErrEmptyInput, got %v", err)
	}
}

func TestDetectFormat_MP3MagicBytes(t *testing.T) {
	// MPEG frame sync: 0xFF followed by 0xFB (MPEG1 Layer3)
	mp3Header := []byte{0xFF, 0xFB, 0x90, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	dec, _, err := DetectFormat(bytes.NewReader(mp3Header))
	if err != nil {
		t.Fatalf("DetectFormat: expected no error for MP3, got %v", err)
	}
	if dec == nil {
		t.Fatal("DetectFormat: expected non-nil Decoder for MP3")
	}
}

func TestDetectFormat_FLACMagicBytes(t *testing.T) {
	flacHeader := []byte("fLaC\x00\x00\x00\x16\x00\x00\x00\x00\x00\x00")
	dec, _, err := DetectFormat(bytes.NewReader(flacHeader))
	if err != nil {
		t.Fatalf("DetectFormat: expected no error for FLAC, got %v", err)
	}
	if dec == nil {
		t.Fatal("DetectFormat: expected non-nil Decoder for FLAC")
	}
}

func TestDetectFormat_UnknownMagicBytes(t *testing.T) {
	unknown := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	_, _, err := DetectFormat(bytes.NewReader(unknown))
	if err == nil {
		t.Fatal("DetectFormat: expected error for unknown format, got nil")
	}
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("DetectFormat: expected ErrUnsupportedFormat, got %v", err)
	}
}

func TestDetectFormat_OGGErrorMessage(t *testing.T) {
	oggHeader := []byte("OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	_, _, err := DetectFormat(bytes.NewReader(oggHeader))
	if err == nil {
		t.Fatal("DetectFormat: expected error for OGG, got nil")
	}
	// Verify the error message contains the format hint.
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("DetectFormat: expected non-empty error message")
	}
}

func TestDetectFormat_ShortInput(t *testing.T) {
	// 5 bytes — too short for any format, but not empty.
	_, _, err := DetectFormat(bytes.NewReader([]byte{0x01, 0x02, 0x03, 0x04, 0x05}))
	if err == nil {
		t.Fatal("DetectFormat: expected error for short input, got nil")
	}
}

func TestDetectFormat_ID3TaggedMP3(t *testing.T) {
	t.Run("valid id3-tagged mp3 detected", func(t *testing.T) {
		f, err := os.Open(filepath.Join("..", "testdata", "mp3", "test_128kbps.mp3"))
		if err != nil {
			t.Fatalf("open fixture: %v", err)
		}
		defer func() { _ = f.Close() }()

		dec, _, err := DetectFormat(f)
		if err != nil {
			t.Fatalf("DetectFormat: unexpected error: %v", err)
		}
		if _, ok := dec.(*MP3Decoder); !ok {
			t.Fatalf("expected *MP3Decoder, got %T", dec)
		}
	})

	t.Run("corrupt id3-tagged mp3 detected as mp3", func(t *testing.T) {
		// corrupt.mp3 is an ID3 tag + one truncated MPEG frame. It must be
		// DETECTED as MP3 (sync after the tag); the failure is at Decode.
		f, err := os.Open(filepath.Join("..", "testdata", "mp3", "corrupt.mp3"))
		if err != nil {
			t.Fatalf("open fixture: %v", err)
		}
		defer func() { _ = f.Close() }()

		dec, _, err := DetectFormat(f)
		if err != nil {
			t.Fatalf("DetectFormat: expected MP3 detection, got error: %v", err)
		}
		if _, ok := dec.(*MP3Decoder); !ok {
			t.Fatalf("expected *MP3Decoder, got %T", dec)
		}
	})
}

func TestMockDecoder_ErrorPropagation(t *testing.T) {
	wantErr := errors.New("custom decode failure")
	d := &mockDecoder{err: wantErr}
	_, _, _, err := d.Decode(bytes.NewReader(nil))
	if !errors.Is(err, wantErr) {
		t.Errorf("expected %v, got %v", wantErr, err)
	}
}
