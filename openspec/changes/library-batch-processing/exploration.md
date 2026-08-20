# Exploration: library-batch-processing

## Current State

The pipeline is **single-file, sequential, in-memory**. `cmd/mapa-musical/main.go`:
- `main()` opens `storage.NewRepository(":memory:")` (line 122), reads exactly one path from `os.Args[1]` (line 136), calls `pipeline.processFile` once, prints, exits.
- `pipeline.processFile` (lines 32–110) does the full flow on one file: `os.Open` → `TagReader.ReadTags` → `Seek(0)` → `audio.DetectFormat` → `decoder.Decode` → 9 feature extractor calls → `repo.Save`. Every feature warning goes to stderr; any fatal step aborts the whole file.
- There is **no concurrency anywhere** in Go source (grep for `errgroup|WaitGroup|go func|chan |SetLimit|semaphore` matches only an archived design doc). There is **no batch loop**, no folder scan, no persistence beyond `:memory:` in the CLI.

### FFT redundancy (verified exact cost)

Confirmed: **3 full FFT passes per track**, all over the *same* `samples` buffer (processFile passes `samples` to each extractor):
- `magnitudeSpectrum()` = Hann window + `fft.FFTReal` → called by `ExtractSpectralCentroid` (features.go:115) **and** `ExtractChroma` (features.go:218). **2 FFTs, both Hann.**
- `powerSpectrumFunc()` = Hamming window + `fft.FFTReal` → called by `ExtractMFCCs` (features.go:257). **1 FFT, Hamming.**

Window functions differ: `hannWindow` = `0.5*(1−cos)` (α=0.5) vs `hammingWindow` = `0.54−0.46*cos` (α=0.54). Both are generalized-cosine windows; they differ only in sidelobe rolloff. They **cannot** share one FFT as-is because the FFT output is window-dependent — but `power = magnitude²`, so power can be derived from magnitude if a single window is used.

### Storage

`storage/repository.go` (modernc.org/sqlite, driver name `"sqlite"`, pure Go, no cgo):
- `Repository` interface: `Save(ctx, Track, TrackFeatures)`, `GetByID(ctx, id)`, `List(ctx)`, `Close()`. Implemented by `sqliteRepo` wrapping one `*sql.DB`.
- `NewRepository(path)` = `sql.Open("sqlite", path)` + `migrate()` (CREATE TABLE IF NOT EXISTS for `tracks` + `track_features`, PK on `id`/`track_id`, FK cascade). `:memory:` is just a special DSN value — a file path already works (test `TestRepository_ReopenPreservesData` proves persistence).
- `Save` uses per-call `BeginTx` + two `ExecContext` inserts (no prepared-statement reuse, no batching). Unique-violation detection via string match on the error chain (`isUniqueConstraint`).
- **No WAL, no `busy_timeout`, no indexes** beyond the PKs, no `SetMaxOpenConns` tuning. Default database/sql pool → SQLite single-writer will serialize (or throw `SQLITE_BUSY`) under concurrent `Save`.

## Affected Areas

- `cmd/mapa-musical/main.go` — `os.Args[1]` single-file flow; needs folder-scan + worker-pool driver + failure summary + persistent DB path.
- `audio/features.go` + `audio/dsp.go` — 3× redundant FFT; introduce shared spectrum (once per track).
- `storage/repository.go` — add WAL, `busy_timeout`, connection tuning, fingerprint column + indexes, prepared statements, optional batch tx, and a `SaveBatch`/dedupe query.
- `model/track.go` — likely add a fingerprint field (or keep fingerprint in storage layer only).
- `metadata/reader.go` — `TagReader.ReadTags(io.ReadSeeker)` requires a seekable stream; constrains remote (Navidrome) source design.
- New package `library/` — `LibrarySource` interface + local folder scanner (does not exist today).

## Approaches

### 1. FFT reuse — unify window vs keep both
- **A. Unify on Hann, one FFT per track.** Compute `Spectrum{Mags, Power(=mag²), N}` once; centroid, chroma, and MFCC all consume it. Removes **2 of 3 FFTs**. MFCC numerics shift slightly (Hann vs Hamming) — must verify `features_test.go` tolerance still passes. Hann matches the librosa default (its STFT/MFCC use Hann), so this is the more standard choice anyway.
- **B. Keep Hamming for MFCC, share Hann FFT between centroid+chroma.** Removes **1 of 3 FFTs**, zero numerical change to any feature. Lower risk, smaller win.
- *Effort*: A = Medium (test re-baseline), B = Low.

### 2. Concurrency — worker pool vs errgroup
- **A. Plain worker pool (channels, stdlib only).** `jobs chan string`, `results chan Result{Path, Track, Err}`, `runtime.GOMAXPROCS(0)` workers, one collector goroutine accumulating per-file results + errors. Continue-on-error is natural (no fail-fast). **No new dependency** (`x/sync` is absent from go.sum). Full control of result/error aggregation and context cancellation.
- **B. `errgroup.SetLimit(n)` (golang.org/x/sync/errgroup).** Clean bounded concurrency, but: (a) adds a new direct dependency; (b) `Wait()` returns only the *first* error (fail-fast), so per-file error collection still needs a mutex-guarded slice — little saved vs a hand-rolled pool for a continue-on-error workload.
- **C. `errgroup` + semaphore channel.** Same dependency cost as B, more boilerplate than `SetLimit`.
- *Effort*: A = Low–Medium, B = Medium (new dep + error plumbing), C = Medium.

