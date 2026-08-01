# Tasks: Audio Feature Extraction

## Review Workload Forecast

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

~1200+ lines across 6 packages, 7 extractors, 3 decoders, SQLite, tests, Python cleanup.

### Suggested Work Units

| Unit | PR | Focused test | Rollback |
|------|----|-------------|----------|
| 1: Scaffold + domain model | PR 1 | `go test ./model/... -v` | Delete go.mod, restore Python artifacts |
| 2: Decoder + WAV/MP3/FLAC | PR 2 | `go test ./audio/... -v -run Decode` | Delete audio/*.go (except features.go) |
| 3: 7 extractors (RMS→Key) | PR 3 | `go test ./audio/... -v -run Feature` | Delete audio/features.go + test file |
| 4: Metadata + Persistence + CLI + E2E | PR 4 | `go test ./metadata/... ./storage/... -v` | Delete metadata/, storage/, cmd/ |

Runtime harness: `go build ./...` (all units). E2E: WAV fixture pipeline → verify DB.

## Phase 1: Foundation (Project Scaffold + Domain Model)

- [x] 1.1 Delete Python artifacts: `src/`, `tests/`, `pytest.ini`, `requirements*.txt`
- [x] 1.2 Init Go module (`go mod init` at Go 1.26), create `.golangci.yml`, update `.gitignore` (Go patterns), update `README.md` (Go stack)
- [x] 1.3 Create directories: `model/`, `audio/`, `metadata/`, `storage/`, `testdata/`
- [x] 1.4 RED → GREEN → REFACTOR: `model/track.go` — Track, TrackFeatures, Collection + JSON tags. Table-driven round-trip, defaults, empty-collection tests (4 reqs, 9 scenarios)

## Phase 2: Audio Decoding (Decoder Interface + Format Adapters)

- [x] 2.1 RED → GREEN → REFACTOR: `audio/decoder.go` — Decoder interface + DetectFormat. Mock test, `ErrUnsupportedFormat`, empty-input error (interface + format detection reqs)
- [x] 2.2 RED → GREEN: `audio/wav.go` — 16/24/32-bit WAV → float64 [-1.0, 1.0]; fixtures: 44100 Hz stereo, unsupported 8-bit error
- [x] 2.3 RED → GREEN: `audio/mp3.go` — MPEG Layer III decode; fixtures: 128 kbps sample count ±2%, corrupt stream error
- [x] 2.4 RED → GREEN: `audio/flac.go` — FLAC → PCM float64; fixture: 16-bit lossless verify

## Phase 3: Feature Extraction (go-dsp FFT Primitives)

- [ ] 3.1 RED → GREEN: RMS (sine 0.5→0.5±0.01, silence→0.0), ZCR (440 Hz→880±5%, DC-offset→0), SpectralCentroid (100 Hz→100±10 Hz)
- [ ] 3.2 RED → GREEN: BPM (120 BPM click ±5%, short→warn), Chroma (C-major→bins C/E/G highest), MFCCs (1024-frame→13 coeffs, empty→zeros), Key (A-minor→"A minor")
- [ ] 3.3 REFACTOR: extract shared FFT helpers, window functions, normalization constants

## Phase 4: Metadata + Persistence

- [ ] 4.1 RED → GREEN: `metadata/reader.go` — TagReader (dhowden/tag). Tests: MP3 ID3v2, FLAC Vorbis, untagged WAV→empty, missing file→error
- [ ] 4.2 RED → GREEN: `metadata/enricher.go` — MusicBrainzClient (go-musicbrainzws2) with golden HTTP fixtures. Tests: valid MBID, unknown→not-found, timeout/503→retryable-error
- [ ] 4.3 RED → GREEN: `storage/repository.go` — SQLite (modernc.org/sqlite), auto-migrate tracks+features tables with FK. Save+GetByID+List via `:memory:`, sentinel `ErrNotFound`/`ErrConflict`, corrupt-DB no-panic (5 reqs)

## Phase 5: CLI Integration + E2E + Coverage

- [ ] 5.1 RED → GREEN: `cmd/mapa-musical/main.go` — wire pipeline: decode → extract → enrich → persist
- [ ] 5.2 E2E: full pipeline on short WAV fixture in `testdata/` → verify SQLite row matches extracted features
- [ ] 5.3 Verify: `go test ./... -cover` ≥80%, `golangci-lint run` clean, `go build ./...` passes
