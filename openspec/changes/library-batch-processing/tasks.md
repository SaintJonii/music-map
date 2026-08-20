# Tasks: Library Batch Processing

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~900 additions+deletions |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 5 |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

Threat matrix: N/A — no shell/subprocess/VCS/PR boundary, no threat tasks.

### Suggested Work Units

| Unit | Goal | Likely PR | Focused test command | Runtime harness | Rollback boundary |
|------|------|-----------|----------------------|-----------------|-------------------|
| 1 | `library/` discovery | PR 1 | `go test ./library/... -v` | `go build ./...` | delete `library/` |
| 2 | FFT spectrum share | PR 2 | `go test ./audio/... -v` | `features_test.go` unchanged | revert `audio/dsp.go`+`audio/features.go` |
| 3 | `batch/` runner | PR 3 | `go test ./batch/... -v` | fake `LibrarySource` in-process | delete `batch/` |
| 4 | storage hardening | PR 4 | `go test ./storage/... -v` | `t.TempDir()` file DB reopen | revert `storage/` |
| 5 | CLI wiring + E2E | PR 5 | `go test ./cmd/... -v` | `go run ./cmd/mapa-musical testdata/...` | revert `cmd/` |

## Phase 1: Foundation — library discovery + FFT spectrum

- [x] 1.1 RED: `library/scanner_test.go` — table-driven `t.TempDir()` tests for `List`/`Open` (library-scanning: nested scan, non-audio ignored, empty folder, unreadable folder, open listed, missing file)
- [x] 1.2 GREEN: `library/library.go` — `TrackRef{ID,Size,ModTime}`, `LibrarySource`, `ReadSeekCloser`
- [x] 1.3 GREEN: `library/scanner.go` — `local.Scanner` via `filepath.WalkDir`, filter `.wav/.mp3/.flac`
- [x] 1.4 REFACTOR: `audio/dsp.go` export `Spectrum{Mags,Power,N}` + `HannMagnitudeSpectrum`; `audio/features.go` centroid+chroma consume shared spectrum (MFCC keeps Hamming `powerSpectrumFunc`) — `audio/features_test.go` MUST pass unchanged

## Phase 2: Core — batch runner

- [ ] 2.1 RED: `batch/runner_test.go` — fake `LibrarySource`; scenarios: valid batch succeeds, corrupt file isolated, mixed summary, cancel mid-run, deterministic across worker counts
- [ ] 2.2 GREEN: `batch/runner.go` — worker pool (`jobs`/`results`, `GOMAXPROCS`), single collector owns DB writes, `Result`, `Summary`, `analyze` tee+hash, `context` cancellation

## Phase 3: Core — storage hardening

- [ ] 3.1 RED: `storage/repository_test.go` — fingerprint persists, duplicate skipped, concurrent saves no busy, idempotent re-run, WAL reopen, `ErrConflict`
- [ ] 3.2 GREEN: `storage/repository.go` — DSN `_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)` (file paths only), `SetMaxOpenConns(1)`, prepared stmts, `fingerprint`/`size`/`mod_time` columns + partial UNIQUE index + `idx_tracks_file_path`, `SaveAnalyzed`/`FindByPath`/`FingerprintExists`

## Phase 4: Wiring + integration

- [ ] 4.1 GREEN: `cmd/mapa-musical/main.go` — wire `Scanner`+`repo`+`runner`; persistent DB path; end-of-run summary
- [ ] 4.2 E2E: `cmd/mapa-musical/main_test.go` — batch over `testdata/` fixtures + `corrupt.mp3`

## Phase 5: Verify

- [ ] 5.1 `go test ./... -v -cover -coverprofile=coverage.out` ≥80%; `go build ./...`; `gofmt -l .`
