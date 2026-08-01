# Design: Audio Feature Extraction Pipeline

## Technical Approach

Pure Go pipeline, TDD from go-dsp primitives. Flat package layout — `model/` has zero imports, `audio/`, `metadata/`, `storage/` each implement interfaces at their boundary. **Tags-first data source priority**: local ID3v2/Vorbis tags → (future Navidrome/Subsonic API) → acoustic analysis fallback → MusicBrainz enrichment via ISRC. Sequential processing: read tags → decode only if features missing → extract features → enrich → persist. Every stage independently testable via interface mocks. A `LibrarySource` interface abstracts the music source (filesystem today, Navidrome post-MVP).

Post-MVP: PCA/UMAP dimensionality reduction collapses feature vectors into 2D coordinates for the interactive scatter-plot map.

## Architecture Decisions

| Decision | Options | Tradeoffs | Choice |
|---|---|---|---|
| Project layout | A) `cmd/`+`internal/` vs B) flat packages | A scales with modules but over-engineers day-zero; B simpler, refactorable later | **B — flat** (`model/`, `audio/`, `metadata/`, `storage/`) |
| Package boundaries | 1 shared `internal/` vs domain-driven packages | Domain packages enforce separation without `internal/` visibility restrictions | **4 packages**: `model`, `audio`, `metadata`, `storage` |
| Concurrency model | Sequential vs fan-out goroutine pipeline | Sequential is simpler to test/trace; concurrency adds goroutine-leak risk without measurably speeding up single-file processing | **Sequential** — single goroutine per file |
| Error handling | Sentinel errors + `%w` vs custom error types vs error values | Sentinel errors (`ErrNotFound`, `ErrUnsupportedFormat`) enable `errors.Is`; custom types added only when callers need structured fields | **Sentinel errors** with `fmt.Errorf("%w")` wrapping |
| DI approach | Constructor injection vs functional options vs global registry | Constructor injection is explicit, zero magic, testable; functional options overkill for ≤5 deps | **Constructor injection** — each package exposes `NewX(deps...) *X` |
| Data source priority | Tags-first vs audio-first | Tags are O(1) per file vs O(n) per second of audio; ISRC enables precise MusicBrainz matching | **Tags-first**: read tags → decode only when features are missing → MusicBrainz via ISRC |
| Library source | Direct filesystem vs Navidrome/Subsonic API | Filesystem is simpler for MVP; API enables remote access + pre-built index post-MVP | **Filesystem now, `LibrarySource` interface designed for swap** |
| Test fixtures | Synthetic generation vs real audio files in `testdata/` | Synthetic signals test extractors with known math; real fixtures validate decoder integration | **Both**: sine/tone generators for extractors, real WAV/MP3/FLAC in `testdata/` for decoders |
| MB HTTP testing | httptest.Server mocks vs recorded golden responses | Golden responses (stored MB API replies) exercise unmarshalling; httptest verifies request shape | **Recorded golden responses** for `metadata/` integration tests |

## Data Flow

```
AudioFile ──→ TagReader ──→ Metadata (Title, Artist, ISRC, Genre, Year...)
     │              │
     │              └──→ MusicBrainzClient (enrich via ISRC/MBID)
     │                                         │
     │   (only if features missing)            ▼
     └──→ Decoder ──→ []float64 PCM        Enriched
                         │                    │
                         ▼                    │
                   FeatureExtractor           │
                         │                    │
                         ▼                    ▼
                   TrackFeatures    Repository.Save(Track + Features)
                                                  │
                                                  ▼
                                              SQLite
```

## File Changes

