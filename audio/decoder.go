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

	// ID3v2-tagged files: the MPEG frame sync lives after the tag. go-mp3
	// skips the tag itself, so we only need to skip it far enough to confirm
	// the sync, then hand back the intact stream.
	if n >= 10 && buf[0] == 'I' && buf[1] == 'D' && buf[2] == '3' {
		return detectID3v2(buf[:n], r)
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

// id3v2TagSize returns the total byte size of an ID3v2 tag (10-byte header +
// body + optional 10-byte footer) encoded in the header, or 0 if the header is
// not a valid ID3v2 header.
func id3v2TagSize(header []byte) int {
	if len(header) < 10 || header[0] != 'I' || header[1] != 'D' || header[2] != '3' {
		return 0
	}
	// Bytes 6-9 are the tag body size, syncsafe (7 bits per byte, MSB ignored).
	size := int(header[6]&0x7F)<<21 |
		int(header[7]&0x7F)<<14 |
		int(header[8]&0x7F)<<7 |
		int(header[9]&0x7F)
	total := 10 + size
	// ID3v2.4 flag bit 4 (0x10) marks a 10-byte footer after the body.
	if header[5]&0x10 != 0 {
		total += 10
	}
	return total
}

// detectID3v2 handles files that begin with an ID3v2 tag. It skips the tag,
// verifies the MPEG frame sync immediately after it, and returns an MP3Decoder
// together with the intact stream (every consumed byte restored via
// io.MultiReader) so go-mp3 can re-skip the tag itself. A missing or
// non-MPEG sync after the tag is ErrUnsupportedFormat.
func detectID3v2(header []byte, r io.Reader) (Decoder, io.Reader, error) {
	tagSize := id3v2TagSize(header)
	if tagSize <= 0 {
		return nil, nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, detectTypeName(header))
	}

	// Read the rest of the tag plus two bytes of the audio frame to verify
	// MPEG sync. need = header already read + remainder + 2 sync bytes.
	need := tagSize + 2
	var rest []byte
	if need > len(header) {
		extra := make([]byte, need-len(header))
		m, err := io.ReadFull(r, extra)
		if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
			return nil, nil, fmt.Errorf("audio: failed to read id3 tag: %w", err)
		}
		rest = append(header, extra[:m]...)
	} else {
		rest = header[:need]
	}

	// Verify MPEG frame sync immediately after the tag.
	syncOff := tagSize
	if syncOff+1 >= len(rest) {
		return nil, nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, detectTypeName(header))
	}
	if rest[syncOff] != 0xFF || (rest[syncOff+1]&0xE0) != 0xE0 {
		return nil, nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, detectTypeName(header))
	}

	// Hand back the intact stream; go-mp3 re-skips the ID3v2 tag on Decode.
	restored := io.MultiReader(bytes.NewReader(rest), r)
	return &MP3Decoder{}, restored, nil
}

// detectTypeName returns a human-readable format name from magic bytes.
func detectTypeName(buf []byte) string {
	if len(buf) >= 4 && buf[0] == 'O' && buf[1] == 'g' && buf[2] == 'g' && buf[3] == 'S' {
		return ".ogg"
	}
	return "unknown"
}
