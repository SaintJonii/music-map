# Exploration: Go Audio Ecosystem for Mapa Musical TDD

> **Date**: 2026-07-30
> **Topic Slug**: `go-audio-ecosystem`
> **Standalone**: Yes — not tied to a named SDD change

## Current State

The project is at **day zero**. No Go code exists. The repo contains:
- Python source stubs (`src/mapa_musical/__init__.py`)
- Python test scaffolding (`tests/`)
- `openspec/config.yaml` (SDD config, `tdd: true`, coverage threshold 80%)
- Go 1.26.5 is installed on the system
- No `go.mod` file exists yet

The original README describes four phases:
1. **Fase 1**: Extract acoustic features from audio files (BPM, key, energy, etc.)
2. **Fase 2**: Enrich with MusicBrainz metadata
3. **Fase 3**: Visualize the collection as an interactive map
4. **Fase 4**: Group by acoustic similarity

We are evaluating Go libraries to rebuild this from scratch in Go, TDD-first.

## Domain: Audio Decoding & Feature Extraction

### Audio Decoders (INPUT layer)

These are all mature and production-ready:

| Library | Format | Version | Maintained | Importers | Notes |
|---------|--------|---------|------------|-----------|-------|
| `github.com/go-audio/wav` | WAV | v1.1.0 | ✅ (2022) | 314 | Full decoder + encoder, metadata, cue points. Uses `go-audio/audio` buffer types. |
| `github.com/hajimehoshi/go-mp3` | MP3 | v0.3.4 | ✅ (2022) | 294 | Pure Go MP3 decoder. Outputs 16-bit 2-channel PCM. `Read`, `SampleRate`, `Seek`. |
| `github.com/mewkiz/flac` | FLAC | v1.0.13 | ✅ (Jul 2025) | 96 | Full FLAC decoder + encoder. Frame-level API, seek support, metadata parsing. Excellent docs. |
| `github.com/go-audio/audio` | (PCM utils) | v1.0.0 | ⚠️ (2018) | 251 | Core audio buffer abstractions (`FloatBuffer`, `IntBuffer`, `PCMBuffer`). Utility layer. |
| `github.com/mjibson/go-dsp/wav` | WAV (basic) | — | ⚠️ | via go-dsp | Basic WAV reader. Part of go-dsp package. |

**Verdict**: The decoding layer is solid. The go-audio ecosystem covers WAV, the mewkiz/flac library is excellent for FLAC (actively maintained!), and go-mp3 is battle-tested. These three cover the main audio formats in a music library.

### DSP / Feature Extraction Libraries (ANALYSIS layer)

This is where the gap is. Go has NO librosa equivalent with built-in BPM, key, chroma, or MFCC extraction.

| Library | What it provides | Stars | License | Gap |
|---------|-----------------|-------|---------|-----|
| `github.com/madelynnblue/go-dsp` | FFT, spectral (Pwelch), window functions, basic WAV reader | 915 | ISC | **No BPM, key, chroma, MFCC, or onset detection** |
| `github.com/go-audio/audio` | Buffer interfaces, format conversion | — | Apache-2.0 | Utility only, no analysis |

Specific analysis provided by `go-dsp`:
- **`fft`**: FFTReal, FFT (complex), IFFT — fundamental building block for spectral analysis
- **`spectral`**: Power spectral density (Pwelch) — needed for energy/spectral centroid
- **`window`**: Hamming, Hann, Bartlett, Blackman, etc. — needed before FFT
- **`dsputils`**: Data structures for DSP operations
- **`wav`**: WAV file reader

`go-dsp` provides enough primitives to IMPLEMENT the missing feature extractors from scratch. This is both a risk (more work) and an opportunity (perfect for TDD learning — implement algorithms test-first).

### Missing Analysis Algorithms (would need implementation)

| Feature | Python (librosa) | Go status | Complexity |
|---------|-----------------|-----------|------------|
| BPM / Tempo | `librosa.beat.tempo` | **NOT AVAILABLE** | Medium — onset detection + autocorrelation on go-dsp FFT |
| Key detection | `librosa.feature.tonnetz` + chroma | **NOT AVAILABLE** | High — chromagram + Krumhansl-Schmuckler key profiles |
| Spectral centroid | `librosa.feature.spectral_centroid` | **NOT AVAILABLE** | Low — weighted mean of FFT magnitudes |
| Spectral rolloff | `librosa.feature.spectral_rolloff` | **NOT AVAILABLE** | Low — cumulative sum of FFT magnitudes |
| Energy / RMS | `librosa.feature.rms` | **NOT AVAILABLE** | Low — root mean square of signal |
| Chroma features | `librosa.feature.chroma_stft` | **NOT AVAILABLE** | Medium — FFT bin → chroma mapping |
| MFCC | `librosa.feature.mfcc` | **NOT AVAILABLE** | Medium-High — mel filterbank + DCT |
| Zero-crossing rate | `librosa.feature.zero_crossing_rate` | **NOT AVAILABLE** | Very Low — simple signal threshold |

