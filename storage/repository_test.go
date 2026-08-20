package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/SaintJonii/music-map/model"
)

func TestRepository_SaveAndGetByID(t *testing.T) {
	repo, err := NewRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	ctx := context.Background()
	track := model.Track{
		ID:    "track-1",
		Title: "Save Test",
	}
	features := model.TrackFeatures{
		BPM: 120.0,
		Key: "C major",
		ZCR: 0.15,
	}

	if err := repo.Save(ctx, track, features); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	gotTrack, gotFeatures, err := repo.GetByID(ctx, "track-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if gotTrack.Title != "Save Test" {
		t.Errorf("expected Title 'Save Test', got %q", gotTrack.Title)
	}
	if gotFeatures.BPM != 120.0 {
		t.Errorf("expected BPM 120.0, got %v", gotFeatures.BPM)
	}
	if gotFeatures.Key != "C major" {
		t.Errorf("expected Key 'C major', got %q", gotFeatures.Key)
	}
}

func TestRepository_DuplicateID_ReturnsConflict(t *testing.T) {
	repo, err := NewRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	ctx := context.Background()
	track := model.Track{ID: "dup-1", Title: "First"}
	features := model.TrackFeatures{}

	if err := repo.Save(ctx, track, features); err != nil {
		t.Fatalf("first Save failed: %v", err)
	}

	// Second save with same ID.
	err = repo.Save(ctx, track, features)
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got: %v", err)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	repo, err := NewRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	ctx := context.Background()
	_, _, err = repo.GetByID(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestRepository_List(t *testing.T) {
	repo, err := NewRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	ctx := context.Background()
	tracks := []model.Track{
		{ID: "l-1", Title: "First"},
		{ID: "l-2", Title: "Second"},
		{ID: "l-3", Title: "Third"},
	}

	for _, tr := range tracks {
		if err := repo.Save(ctx, tr, model.TrackFeatures{}); err != nil {
			t.Fatalf("Save %s failed: %v", tr.ID, err)
		}
	}

	results, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 tracks, got %d", len(results))
	}

	// Verify insertion order.
	if results[0].Title != "First" {
		t.Errorf("expected 'First' first, got %q", results[0].Title)
	}
	if results[1].Title != "Second" {
		t.Errorf("expected 'Second' second, got %q", results[1].Title)
	}
	if results[2].Title != "Third" {
		t.Errorf("expected 'Third' third, got %q", results[2].Title)
	}
}

func TestRepository_List_Empty(t *testing.T) {
	repo, err := NewRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	ctx := context.Background()
	results, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List on empty DB failed: %v", err)
	}

	if results == nil {
		t.Error("expected empty slice (not nil)")
	} else if len(results) != 0 {
		t.Errorf("expected 0 tracks, got %d", len(results))
	}
}

func TestRepository_SchemaAutoMigration(t *testing.T) {
	// First save on a fresh :memory: DB creates the schema.
	repo, err := NewRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	ctx := context.Background()
	track := model.Track{
		ID:          "schema-1",
		Title:       "Schema Test",
		Artist:      "Artist",
		Album:       "Album",
		AlbumArtist: "AlbumArtist",
		Genre:       "Genre",
		Year:        2024,
		TrackNumber: 1,
		ISRC:        "US-TST-12-00001",
	}

	features := model.TrackFeatures{
		BPM:              128.0,
		Key:              "A minor",
		Energy:           0.75,
		Danceability:     0.82,
		Acousticness:     0.15,
		SpectralCentroid: 2200.0,
		Chroma:           [12]float64{0.1, 0.0, 0.2, 0.0, 0.3, 0.0, 0.1, 0.0, 0.2, 0.0, 0.1, 0.0},
		MFCCs:            [13]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0, 12.0, 13.0},
		ZCR:              0.25,
	}

	if err := repo.Save(ctx, track, features); err != nil {
		t.Fatalf("Save with full schema failed: %v", err)
	}

	// Retrieve and verify all fields persisted.
	gotTrack, gotFeatures, err := repo.GetByID(ctx, "schema-1")
	if err != nil {
		t.Fatalf("GetByID after migration failed: %v", err)
	}

	if gotTrack.Title != "Schema Test" {
		t.Errorf("Title mismatch: got %q", gotTrack.Title)
	}
	if gotTrack.Artist != "Artist" {
		t.Errorf("Artist mismatch: got %q", gotTrack.Artist)
	}
	if gotTrack.Year != 2024 {
		t.Errorf("Year mismatch: got %d", gotTrack.Year)
	}
	if gotTrack.ISRC != "US-TST-12-00001" {
		t.Errorf("ISRC mismatch: got %q", gotTrack.ISRC)
	}
	if gotFeatures.BPM != 128.0 {
		t.Errorf("BPM mismatch: got %v", gotFeatures.BPM)
	}
	if gotFeatures.Key != "A minor" {
		t.Errorf("Key mismatch: got %q", gotFeatures.Key)
	}
}

