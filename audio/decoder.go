package audio

import (
	"errors"
	"fmt"
	"io"
)

// ErrUnsupportedFormat is returned when the audio format cannot be decoded.
var ErrUnsupportedFormat = errors.New("unsupported audio format")

// ErrEmptyInput is returned when the input stream has no data.
var ErrEmptyInput = errors.New("audio: empty input")

// Decoder decodes an audio stream into normalized PCM float64 samples.
// Samples are normalized to the range [-1.0, 1.0].
type Decoder interface {
	Decode(r io.Reader) (samples []float64, sampleRate int, channels int, err error)
}

// DetectFormat inspects the beginning of an audio stream and returns the
// appropriate Decoder for the detected format. It peeks at the stream's
// magic bytes to determine the format.
func DetectFormat(r io.Reader) (Decoder, error) {
	// Peek at first 12 bytes to detect format.
	buf := make([]byte, 12)
	n, err := io.ReadFull(r, buf)
	if n == 0 {
		return nil, ErrEmptyInput
	}
	if err != nil {
		return nil, fmt.Errorf("audio: failed to read format header: %w", err)
	}

	// WAV detection: "RIFF" at offset 0, "WAVE" at offset 8.
	if buf[0] == 'R' && buf[1] == 'I' && buf[2] == 'F' && buf[3] == 'F' &&
		buf[8] == 'W' && buf[9] == 'A' && buf[10] == 'V' && buf[11] == 'E' {
		return &WAVDecoder{}, nil
	}

	// MP3 detection: sync bits 0xFF followed by 0xE0-0xFF.
	if buf[0] == 0xFF && (buf[1]&0xE0) == 0xE0 {
		return &MP3Decoder{}, nil
	}

	// FLAC detection: "fLaC" marker.
	if buf[0] == 'f' && buf[1] == 'L' && buf[2] == 'a' && buf[3] == 'C' {
		return &FLACDecoder{}, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, detectTypeName(buf))
}

// detectTypeName returns a human-readable format name from magic bytes.
func detectTypeName(buf []byte) string {
	if len(buf) >= 4 && buf[0] == 'O' && buf[1] == 'g' && buf[2] == 'g' && buf[3] == 'S' {
		return ".ogg"
	}
	return "unknown"
}