### Alternative: cgo bindings to C libraries

Using `cgo` to wrap a C/C++ audio library (like Essentia, aubio, or librosa's C backend) is possible but adds complexity:
- Cross-compilation breaks
- Build-time dependency on C toolchain
- Against the "pure Go" TDD learning spirit

**Recommendation**: Implement feature extraction from first principles in Go using `go-dsp` primitives. This is the core TDD learning opportunity. Start with simple features (RMS energy, zero-crossing rate, spectral centroid) and progressively add BPM and key detection.

## Domain: Metadata & MusicBrainz

### Local Tag Reading

| Library | What it does | Version | Importers | Notes |
|---------|-------------|---------|-----------|-------|
| `github.com/dhowden/tag` | Read ID3v1/v2, MP4, FLAC, OGG tags | pseudo-v0 (no tag) | 351 | De facto standard for Go metadata reading |

`dhowden/tag` provides a simple `Metadata` interface:
```go
type Metadata interface {
    Format() Format; FileType() FileType
    Title() string; Album() string; Artist() string
    AlbumArtist() string; Composer() string; Genre() string
    Year() int; Track() (int, int); Disc() (int, int)
    Picture() *Picture; Lyrics() string
    Comment() string; Raw() map[string]interface{}
}
```

**Crucially**, `dhowden/tag` also has a **`mbz` sub-package** that extracts MusicBrainz Picard-specific tags already embedded in the file. This is the bridge between local metadata and MusicBrainz enrichment.

### MusicBrainz API Client

| Library | What it does | Version | Active | Notes |
|---------|-------------|---------|--------|-------|
| `go.uploadedlobster.com/musicbrainzws2` | Full MusicBrainz API v2 client | v0.19.0 | ✅ (Jun 2026) | lookup, browse, search, submit ISRCs, collections |

This is the only serious Go MusicBrainz client. Supports lookup, browse, and search across all entities (artists, releases, recordings, etc.). Last commit was June 2026 — actively maintained. MIT license.

Import path: `go get go.uploadedlobster.com/musicbrainzws2`

### Metadata Pipeline

```
Audio File → dhowden/tag → local metadata (title, artist, album, etc.)
                          → dhowden/tag/mbz → existing MusicBrainz IDs
                          → go-musicbrainzws2 → enrichment from API
```

## Domain: Go TDD & Project Layout

### Project Structure

Two viable approaches:

**A. Standard Go Project Layout** (`cmd/`, `internal/`, `pkg/`)
```
mapa-musical-tdd/
├── cmd/
│   └── mapa-musical/         # CLI entry point
├── internal/
│   ├── audio/                # Feature extraction (TDD from scratch)
│   │   ├── decoder.go        #   AudioFile interface
│   │   ├── features.go       #   BPM, key, energy extractors
│   │   └── features_test.go
│   ├── metadata/             # Tag reading + MusicBrainz
│   │   ├── reader.go
│   │   ├── enricher.go
│   │   └── enricher_test.go
│   ├── model/                # Domain types
│   │   ├── track.go
│   │   └── collection.go
│   └── storage/              # Persistence (SQLite/JSON)
├── pkg/                      # Public API (if any)
├── go.mod
├── go.sum
└── README.md
```

**B. Flat / Domain-focused** (simpler, better for learning)
```
mapa-musical-tdd/
├── cmd/mapa-musical/
├── audio/
├── metadata/
├── model/
├── storage/
├── go.mod
└── README.md
```

**Recommendation**: Start with flat layout (B) for the learning phase. Migrate to standard layout (A) when the codebase grows beyond ~15 packages. Early-stage projects benefit from simplicity.

### Go TDD Patterns

Go's built-in testing is excellent:

- **Table-driven tests**: Define input/output pairs in a slice, iterate.
- **Subtests** (`t.Run`): Organize by scenario within a single test function.
- **Golden files** (`testdata/`): Store expected output for complex assertions.
- **Coverage**: `go test ./... -cover -coverprofile=coverage.out`
- **Race detector**: `go test -race ./...`
- **Linting**: `golangci-lint run` (configure via `.golangci.yml`)
- **Formatting**: `gofmt`/`goimports` — mandatory, enforced by CI

Config excerpt from `openspec/config.yaml`:
```yaml
apply:
  tdd: true
  test_command: "go test ./... -v -cover -coverprofile=coverage.out"
verify:
  test_command: "go test ./... -v -cover -coverprofile=coverage.out"
  coverage_threshold: 80
```

## Domain: Architecture Patterns

### Hexagonal / Clean Architecture in Go

Go's implicit interfaces make hexagonal architecture natural. Key patterns:

**Ports & Adapters**:
```go
// Port (interface) — defined in domain package
type FeatureExtractor interface {
    ExtractBPM(samples []float64, sampleRate int) (float64, error)
    ExtractKey(samples []float64, sampleRate int) (string, error)
}

// Adapter (implementation) — in audio package
type GoDSPExtractor struct {
    // uses go-dsp primitives internally
}

// Wiring (in main or a wire package)
func main() {
    extractor := audio.NewGoDSPExtractor()
    service := service.NewAnalysisService(extractor)
}
```

**Dependency Injection via Constructor**:
```go
func NewAnalysisService(extractor FeatureExtractor) *AnalysisService {
    return &AnalysisService{extractor: extractor}
}
```

**Package dependency direction**:
```
cmd → service → domain, audio, metadata, storage
                    ↑
                 model (no deps)
```

`model/` has zero imports. `audio/`, `metadata/`, `storage/` implement interfaces defined in `model/` or `domain/`.

### Pipeline Architecture

The data flow is naturally a pipeline:

```
AudioFile → Decoder → PCM samples → FeatureExtractor → TrackFeatures
                                                         ↓
AudioFile → TagReader → LocalMetadata ──────────────────→ Track (enriched)
                                                         ↓
              MusicBrainzClient ← MusicBrainz IDs ───────→ EnrichedTrack
                                                         ↓
                                                     Storage (SQLite/JSON)
```

Each stage is independently testable via interfaces:
- `Decoder` interface: test with real audio fixtures in `testdata/`
- `FeatureExtractor` interface: test with synthetic signals (sine waves at known BPM/frequency)
- `TagReader` interface: test with fixture files with known tags
- `MusicBrainzClient` interface: test with HTTP mocks or recorded responses (golden files)

## Domain: Visualization Options

### Go-native options (limited)

| Option | Type | Maturity | Notes |
|--------|------|----------|-------|
| `gonum/plot` | Static charts | Mature | Basic 2D plotting, not interactive |
| `go-echarts` | Interactive charts | Good | Go wrapper for Apache ECharts, generates JS/HTML |
| `fyne` | Desktop GUI | Mature | Cross-platform GUI toolkit, could render scatter plots |
| `gioui` | UI toolkit | Emerging | Immediate-mode GUI, low-level |

### Go Backend + Web Frontend (RECOMMENDED)

| Stack | Role |
|-------|------|
| Go HTTP server | API serving track data, feature vectors, similarity queries |
| `net/http` or `chi` | Routing |
| JSON/Protobuf | Data serialization |
| D3.js / Three.js / Leaflet | Interactive 2D/3D visualization in browser |
| Wasm + Go | Potential for client-side audio analysis |

**Recommendation**: The "musical map" concept (2D/3D spatial visualization of tracks by acoustic similarity) is inherently visual and interactive. A Go backend + web frontend is the right split. Go handles the heavy lifting (audio analysis, metadata, similarity computation), JavaScript/TypeScript handles rendering. This also opens the door to Go-compiled-to-Wasm for browser-side analysis if needed later.

## Approaches Summary

### 1. Pure Go + Build from DSP primitives (RECOMMENDED)
- Use `go-dsp` for FFT, spectral analysis, window functions
- Implement BPM, key, chroma, MFCC from first principles in TDD
- Use `go-audio/wav`, `go-mp3`, `mewkiz/flac` for decoding
- Use `dhowden/tag` + `go-musicbrainzws2` for metadata
- Go backend + web frontend for visualization
- **Pros**: Maximum learning, idiomatic Go, no C dependencies, full control
- **Cons**: BPM/key detection requires DSP knowledge, more code to write
- **Complexity**: Medium-High (but perfect for TDD learning)

### 2. cgo wrapper around C library (NOT recommended)
- Wrap Essentia or aubio via cgo
- **Pros**: Feature extraction "just works"
- **Cons**: Build complexity, cross-compilation breaks, less Go-idiomatic, defeats TDD learning purpose
- **Complexity**: Low for features, High for build maintenance

### 3. Hybrid: Go for infrastructure, call Python for analysis (NOT recommended for this project)
- Go orchestrates, calls Python scripts via `os/exec` for feature extraction
- **Pros**: Reuse librosa's rich feature set
- **Cons**: Two runtimes, deployment complexity, not a "Go project"
- **Complexity**: Low implementation, High operational

## Recommendation

**Go with Approach 1 — Pure Go, TDD from DSP primitives.**

This is a TDD learning laboratory. Building feature extractors from first principles using FFT primitives is EXACTLY the kind of deep learning this project aims for. The libraries exist for the hard parts (decoding, FFT), and the "missing" libraries are the learning opportunity.

**Phased implementation order**:

1. **Setup**: `go mod init`, project structure, CI with `go test`, `golangci-lint`
2. **Model**: Define `Track`, `TrackFeatures`, `Collection` types in `model/` — test the types
3. **Audio Decoding**: Interface + implementations for WAV, MP3, FLAC — test with fixture files
4. **Feature Extraction (easy first)**: RMS energy, zero-crossing rate, spectral centroid — test with synthetic signals
5. **Feature Extraction (complex)**: BPM detection (autocorrelation on onset envelope), chroma features
6. **Metadata**: `dhowden/tag` integration, `go-musicbrainzws2` enrichment — test with mocks
7. **Storage**: SQLite persistence for `Track` + features
8. **Pipeline**: Wire everything together, process a real music directory
9. **Visualization**: Go HTTP API server + web frontend with D3.js/Three.js

## Risks

1. **BPM detection accuracy**: Self-implemented BPM detection may not match librosa's accuracy. This is acceptable for a learning lab — the TDD process IS the goal.
2. **Key detection complexity**: Musical key detection (chromagram + Krumhansl-Schmuckler profiles) is non-trivial. Could be deferred to later phases.
3. **dhowden/tag has no tagged version**: Pseudo-version only. This breaks Go's semantic versioning but is practically fine (351 importers can't be wrong).
4. **go-musicbrainzws2 is on SourceHut**: Slightly less discoverable than GitHub-hosted packages. Import path is `go.uploadedlobster.com/musicbrainzws2`, not the sourcehut URL.
5. **go-audio/audio v1.0.0 is from 2018**: The core audio package is old but stable — the buffer abstractions haven't needed changes. The decoders (wav v1.1.0, 2022) are fresher.
6. **Visualization ecosystem**: Go has no equivalent to Python's plotly/matplotlib ecosystem. The web frontend route is correct but adds a separate technology stack to learn/manage.

