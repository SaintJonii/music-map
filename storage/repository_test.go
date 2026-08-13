package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

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