| File | Action | Description |
|---|---|---|
| `go.mod` | Create | Module `github.com/[user]/mapa-musical-tdd`, Go 1.26 |
| `model/track.go` | Create | `Track`, `TrackFeatures`, `Collection` structs + JSON tags |
| `audio/decoder.go` | Create | `Decoder` interface + `DetectFormat(r io.Reader) (Decoder, error)` factory |
| `audio/wav.go` | Create | WAV adapter via go-audio/wav |
| `audio/mp3.go` | Create | MP3 adapter via hajimehoshi/go-mp3 |
| `audio/flac.go` | Create | FLAC adapter via mewkiz/flac |
| `audio/features.go` | Create | `FeatureExtractor` interface, RMS, ZCR, spectral centroid, BPM, chroma, MFCCs, key |
| `metadata/reader.go` | Create | `TagReader` interface, dhowden/tag wrapper |
| `metadata/enricher.go` | Create | `MusicBrainzClient` interface, go-musicbrainzws2 wrapper |
| `storage/repository.go` | Create | `Repository` interface with `Save`, `GetByID`, `List`; modernc.org/sqlite |
| `cmd/mapa-musical/main.go` | Create | CLI entry point wiring packages together |
| `.golangci.yml` | Create | Linter config: errcheck, gosec, govet, staticcheck |
| `.gitignore` | Modify | Add Go patterns (`*.exe`, `*.test`, `coverage.out`, `vendor/`) |
| `README.md` | Modify | Replace Python stack with Go, update test commands |
| `src/main.py` | Delete | Python stub removal |
| `tests/test_main.py` | Delete | Python test removal |
| `tests/conftest.py` | Delete | Python fixture removal |
| `pytest.ini` | Delete | Python config removal |
| `requirements.txt` | Delete | Python dep removal |
| `requirements-dev.txt` | Delete | Python dev dep removal |

## Interfaces / Contracts

```go
// audio/decoder.go
type Decoder interface {
    Decode(r io.Reader) (samples []float64, sampleRate int, channels int, err error)
}

// audio/features.go
type FeatureExtractor interface {
    ExtractRMS(samples []float64) float64
    ExtractZCR(samples []float64, sampleRate int) float64
    ExtractSpectralCentroid(samples []float64, sampleRate int) (float64, error)
    ExtractBPM(samples []float64, sampleRate int) (float64, error)
    ExtractChroma(samples []float64, sampleRate int) ([12]float64, error)
    ExtractMFCCs(frame []float64) ([13]float64, error)
    ExtractKey(chroma [12]float64) string
    ExtractDanceability(bpm, energy float64, zcr float64) float64
    ExtractAcousticness(spectralCentroid float64, energy float64) float64
}

// metadata/reader.go
type TagReader interface {
    ReadTags(r io.ReadSeeker) (model.Track, error)
}

// metadata/enricher.go
type MusicBrainzClient interface {
    LookupByISRC(ctx context.Context, isrc string) (model.Track, error)
    LookupByMBID(ctx context.Context, mbid string) (model.Track, error)
}

// metadata/source.go — abstract music library source (filesystem now, Navidrome post-MVP)
type LibrarySource interface {
    ListTracks(ctx context.Context) ([]string, error)   // returns file paths or stream URLs
    OpenTrack(ctx context.Context, id string) (io.ReadCloser, error)
}

// storage/repository.go
type Repository interface {
    Save(ctx context.Context, track model.Track, features model.TrackFeatures) error
    GetByID(ctx context.Context, id string) (model.Track, model.TrackFeatures, error)
    List(ctx context.Context) ([]model.Track, error)
}
```

Sentinel errors: `audio.ErrUnsupportedFormat`, `metadata.ErrNotFound`, `storage.ErrNotFound`, `storage.ErrConflict`.

## Testing Strategy

| Layer | What | Approach |
|---|---|---|
| Unit — `model/` | Struct marshalling, defaults | Table-driven JSON round-trip tests, `go-cmp` for deep equality |
| Unit — `audio/features` | RMS, ZCR, spectral centroid, chroma, MFCCs, BPM, key | Table-driven with synthetic sine/clicks/white-noise. Assert within tolerance (±5% for BPM, ±0.01 for energy) |
| Integration — `audio/` decoders | WAV/MP3/FLAC decode | Fixture files in `testdata/` (synthetic tones encoded to each format). Assert sample count, rate, range |
| Integration — `metadata/` | Tag reading, MB enrichment | Real tagged fixtures + recorded golden HTTP responses (httptest or recorded JSON blobs) |
| Unit — `storage/` | CRUD, schema migration, error paths | In-memory SQLite (`:memory:`), table-driven save+retrieve+conflict scenarios |
| E2E | Full pipeline | `cmd/` integration test: decode+extract+enrich+save a real track fixture, verify DB row |

Coverage target: ≥80% per `openspec/config.yaml`. TDD enforced — RED tests written first, then GREEN implementation.

## Threat Matrix

N/A — no routing, shell, subprocess, VCS/PR automation, executable-file classification, or process-integration boundary. This is a pure audio processing pipeline with file I/O and HTTP API calls.

## Migration / Rollout

No migration required. Day-zero project with only Python stubs to delete. Go module is greenfield.

## Open Questions

None.