## Library Decision Table

| Purpose | Library | Verdict |
|---------|---------|---------|
| WAV decode | `github.com/go-audio/wav` v1.1.0 | ✅ USE |
| MP3 decode | `github.com/hajimehoshi/go-mp3` v0.3.4 | ✅ USE |
| FLAC decode | `github.com/mewkiz/flac` v1.0.13 | ✅ USE |
| PCM buffer types | `github.com/go-audio/audio` v1.0.0 | ✅ USE (for interfacing with decoders) |
| DSP primitives | `github.com/madelynnblue/go-dsp` | ✅ USE (FFT, spectral, windows) |
| Tag reading | `github.com/dhowden/tag` | ✅ USE |
| MusicBrainz API | `go.uploadedlobster.com/musicbrainzws2` v0.19.0 | ✅ USE |
| Feature extraction | **BUILD from go-dsp primitives** | 🔨 BUILD |
| BPM detection | **IMPLEMENT via FFT + autocorrelation** | 🔨 BUILD |
| Key detection | **IMPLEMENT via chromagram + K-S profiles** | 🔨 BUILD |
| HTTP routing | `net/http` (stdlib) or `chi` | ✅ USE |
| Persistent storage | SQLite via `modernc.org/sqlite` or `github.com/mattn/go-sqlite3` | TBD |
| Visualization | Go HTTP API + D3.js/Three.js frontend | ✅ RECOMMENDED |

## Ready for Next Phase

**Yes** — this exploration provides sufficient information to proceed to:
- `sdd-init` (if not already done)
- `sdd-propose` for the first change (project setup & model types)
