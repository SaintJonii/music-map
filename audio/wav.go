package audio

import (
	"bytes"
	"fmt"
	"io"
	"math"

	wavlib "github.com/go-audio/wav"
)

// WAVDecoder decodes WAV audio using the go-audio/wav library.
// Supports 16-bit, 24-bit, and 32-bit integer PCM. 8-bit WAV returns ErrUnsupportedFormat.
type WAVDecoder struct{}

// Decode implements the Decoder interface for WAV files.
// Samples are normalized to float64 in the range [-1.0, 1.0].
func (d *WAVDecoder) Decode(r io.Reader) ([]float64, int, int, error) {
	// Prefer direct seekable reader to avoid buffering large files in memory.
	// Fall back to io.ReadAll only when the reader is not seekable.
	var seeker io.ReadSeeker
	if rs, ok := r.(io.ReadSeeker); ok {
		seeker = rs
	} else {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("wav: failed to read input: %w", err)
		}
		seeker = bytes.NewReader(data)
	}

	dec := wavlib.NewDecoder(seeker)
	if !dec.IsValidFile() {
		return nil, 0, 0, fmt.Errorf("wav: invalid file: %w", dec.Err())
	}

	// Read headers to get format info.
	dec.ReadInfo()
	if err := dec.Err(); err != nil {
		return nil, 0, 0, fmt.Errorf("wav: failed to read headers: %w", err)
	}

	// Reject unsupported bit depths (only 8-bit is unsupported for now).
	bitDepth := int(dec.BitDepth)
	if bitDepth == 8 {
		return nil, 0, 0, fmt.Errorf("%w: 8-bit WAV is not supported", ErrUnsupportedFormat)
	}

	buf, err := dec.FullPCMBuffer()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("wav: failed to decode PCM: %w", err)
	}

	samples := intSamplesToFloat64(buf.Data, bitDepth)
	sampleRate := int(dec.SampleRate)
	channels := int(dec.NumChans)

	return samples, sampleRate, channels, nil
}

// intSamplesToFloat64 converts raw integer PCM samples to float64 in [-1.0, 1.0].
func intSamplesToFloat64(data []int, bitDepth int) []float64 {
	if len(data) == 0 {
		return nil
	}

	// Normalize by 2^(bitDepth-1) — the standard audio convention.
	// Use int64 to avoid overflow at 32-bit depth.
	norm := 1.0 / float64(int64(1)<<uint(bitDepth-1))

	samples := make([]float64, len(data))
	for i, v := range data {
		samples[i] = math.Max(-1.0, math.Min(1.0, float64(v)*norm))
	}
	return samples
}
