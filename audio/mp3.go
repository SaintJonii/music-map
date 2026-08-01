package audio

import (
	"encoding/binary"
	"fmt"
	"io"

	mp3lib "github.com/hajimehoshi/go-mp3"
)

// MP3Decoder decodes MPEG Audio Layer III files.
// Output is always 16-bit stereo PCM, normalized to float64 [-1.0, 1.0].
type MP3Decoder struct{}

// Decode implements the Decoder interface for MP3 files.
func (d *MP3Decoder) Decode(r io.Reader) ([]float64, int, int, error) {
	dec, err := mp3lib.NewDecoder(r)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("mp3: failed to create decoder: %w", err)
	}

	sampleRate := dec.SampleRate()

	// Read all PCM data. go-mp3 always outputs 16-bit stereo PCM.
	var pcmData []byte
	buf := make([]byte, 4096)
	for {
		n, readErr := dec.Read(buf)
		if n > 0 {
			pcmData = append(pcmData, buf[:n]...)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, 0, 0, fmt.Errorf("mp3: decode error: %w", readErr)
		}
	}

	if len(pcmData) == 0 {
		return nil, 0, 0, fmt.Errorf("mp3: no PCM data decoded")
	}

	// Each sample is 2 bytes (int16 LE). Interleaved stereo: L0,R0,L1,R1,...
	samples := pcmBytesToFloat64(pcmData)

	// go-mp3 always outputs stereo (2 channels), per its documentation.
	channels := 2

	return samples, sampleRate, channels, nil
}

// pcmBytesToFloat64 converts 16-bit little-endian PCM bytes to float64 [-1.0, 1.0].
func pcmBytesToFloat64(data []byte) []float64 {
	count := len(data) / 2
	if count == 0 {
		return nil
	}
	samples := make([]float64, count)
	for i := range count {
		val := int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
		samples[i] = float64(val) / 32768.0
	}
	return samples
}
