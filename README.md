# Mapa Musical TDD

**A TDD lab for enriching and exploring music collections.**

---

## 🎯 Overview

Mapa Musical is a personal project designed as a **Test-Driven Development (TDD) laboratory** and exploration space for:

- **Enriching** a personal music library with acoustic feature extraction and discographic metadata
- **Visualizing** relationships and patterns within a music collection interactively
- **Learning** TDD in a practical, real-world context

### Why TDD Here

This project is deliberately a **TDD learning lab**. Every feature is developed by writing tests first, enabling:
- Understanding expected behavior before implementation
- Maintaining clean, reliable code from day one
- Documenting system behavior through tests

---

## 📚 Current Status

**Status**: 🟡 Foundation Phase

- ✅ Repository structure
- ✅ Goals and vision definition
- ⏳ Go module initialization and domain model (in progress)
- ⏳ Audio feature extraction pipeline (next)

---

## 🗓️ Roadmap

### Phase 1: Foundation (Go Module + Domain Model)
- Initialize Go module (`go mod init`)
- Define domain types: Track, TrackFeatures, Collection
- Set up CI/linter baseline

### Phase 2: Audio Decoding
- Decoder interface with WAV, MP3, FLAC adapters
- PCM float64 conversion

### Phase 3: Feature Extraction
- Extract BPM, key, energy, spectral centroid, chroma, MFCCs, and ZCR
- Built on go-dsp FFT primitives

### Phase 4: Metadata + Persistence
- Local tag reading (dhowden/tag)
- MusicBrainz enrichment
- SQLite persistence

### Phase 5: CLI + Visualization
- Wire pipeline together
- Interactive 2D visualization (post-MVP)

---

## 🛠️ Tech Stack

- **Go** 1.26+
- **go test** + **coverage** — TDD and coverage baseline
- **golangci-lint** — Linter
- **go-dsp** — FFT/DSP primitives
- **MusicBrainz API** — Metadata enrichment (planned)

---

## 📋 Project Structure

```
mapa-musical-tdd/
├── model/              # Domain types (Track, TrackFeatures, Collection)
├── audio/              # Audio decoding + feature extraction
├── metadata/           # Tag reading + MusicBrainz enrichment
├── storage/            # SQLite persistence
├── testdata/           # Test fixtures (audio files, HTTP responses)
├── openspec/           # SDD specs, design, and tasks
├── go.mod
├── .golangci.yml
├── README.md
└── .gitignore
```

---

## 🧪 How to Run Tests

```bash
# Run all tests with coverage
go test ./... -v -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run linter
golangci-lint run ./...
```

---

**Personal learning and exploration project. 🎵🧪**