func TestRepository_ReopenPreservesData(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// First session: save a track.
	repo1, err := NewRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create first repository: %v", err)
	}

	ctx := context.Background()
	track := model.Track{ID: "persist-1", Title: "Persistent"}
	if err := repo1.Save(ctx, track, model.TrackFeatures{}); err != nil {
		t.Fatalf("first Save failed: %v", err)
	}
	_ = repo1.Close()

	// Second session: open the same file.
	repo2, err := NewRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen repository: %v", err)
	}
	defer func() { _ = repo2.Close() }()

	gotTrack, _, err := repo2.GetByID(ctx, "persist-1")
	if err != nil {
		t.Fatalf("GetByID after reopen failed: %v", err)
	}

	if gotTrack.Title != "Persistent" {
		t.Errorf("expected 'Persistent', got %q", gotTrack.Title)
	}
}

func TestRepository_CorruptDatabase_NoPanic(t *testing.T) {
	dir := t.TempDir()
	corruptPath := filepath.Join(dir, "corrupt.db")

	// Write random bytes as a "corrupt" database.
	if err := os.WriteFile(corruptPath, []byte("this is not a valid SQLite database"), 0644); err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}

	// Should return error, not panic.
	_, err := NewRepository(corruptPath)
	if err == nil {
		t.Fatal("expected error opening corrupt DB, got nil")
	}
	t.Logf("corrupt DB error (expected): %v", err)
}

// --- Phase 3 (PR4): storage hardening ---

