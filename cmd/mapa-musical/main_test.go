package main

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SaintJonii/music-map/batch"
	"github.com/SaintJonii/music-map/library"
	"github.com/SaintJonii/music-map/storage"
)

// fixtureRoot is the path to the shared test fixtures, relative to this package.
const fixtureRoot = "../../testdata"

// failureBaseNames maps each failure's ref to its base filename for stable
// assertions regardless of the absolute path the scanner produces.
func failureBaseNames(failures []batch.Failure) map[string]struct{} {
	m := make(map[string]struct{}, len(failures))
	for _, f := range failures {
		m[filepath.Base(f.Ref.ID)] = struct{}{}
	}
	return m
}

func TestBatchRun_FixturesAndCorruptMP3(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "library.db")

	summary, err := runBatch(context.Background(), fixtureRoot, dbPath)
	if err != nil {
		t.Fatalf("runBatch failed: %v", err)
	}

	// testdata has 7 audio fixtures: 5 valid (16/24/32-bit WAV + FLAC + MP3)
	// and 2 that fail analysis (8-bit WAV, corrupt MP3).
	if summary.Total != 7 {
		t.Errorf("Total = %d, want 7", summary.Total)
	}
	if summary.Succeeded != 5 {
		t.Errorf("Succeeded = %d, want 5", summary.Succeeded)
	}
	if summary.Skipped != 0 {
		t.Errorf("Skipped = %d, want 0 on first run", summary.Skipped)
	}
	if summary.Failed != 2 {
		t.Fatalf("Failed = %d, want 2 (failures: %v)", summary.Failed, summary.Failures)
	}

	// The deliberately-bad fixtures must be reported as per-file failures, not
	// abort the whole run.
	failed := failureBaseNames(summary.Failures)
	for _, want := range []string{"corrupt.mp3", "unsupported_8bit.wav"} {
		if _, ok := failed[want]; !ok {
			t.Errorf("expected a failure for %s, got %v", want, failed)
		}
	}

	// Persistence: reopen the file DB and confirm only the 5 valid files were
	// written (the 2 failures must not be persisted).
	repo, err := storage.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("reopen repository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	tracks, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tracks) != 5 {
		t.Fatalf("persisted %d tracks, want 5", len(tracks))
	}

	persisted := make(map[string]struct{}, len(tracks))
	for _, tr := range tracks {
		persisted[filepath.Base(tr.FilePath)] = struct{}{}
	}
	for _, want := range []string{
		"stereo_44100_16bit.wav",
		"stereo_44100_24bit.wav",
		"stereo_44100_32bit.wav",
		"test_16bit.flac",
		"test_128kbps.mp3",
	} {
		if _, ok := persisted[want]; !ok {
			t.Errorf("expected %s persisted, got %v", want, persisted)
		}
	}
}

func TestBatchRun_IdempotentRerun(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "library.db")

	first, err := runBatch(context.Background(), fixtureRoot, dbPath)
	if err != nil {
		t.Fatalf("first runBatch failed: %v", err)
	}
	if first.Succeeded != 5 || first.Failed != 2 {
		t.Fatalf("first run: Succeeded=%d Failed=%d, want 5/2", first.Succeeded, first.Failed)
	}

	// Second run over the same fixtures: unchanged files are skipped via the
	// size+mtime fast-path; the bad fixtures fail again.
	second, err := runBatch(context.Background(), fixtureRoot, dbPath)
	if err != nil {
		t.Fatalf("second runBatch failed: %v", err)
	}

	if second.Succeeded != 0 {
		t.Errorf("second run Succeeded = %d, want 0", second.Succeeded)
	}
	if second.Skipped != 5 {
		t.Errorf("second run Skipped = %d, want 5", second.Skipped)
	}
	if second.Failed != 2 {
		t.Errorf("second run Failed = %d, want 2", second.Failed)
	}

	// Idempotency: no duplicate rows after re-run.
	repo, err := storage.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("reopen repository failed: %v", err)
	}
	defer func() { _ = repo.Close() }()

	tracks, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tracks) != 5 {
		t.Errorf("after re-run, persisted %d tracks, want 5 (no duplicates)", len(tracks))
	}
}

func TestBatchRun_NonexistentFolder(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "library.db")
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	_, err := runBatch(context.Background(), missing, dbPath)
	if err == nil {
		t.Fatal("expected error for nonexistent folder, got nil")
	}
}

func TestPrintSummary_OutputsCountsAndFailures(t *testing.T) {
	var buf bytes.Buffer
	s := batch.Summary{
		Total:     3,
		Succeeded: 1,
		Skipped:   1,
		Failed:    1,
		Failures: []batch.Failure{{
			Ref: library.TrackRef{ID: "/music/corrupt.mp3"},
			Err: errors.New("unsupported audio format"),
		}},
	}

	printSummary(&buf, s)

	got := buf.String()
	for _, want := range []string{
		"3",                        // total count appears
		"1",                        // counts appear
		"corrupt.mp3",              // per-file failure path
		"unsupported audio format", // per-file failure reason
	} {
		if !strings.Contains(got, want) {
			t.Errorf("summary output missing %q:\n%s", want, got)
		}
	}
}
