package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/SaintJonii/music-map/batch"
	"github.com/SaintJonii/music-map/library"
	"github.com/SaintJonii/music-map/storage"
)

// dbFileName is the persistent SQLite database used by the CLI. It lives in the
// current working directory.
const dbFileName = "mapa-musical.db"

// dedupeSaver adapts storage.Repository to batch.Saver, owning the dedupe
// policy the runner's collector relies on:
//   - unchanged file (same size and mod time) → skipped without re-hash/rewrite
//   - duplicate content (same id or fingerprint) → ErrConflict → skipped
//   - anything else → saved
type dedupeSaver struct {
	repo storage.Repository
}

// SaveAnalyzed implements batch.Saver.
func (s *dedupeSaver) SaveAnalyzed(ctx context.Context, a batch.AnalyzedTrack) (bool, error) {
	// Fast path: skip unchanged files by size+mtime before re-hashing or writing.
	stored, err := s.repo.FindByPath(ctx, a.Track.FilePath)
	switch {
	case err == nil:
		if stored.Size == a.Size && stored.ModTime.Equal(a.ModTime) {
			return true, nil
		}
	case errors.Is(err, storage.ErrNotFound):
		// First time seeing this path — fall through to save.
	default:
		return false, fmt.Errorf("find by path: %w", err)
	}

	err = s.repo.SaveAnalyzed(ctx, storage.AnalyzedTrack{
		Track:       a.Track,
		Features:    a.Features,
		Fingerprint: a.Fingerprint,
		Size:        a.Size,
		ModTime:     a.ModTime,
	})
	if errors.Is(err, storage.ErrConflict) {
		// Already analyzed (same id or same content fingerprint) — a skip, not
		// a failure.
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("save analyzed: %w", err)
	}
	return false, nil
}

// runBatch wires the local scanner, the persistent repository, and the batch
// runner, runs the analysis, and returns the aggregated summary. A corrupt file
// is reported in the summary; only setup errors (e.g. storage init, scan
// failure) are returned as an error.
func runBatch(ctx context.Context, root, dbPath string) (batch.Summary, error) {
	repo, err := storage.NewRepository(dbPath)
	if err != nil {
		return batch.Summary{}, fmt.Errorf("initialize storage: %w", err)
	}
	defer func() { _ = repo.Close() }()

	scanner := library.NewScanner(root)
	runner := batch.NewRunner(scanner, &dedupeSaver{repo: repo})

	summary, err := runner.Run(ctx)
	if err != nil {
		return summary, fmt.Errorf("run: %w", err)
	}
	return summary, nil
}

// printSummary writes the end-of-run summary (counts + per-file failures) to w.
func printSummary(w io.Writer, s batch.Summary) {
	fmt.Fprintf(w, "Scanned:  %d\n", s.Total)
	fmt.Fprintf(w, "Analyzed: %d\n", s.Succeeded)
	fmt.Fprintf(w, "Skipped:  %d\n", s.Skipped)
	fmt.Fprintf(w, "Failed:   %d\n", s.Failed)
	for _, f := range s.Failures {
		fmt.Fprintf(w, "  - %s: %v\n", f.Ref.ID, f.Err)
	}
}

func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <library-folder>\n", os.Args[0])
		os.Exit(1)
	}

	summary, err := runBatch(ctx, os.Args[1], dbFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printSummary(os.Stdout, summary)
}
