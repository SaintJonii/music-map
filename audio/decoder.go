package audio

import (
	"bytes"
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
// appropriate Decoder for the detected format and a reader that includes the
// consumed magic bytes. The returned reader is safe to pass directly to
// Decoder.Decode without seeking.
func DetectFormat(r io.Reader) (Decoder, io.Reader, error) {
	// Peek at first 12 bytes to detect format.
	buf := make([]byte, 12)
	n, err := io.ReadFull(r, buf)
	if n == 0 {
		return nil, nil, ErrEmptyInput
	}
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, nil, fmt.Errorf("audio: failed to read format header: %w", err)
	}
	if n < 4 && err == io.ErrUnexpectedEOF {
		return nil, nil, ErrEmptyInput
	}

	// Prepend consumed bytes so the reader is intact for the decoder.
	restored := io.MultiReader(bytes.NewReader(buf[:n]), r)

	// WAV detection: "RIFF" at offset 0, "WAVE" at offset 8.
	if buf[0] == 'R' && buf[1] == 'I' && buf[2] == 'F' && buf[3] == 'F' &&
		buf[8] == 'W' && buf[9] == 'A' && buf[10] == 'V' && buf[11] == 'E' {
		return &WAVDecoder{}, restored, nil
	}

	// MP3 detection: sync bits 0xFF followed by 0xE0-0xFF.
	if buf[0] == 0xFF && (buf[1]&0xE0) == 0xE0 {
		return &MP3Decoder{}, restored, nil
	}

	// FLAC detection: "fLaC" marker.
	if buf[0] == 'f' && buf[1] == 'L' && buf[2] == 'a' && buf[3] == 'C' {
		return &FLACDecoder{}, restored, nil
	}

	return nil, nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, detectTypeName(buf))
}

// detectTypeName returns a human-readable format name from magic bytes.
func detectTypeName(buf []byte) string {
	if len(buf) >= 4 && buf[0] == 'O' && buf[1] == 'g' && buf[2] == 'g' && buf[3] == 'S' {
		return ".ogg"
	}
	return "unknown"
}