### 3. Storage upgrades
- **A. Minimal:** WAL + `busy_timeout` via DSN `?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)`, `db.SetMaxOpenConns(1)` (single-writer), prepared statements cached on the repo struct, fingerprint column + UNIQUE index for dedupe. Per-file transactions stay.
- **B. A + batch transaction path** (`SaveBatch(ctx, []Track, []Features)` wrapping N inserts in one tx) for throughput. Optional at 10k scale — per-file tx is ~10k small commits, fine under WAL.
- *Effort*: A = Medium, B = Medium–High.

### 4. Fingerprint — content hash vs path hash
- **A. Content hash (SHA-256 of audio bytes).** Survives rename/move. Costs a full read per file on the *skip-check* pass. Mitigate with a **size+mtime fast-path**: if size AND mtime match a stored row, trust unchanged and skip re-hash; only re-hash on mismatch (new/moved files). On first analysis, tee the decode read through the hasher so the fingerprint is essentially free.
- **B. Path hash (SHA-256 of path).** ~zero cost, but breaks on any rename/move → silent re-analysis. Weak for a music library where files move constantly.
- *Effort*: A = Medium, B = Low.

### 5. LibrarySource abstraction
- **A. New `library/` package**, `local.Scanner` (filepath.WalkDir) implementing `LibrarySource`, Navidrome-ready interface. Proposed minimal shape (see Recommendation).
- **B. Keep in `metadata/`** (archived design sketch placed it there). Weaker — source discovery ≠ metadata.

## Recommendation

1. **FFT**: Approach **1A** (unify on Hann, one FFT, derive power = mag²) — the biggest single win (3→1 FFT) and aligns with librosa convention; re-baseline MFCC tests under tolerance. Fall back to **1B** if test drift is unacceptable. Introduce a `Spectrum` type computed once in the pipeline and thread it through the spectral extractors.
2. **Concurrency**: Approach **2A** — plain worker pool with channels, `GOMAXPROCS` workers (8 cores here), per-file `Result{Path, Err}` collected by a single goroutine; `context` for Ctrl-C. Dependency-free and a natural fit for continue-on-error. (Reserve `errgroup.SetLimit` for a later fail-fast task.)
3. **Storage**: Approach **3A** — WAL + `busy_timeout`, `SetMaxOpenConns(1)`, cached prepared statements, `fingerprint TEXT` column + `UNIQUE INDEX idx_tracks_fingerprint`, plus `CREATE INDEX idx_tracks_file_path`; add BPM/genre/key indexes later. Keep per-file transactions; defer `SaveBatch` unless a benchmark shows need.
4. **Fingerprint**: Approach **4A** — content hash with size+mtime fast-path. Dedupe-safe under concurrency via the **DB UNIQUE index as the backstop**: the skip-check pre-queries fingerprints into an in-memory set (10k entries is trivial), and any check-then-insert race resolves to `ErrConflict`, which the worker treats as "already analyzed" (idempotent skip).
5. **LibrarySource**: Approach **5A** — new `library/` package.

Proposed minimal interface (Navidrome-ready):
```go
package library

type TrackRef struct {
    ID      string // path today, stream/record ID for Navidrome
    Size    int64
    ModTime time.Time
}

type LibrarySource interface {
    List(ctx context.Context) ([]TrackRef, error)
    Open(ctx context.Context, ref TrackRef) (io.ReadCloser, error)
}
```
Note the seekability tension: `TagReader.ReadTags(io.ReadSeeker)` + `DetectFormat`'s re-read requires a seekable stream. Local files (`*os.File`) satisfy this; a Navidrome HTTP stream would need buffering (memory/temp) to be seekable — flag in design.

## Risks

- **MFCC value drift** from window unification (Hann→Hamming change) — must re-run `audio/features_test.go` with tolerance assertions; if it fails badly, drop to 1B.
- **SQLite single-writer** under concurrency: with a worker pool, concurrent `Save` needs WAL + `busy_timeout` + `SetMaxOpenConns(1)`, or `SQLITE_BUSY` errors surface as spurious failures.
- **Remote-source seekability**: `TagReader`/`DetectFormat` assume `io.ReadSeeker`; a Navidrome HTTP source can't provide that without buffering — don't let `LibrarySource` imply seekability it can't guarantee.
- **Content-hash cost on skip-check**: without the size+mtime fast-path, a full re-hash of ~10k×~30–50MB per scan is expensive (~300–500GB reads); the fast-path is not optional at scale.
- **Test fixtures are thin for batch/E2E**: only 4 WAV + 1 MP3 + 1 FLAC + `corrupt.mp3` (50B) + `unsupported_8bit.wav`. Batch/E2E tests need either synthetic generation (already the repo pattern) or more/duplicated fixtures; `corrupt.mp3` is the "continue-on-error" fixture.

## Ready for Proposal

**Yes.** All six points are verified against real code. Recommend proposal scope: worker-pool batch runner + persistent WAL SQLite + shared FFT spectrum + `library/` package + content-hash dedupe with size+mtime fast-path. Open decisions for the user: (a) window unification (Hann) vs preserve-Hamming, (b) content-hash confirmed as the fingerprint, (c) worker count = `GOMAXPROCS` acceptable. Explicitly out of scope (per product assumptions): PCA/UMAP, frontend.
