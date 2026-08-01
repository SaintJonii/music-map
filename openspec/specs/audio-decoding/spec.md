# audio-decoding Specification

## Purpose

Define the unified `Decoder` interface and format-specific adapters for WAV, MP3, and FLAC files, producing normalized PCM float64 samples.

## Requirements

### Requirement: Decoder Interface

The system MUST define a `Decoder` interface: `Decode(r io.Reader) (samples []float64, sampleRate int, channels int, err error)`.

#### Scenario: Interface contracts are documented

- GIVEN the `Decoder` interface definition
- WHEN reading its godoc
- THEN each return value's contract is described (samples normalized to [-1.0, 1.0], sampleRate in Hz, channels ≥ 1)

#### Scenario: Interface is mockable in tests

- GIVEN a test that needs a fake Decoder
- WHEN a mock implements the `Decoder` interface
- THEN `go vet` confirms interface satisfaction without compilation errors

### Requirement: WAV Decoder

The WAV adapter MUST decode 16-bit, 24-bit, and 32-bit integer PCM WAV files into float64 samples normalized to [-1.0, 1.0].

#### Scenario: 16-bit stereo WAV decodes correctly

- GIVEN a 44100 Hz 16-bit stereo WAV fixture
- WHEN `Decode` is called with the file's reader
- THEN samples are within [-1.0, 1.0]
- AND `sampleRate` is 44100
- AND `channels` is 2

#### Scenario: Unsupported bit depth returns error

- GIVEN an 8-bit WAV file
- WHEN `Decode` is called
- THEN an error is returned with message indicating unsupported bit depth

### Requirement: MP3 Decoder

The MP3 adapter MUST decode MPEG Audio Layer III files into mono or stereo PCM float64.

#### Scenario: 128 kbps MP3 decodes to expected sample count

- GIVEN a known MP3 fixture with 132300 samples (3s × 44100)
- WHEN `Decode` is called
- THEN the decoded sample count is within 1% of expected

#### Scenario: Corrupt MP3 returns decode error

- GIVEN a truncated MP3 file
- WHEN `Decode` is called
- THEN a non-nil error is returned
- AND the error message indicates corrupted or incomplete stream

### Requirement: FLAC Decoder

The FLAC adapter MUST decode FLAC files to PCM float64.

#### Scenario: 16-bit FLAC decodes without loss

- GIVEN a lossless FLAC fixture
- WHEN `Decode` is called
- THEN samples fall within [-1.0, 1.0]
- AND `sampleRate` matches the file header

### Requirement: Format Detection

The system MUST detect audio format from file extension or magic bytes and SHALL return a descriptive error for unsupported formats.

#### Scenario: OGG file returns "unsupported format"

- GIVEN a valid OGG Vorbis file
- WHEN format detection runs
- THEN a `"unsupported format: .ogg"` error is returned

#### Scenario: Empty file returns decode error

- GIVEN an empty (zero-byte) file
- WHEN `Decode` is called
- THEN an error is returned with a message indicating empty input
