# Design: Library Batch Processing

## Technical Approach

Scale the single-file pipeline to a folder (~10k tracks): a `library/` discovery layer, a stdlib worker-pool runner in a new `batch/` package, and hardened SQLite persistence. Workers only compute; one collector goroutine owns every DB write. SHA-256 fingerprint is computed while decoding (tee) with a size+mtime fast-path and UNIQUE-index backstop. FFT drops 3→2 passes, zero drift.

## Architecture Decisions

### 1. Library discovery + seekability
| Option | Tradeoff | Decision |
|---|---|---|
| `Open` → `io.ReadCloser` | TagReader/DetectFormat need seek | Reject |
| `Open` → `ReadSeekCloser` | `*os.File` native; HTTP must buffer | **Choose** |

`TrackRef{ID, Size, ModTime}` (`ID` = path today, stream ID for Navidrome). Local scanner returns `*os.File`. Navidrome (future) must spool to temp/memory — a source contract, not an interface promise.

### 2. Concurrency
| Option | Tradeoff | Decision |
|---|---|---|
| `errgroup.SetLimit` | New dep; fails fast | Reject |
| Worker pool (channels) | Stdlib; continue-on-error | **Choose** |

`jobs`/`results` channels, `GOMAXPROCS` workers, one collector owns persistence, `context` cancellation.

### 3. FFT reuse (resolved — do not reopen)
| Option | Tradeoff | Decision |
|---|---|---|
| 3→1 unify Hann | MFCC drifts (Hann ≠ Hamming) | Deferred |
| 3→2 share Hann | Saves 1 FFT, zero drift | **Choose** |

Compute Hann magnitude spectrum ONCE per track (new `audio.Spectrum`), shared by `ExtractSpectralCentroid`+`ExtractChroma`; `ExtractMFCCs` keeps Hamming `powerSpectrumFunc`. Rationale: optimization vs numeric stability.

### 4. Fingerprint / dedupe
| Option | Tradeoff | Decision |
|---|---|---|
| Path hash | Renames → silent re-analysis | Reject |
| Content SHA-256 | Extra read on skip-check | **Choose** + fast-path |

Hash teed through decode. size+mtime fast-path skips re-hash. Partial UNIQUE index backstops races (conflict = "already analyzed").

### 5. Storage
| Option | Tradeoff | Decision |
|---|---|---|
| WAL + `busy_timeout` + `SetMaxOpenConns(1)` + cached stmts | Single-writer | **Choose** |
| `SaveBatch` | Unnecessary at 10k | Deferred |

DSN `?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)` (file paths only; `:memory:` bare), `SetMaxOpenConns(1)`, cached `*sql.Stmt`, `fingerprint`/`size`/`mod_time` columns, partial `CREATE UNIQUE INDEX idx_tracks_fingerprint ... WHERE fingerprint <> ''` (avoids collision on `DEFAULT ''`), plus `idx_tracks_file_path`. Fingerprint/size/mtime stay in storage, NOT `model.Track`.

### 6. Continue-on-error
Collector aggregates per-file failures into a `Summary` (counts + failed paths + reasons). One corrupt file never aborts the run.

## Data Flow

```
Scanner.List ─→ jobs chan ─→ workers (open → tee+hash → tags → decode → features)
                                   │
                                   └─→ results chan ─→ collector (skip | save | summary)
```

## File Changes

| File | Action | Description |
|---|---|---|
| `library/library.go` | Create | `TrackRef`, `LibrarySource`, `ReadSeekCloser` |
| `library/scanner.go` | Create | `local.Scanner` (`filepath.WalkDir`, `.wav/.mp3/.flac`) |
| `library/scanner_test.go` | Create | `t.TempDir()`; filter; WalkDir errors |
| `batch/runner.go` | Create | Worker pool, `Result`, `Summary`, `analyze` (tee+hash) |
| `batch/runner_test.go` | Create | Fake `LibrarySource`; corrupt-file; dedupe skip |
| `cmd/mapa-musical/main.go` | Modify | Wire scanner+repo+runner; persistent DB; summary |
| `storage/repository.go` | Modify | Pragmas, `SetMaxOpenConns(1)`, stmts, columns+indexes, `SaveAnalyzed`, `FindByPath`, `FingerprintExists` |
| `storage/repository_test.go` | Modify | WAL reopen, conflict, fast-path, concurrency |
| `audio/dsp.go` | Modify | Export `Spectrum` + `HannMagnitudeSpectrum` |
| `audio/features.go` | Modify | Centroid/Chroma consume shared spectrum |

## Interfaces / Contracts

```go
package library

type TrackRef struct {
    ID      string
    Size    int64
    ModTime time.Time
}

type ReadSeekCloser interface {
    io.Reader
    io.Seeker
    io.Closer
}

type LibrarySource interface {
    List(ctx context.Context) ([]TrackRef, error)
    Open(ctx context.Context, ref TrackRef) (ReadSeekCloser, error)
}
```

Tee-for-hash (hash whole file while decoding):

```go
h := sha256.New()
f, _ := src.Open(ctx, ref)
_ = tagReader.ReadTags(f)          // tags first (existing strategy)
_, _ = f.Seek(0, io.SeekStart)     // rewind so hash covers whole file
decoder, r, _ := audio.DetectFormat(io.TeeReader(f, h))
samples, sr, ch, _ := decoder.Decode(r)
fingerprint := hex.EncodeToString(h.Sum(nil))
```

## Testing Strategy

| Layer | What | Approach |
|---|---|---|
| Unit | scanner filter/errors | Table-driven, `t.TempDir()` |
| Unit | continue-on-error, dedupe skip, cancel | Fake source+repo; `context.WithCancel` |
| Unit | WAL reopen, UNIQUE conflict, fast-path | `t.TempDir()` DB; `errors.Is(ErrConflict)` |
| Unit | FFT zero-drift | Existing `features_test.go` unchanged must pass |
| Integration | batch over fixtures + `corrupt.mp3` | `TestPipeline_WAVFixture` pattern; `-cover` ≥ 80% |

## Threat Matrix

N/A — no routing, shell, subprocess, VCS/PR automation, executable-file classification, or process-integration boundary.

## Migration / Rollout

No data migration: columns/indexes are additive on `migrate()`. Rollback = git-revert `cmd/`, `storage/`, `audio/`; delete `library/`, `batch/`.

## Open Questions

None blocking. Navidrome HTTP `Open` buffering (memory vs temp-file) is deferred to the network-source change.
