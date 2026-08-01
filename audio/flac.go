package audio

import (
	"fmt"
	"io"
	"math"

	flaclib "github.com/mewkiz/flac"
)

// FLACDecoder decodes FLAC audio using the mewkiz/flac library.
// Output is normalized to float64 [-1.0, 1.0].
type FLACDecoder struct{}

// Decode implements the Decoder interface for FLAC files.
func (d *FLACDecoder) Decode(r io.Reader) ([]float64, int, int, error) {
	stream, err := flaclib.New(r)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("flac: failed to parse stream: %w", err)
	}

	info := stream.Info
	sampleRate := int(info.SampleRate)
	channels := int(info.NChannels)
	bitDepth := int(info.BitsPerSample)

	var allSamples []float64
	for {
		f, err := stream.ParseNext()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, 0, fmt.Errorf("flac: decode error: %w", err)
		}

		// Interleave samples from all subframes into output.
		subframes := f.Subframes
		if len(subframes) == 0 {
			continue
		}
		nsamples := len(subframes[0].Samples)
		for i := 0; i < nsamples; i++ {
			for ch := 0; ch < channels; ch++ {
				if ch < len(subframes) && i < len(subframes[ch].Samples) {
					allSamples = append(allSamples, int32ToFloat64(subframes[ch].Samples[i], bitDepth))
				}
			}
		}
	}

	if len(allSamples) == 0 {
		return nil, 0, 0, fmt.Errorf("flac: no audio samples decoded")
	}

	return allSamples, sampleRate, channels, nil
}

// int32ToFloat64 converts a single int32 PCM sample to float64 in [-1.0, 1.0].
func int32ToFloat64(sample int32, bitDepth int) float64 {
	maxVal := float64(int64(1) << uint(bitDepth-1))
	return math.Max(-1.0, math.Min(1.0, float64(sample)/maxVal))
}
