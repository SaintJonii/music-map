# Proposal: Audio Feature Extraction

## Intent
Build the acoustic feature extraction pipeline for Mapa Musical — the first Go module in a day-zero TDD project. Replace Python stubs with idiomatic Go, extract BPM, key, energy, spectral centroid, chroma, MFCCs, and ZCR from WAV/MP3/FLAC files, enrich with MusicBrainz metadata, and persist results.

## Scope

### In Scope
- Go module init + directory layout (`model/`, `audio/`, `metadata/`, `storage/`)
- Unified `Decoder` interface: WAV, MP3, FLAC adapters via go-audio, go-mp3, mewkiz/flac
- 7 feature extractors from go-dsp FFT primitives: RMS energy, ZCR, spectral centroid → BPM, chroma, MFCCs, key
- Metadata pipeline: dhowden/tag (local tags) + go-musicbrainzws2 (enrichment)
- Domain types: `Track`, `TrackFeatures`, `Collection`
- SQLite persistence for extracted features
- Python artifact cleanup
- CI baseline: `go test -cover`, `golangci-lint`, ≥80% coverage

### Out of Scope
- Interactive 2D visualization (post-MVP), acoustic similarity clustering, streaming, HTTP API

## Capabilities

### New Capabilities
- `audio-decoding`: Decode WAV/MP3/FLAC to PCM float64 via `Decoder` interface with format adapters
- `feature-extraction`: Extract BPM, key, energy, spectral centroid, chroma, MFCCs, ZCR from PCM samples using go-dsp FFT
- `metadata-enrichment`: Read local tags (dhowden/tag) + query MusicBrainz API when MBIDs present
- `domain-model`: `Track`, `TrackFeatures`, `Collection` structs with JSON serialization
- `persistence`: SQLite schema + repository for tracks and features
- `project-scaffold`: Go module, directory layout, CI/lint config, Python cleanup

### Modified Capabilities
None — `openspec/specs/` is empty.

## Approach
Pure Go, TDD from go-dsp primitives. Extractors built test-first against synthetic sine-wave signals. Decoders tested against fixture files in `testdata/`. Metadata tested with recorded HTTP mocks. Interfaces at every boundary. Flat package layout (`audio/`, `metadata/`, `model/`, `storage/`) per exploration recommendation.

## Affected Areas
| Area | Impact | Description |
|------|--------|-------------|
| `/` (root) | New | `go.mod`, `.golangci.yml`, CI config |
| `model/` | New | Domain types |
| `audio/` | New | Decoder adapters, feature extractors |
| `metadata/` | New | Tag reader, MusicBrainz enricher |
| `storage/` | New | SQLite repository |
| `src/`, `tests/`, `*requirements*.txt`, `pytest.ini` | Removed | Python cleanup |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| BPM/key accuracy below librosa parity | Med | Acceptable for TDD lab; validate with synthetic signals |
| dhowden/tag lacks tagged Go version | Low | Pseudo-version OK; 351 importers |
| go-musicbrainzws2 API rate limits | Low | Cached responses, throttled client |
| go-audio/audio v1.0.0 is old (2018) | Low | Buffer abstractions stable; decoders fresher |

## Rollback Plan
Delete Go module, restore Python artifacts from git. Greenfield — no existing Go code to break.

## Dependencies
- Go 1.26.5+ | go-audio/wav, go-mp3, mewkiz/flac, go-dsp, dhowden/tag, go-musicbrainzws2, modernc.org/sqlite

## Success Criteria
- [ ] `go test ./... -cover` ≥80% coverage across all packages
- [ ] All 7 extractors within 5% of reference values for synthetic test signals
- [ ] WAV/MP3/FLAC fixtures decode to bit-identical PCM
- [ ] MusicBrainz enrichment resolves via recorded HTTP fixtures
- [ ] `go build ./...` succeeds with zero warnings
- [ ] `golangci-lint run` passes with zero issues
- [ ] Python artifacts removed from repo
