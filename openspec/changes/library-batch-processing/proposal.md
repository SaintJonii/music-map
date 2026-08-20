# Proposal: Library Batch Processing

## Intent
Scale from one in-memory file to a full library (~10k tracks): scan a folder, analyze concurrently, persist to real SQLite, skip already-analyzed tracks — one corrupt file must not stop the run.

## Scope

### In Scope
- `library/` package: `LibrarySource` interface + local scanner (Navidrome/Subsonic-ready).
- Concurrent runner: worker pool (channels, `GOMAXPROCS` workers), single collector, `context` cancellation.
- Continue-on-error: collect per-file failures, end-of-run summary.
- Persistent SQLite: WAL + `busy_timeout` + `SetMaxOpenConns(1)` + cached prepared statements.
- Incremental dedupe: content-hash fingerprint + size/mtime fast-path + UNIQUE index (idempotent skip).
- FFT reuse: one shared spectrum per track (`power = magnitude²`).

### Out of Scope
- PCA/UMAP, frontend, network sources (interface only).
- `SaveBatch`; BPM/genre/key indexes.

## Capabilities

### New Capabilities
- `library-scanning`: `LibrarySource` (`List`/`Open` over `TrackRef`) + `filepath.WalkDir` scanner.
- `batch-processing`: worker-pool runner with per-file results, failure collection, run summary.

### Modified Capabilities
- `persistence`: WAL/busy_timeout/connection tuning, `fingerprint` column + UNIQUE index + `idx_tracks_file_path`, content-hash dedupe.

### Unchanged
`feature-extraction` (FFT refactor is implementation-only), `domain-model`, `metadata-enrichment`, `audio-decoding`, `project-scaffold`.

## Approach
- Worker pool: `jobs`/`results` channels, `GOMAXPROCS` workers, one collector, `context` for Ctrl-C. Stdlib only.
- Storage: WAL + `busy_timeout(5000)` DSN pragma, `SetMaxOpenConns(1)`, prepared statements on repo; per-file transactions.
- Fingerprint: SHA-256 of bytes, teed through decode; size+mtime fast-path; DB UNIQUE index backstops races (conflict = skip).
- FFT: unify on Hann, one `Spectrum{Mags, Power, N}`; fallback preserve-Hamming (2 FFT) if MFCC tolerance fails.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `library/` | New | `LibrarySource` + `local.Scanner` |
| `cmd/mapa-musical/main.go` | Modified | scan + worker pool + summary + DB path |
| `storage/repository.go` | Modified | WAL, tuning, fingerprint, indexes, dedupe |
| `audio/features.go`, `audio/dsp.go` | Modified | shared `Spectrum`, one FFT |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| MFCC drift (Hann vs Hamming) | Med | re-baseline tolerance tests; fallback 2-FFT |
| `SQLITE_BUSY` under concurrency | Med | WAL + busy_timeout + single conn |
| Content-hash cost on skip-check | Med | size+mtime fast-path |
| Thin batch/E2E fixtures | Low | synthetic generation + `corrupt.mp3` |

## Rollback Plan
Git-revert `cmd/`, `storage/`, `audio/`; delete additive `library/`. Single-file flow untouched behind new driver.

## Dependencies
None external; `modernc.org/sqlite` present.

## Success Criteria
- [ ] Folder batch analyzes all valid files; corrupt file doesn't stop run; summary lists failures.
- [ ] Re-run skips unchanged (size+mtime) and re-hashes moved (content hash).
- [ ] Concurrent saves to file SQLite without `SQLITE_BUSY`.
- [ ] `go test ./... -cover` ≥ 80%; features within tolerance.