func newRepo(t *testing.T, path string) Repository {
	t.Helper()
	repo, err := NewRepository(path)
	if err != nil {
		t.Fatalf("NewRepository(%q) failed: %v", path, err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	return repo
}

func TestRepository_SaveAnalyzed_PersistsFingerprint(t *testing.T) {
	repo := newRepo(t, ":memory:")
	ctx := context.Background()

	modTime := time.Unix(1_700_000_000, 123_456_789)
	a := AnalyzedTrack{
		Track: model.Track{
			ID:       "fp-1",
			Title:    "Fingerprint Track",
			FilePath: "/music/fp-1.mp3",
		},
		Features:    model.TrackFeatures{BPM: 128.0},
		Fingerprint: "deadbeefcafef00d",
		Size:        4096,
		ModTime:     modTime,
	}

	if err := repo.SaveAnalyzed(ctx, a); err != nil {
		t.Fatalf("SaveAnalyzed failed: %v", err)
	}

	stored, err := repo.FindByPath(ctx, "/music/fp-1.mp3")
	if err != nil {
		t.Fatalf("FindByPath failed: %v", err)
	}
	if stored.Fingerprint != "deadbeefcafef00d" {
		t.Errorf("fingerprint: expected %q, got %q", "deadbeefcafef00d", stored.Fingerprint)
	}
	if stored.Size != 4096 {
		t.Errorf("size: expected 4096, got %d", stored.Size)
	}
	if !stored.ModTime.Equal(modTime) {
		t.Errorf("mod_time: expected %v, got %v", modTime, stored.ModTime)
	}
	if stored.Track.ID != "fp-1" {
		t.Errorf("track id: expected fp-1, got %q", stored.Track.ID)
	}

	exists, err := repo.FingerprintExists(ctx, "deadbeefcafef00d")
	if err != nil {
		t.Fatalf("FingerprintExists failed: %v", err)
	}
	if !exists {
		t.Error("expected FingerprintExists to report true after save")
	}
}

func TestRepository_SaveAnalyzed_DuplicateFingerprint_Conflict(t *testing.T) {
	repo := newRepo(t, ":memory:")
	ctx := context.Background()

	first := AnalyzedTrack{
		Track:       model.Track{ID: "dup-a", FilePath: "/music/a.mp3"},
		Fingerprint: "same-content-hash",
	}
	if err := repo.SaveAnalyzed(ctx, first); err != nil {
		t.Fatalf("first SaveAnalyzed failed: %v", err)
	}

	// Different ID and path, same content fingerprint → duplicate.
	second := AnalyzedTrack{
		Track:       model.Track{ID: "dup-b", FilePath: "/music/b.mp3"},
		Fingerprint: "same-content-hash",
	}
	err := repo.SaveAnalyzed(ctx, second)
	if err == nil {
		t.Fatal("expected conflict for duplicate fingerprint, got nil")
	}
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestRepository_FingerprintExists(t *testing.T) {
	repo := newRepo(t, ":memory:")
	ctx := context.Background()

	exists, err := repo.FingerprintExists(ctx, "not-there")
	if err != nil {
		t.Fatalf("FingerprintExists failed: %v", err)
	}
	if exists {
		t.Error("expected false before any save")
	}

	if err := repo.SaveAnalyzed(ctx, AnalyzedTrack{
		Track:       model.Track{ID: "e-1", FilePath: "/music/e-1.mp3"},
		Fingerprint: "hash-1",
	}); err != nil {
		t.Fatalf("SaveAnalyzed failed: %v", err)
	}

	exists, err = repo.FingerprintExists(ctx, "hash-1")
	if err != nil {
		t.Fatalf("FingerprintExists failed: %v", err)
	}
	if !exists {
		t.Error("expected true after save")
	}

	// Empty fingerprint must never be reported present: the partial UNIQUE index
	// excludes DEFAULT '', so many tracks can share the empty fingerprint.
	if err := repo.SaveAnalyzed(ctx, AnalyzedTrack{
		Track: model.Track{ID: "e-2", FilePath: "/music/e-2.mp3"},
	}); err != nil {
		t.Fatalf("SaveAnalyzed (empty fingerprint) failed: %v", err)
	}
	exists, err = repo.FingerprintExists(ctx, "")
	if err != nil {
		t.Fatalf("FingerprintExists(empty) failed: %v", err)
	}
	if exists {
		t.Error("empty fingerprint must report false")
	}
}

func TestRepository_ConcurrentSaves_NoBusy(t *testing.T) {
	dir := t.TempDir()
	repo := newRepo(t, filepath.Join(dir, "concurrent.db"))
	ctx := context.Background()

	const n = 50
	var wg sync.WaitGroup
	errs := make(chan error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a := AnalyzedTrack{
				Track: model.Track{
					ID:       fmt.Sprintf("conc-%d", i),
					Title:    fmt.Sprintf("Track %d", i),
					FilePath: fmt.Sprintf("/music/conc-%d.mp3", i),
				},
				Features:    model.TrackFeatures{BPM: float64(100 + i)},
				Fingerprint: fmt.Sprintf("fp-%032d", i),
				Size:        int64(i * 1024),
			}
			errs <- repo.SaveAnalyzed(ctx, a)
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent SaveAnalyzed failed: %v", err)
		}
	}

	tracks, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tracks) != n {
		t.Errorf("expected %d tracks, got %d", n, len(tracks))
	}
}

func TestRepository_SaveAnalyzed_IdempotentRerun(t *testing.T) {
	repo := newRepo(t, ":memory:")
	ctx := context.Background()

	a := AnalyzedTrack{
		Track:       model.Track{ID: "idem-1", Title: "Idempotent", FilePath: "/music/idem.mp3"},
		Fingerprint: "idem-hash",
		Size:        2048,
		ModTime:     time.Unix(1_700_000_000, 0),
	}
	if err := repo.SaveAnalyzed(ctx, a); err != nil {
		t.Fatalf("first SaveAnalyzed failed: %v", err)
	}

	// Re-running the same track must conflict and must not create a new row.
	err := repo.SaveAnalyzed(ctx, a)
	if err == nil {
		t.Fatal("expected conflict on idempotent re-run, got nil")
	}
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}

	tracks, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tracks) != 1 {
		t.Errorf("expected exactly 1 track after re-run, got %d", len(tracks))
	}
}

func TestRepository_WALReopen_PreservesFingerprint(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "wal.db")
	ctx := context.Background()

	modTime := time.Unix(1_700_000_000, 987_654_321)
	a := AnalyzedTrack{
		Track:       model.Track{ID: "wal-1", Title: "WAL Track", FilePath: "/music/wal.mp3"},
		Fingerprint: "wal-fingerprint",
		Size:        8192,
		ModTime:     modTime,
	}

	repo1, err := NewRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create first repository: %v", err)
	}
	if err := repo1.SaveAnalyzed(ctx, a); err != nil {
		t.Fatalf("first SaveAnalyzed failed: %v", err)
	}
	if err := repo1.Close(); err != nil {
		t.Fatalf("close first repository failed: %v", err)
	}

	repo2, err := NewRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen repository: %v", err)
	}
	defer func() { _ = repo2.Close() }()

	stored, err := repo2.FindByPath(ctx, "/music/wal.mp3")
	if err != nil {
		t.Fatalf("FindByPath after reopen failed: %v", err)
	}
	if stored.Fingerprint != "wal-fingerprint" {
		t.Errorf("fingerprint: expected %q, got %q", "wal-fingerprint", stored.Fingerprint)
	}
	if stored.Size != 8192 {
		t.Errorf("size: expected 8192, got %d", stored.Size)
	}
	if !stored.ModTime.Equal(modTime) {
		t.Errorf("mod_time: expected %v, got %v", modTime, stored.ModTime)
	}
}
