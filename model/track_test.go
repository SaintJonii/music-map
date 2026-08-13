package model

import (
	"encoding/json"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Track
// ---------------------------------------------------------------------------

func TestTrack_HoldsAllFields(t *testing.T) {
	track := Track{
		ID:          "track-001",
		Title:       "Bohemian Rhapsody",
		Artist:      "Queen",
		Album:       "A Night at the Opera",
		AlbumArtist: "Queen",
		Genre:       "Rock",
		Year:        1975,
		TrackNumber: 11,
		Duration:    354.5,
		Format:      "flac",
		BitRate:     960,
		ISRC:        "GBUM71029604",
		FilePath:    "/music/queen/bohemian.flac",
	}

	if track.ID != "track-001" {
		t.Errorf("ID = %q, want %q", track.ID, "track-001")
	}
	if track.Title != "Bohemian Rhapsody" {
		t.Errorf("Title = %q, want %q", track.Title, "Bohemian Rhapsody")
	}
	if track.Artist != "Queen" {
		t.Errorf("Artist = %q, want %q", track.Artist, "Queen")
	}
	if track.Album != "A Night at the Opera" {
		t.Errorf("Album = %q, want %q", track.Album, "A Night at the Opera")
	}
	if track.AlbumArtist != "Queen" {
		t.Errorf("AlbumArtist = %q, want %q", track.AlbumArtist, "Queen")
	}
	if track.Genre != "Rock" {
		t.Errorf("Genre = %q, want %q", track.Genre, "Rock")
	}
	if track.Year != 1975 {
		t.Errorf("Year = %d, want %d", track.Year, 1975)
	}
	if track.TrackNumber != 11 {
		t.Errorf("TrackNumber = %d, want %d", track.TrackNumber, 11)
	}
	if track.Duration != 354.5 {
		t.Errorf("Duration = %f, want %f", track.Duration, 354.5)
	}
	if track.Format != "flac" {
		t.Errorf("Format = %q, want %q", track.Format, "flac")
	}
	if track.BitRate != 960 {
		t.Errorf("BitRate = %d, want %d", track.BitRate, 960)
	}
	if track.ISRC != "GBUM71029604" {
		t.Errorf("ISRC = %q, want %q", track.ISRC, "GBUM71029604")
	}
	if track.FilePath != "/music/queen/bohemian.flac" {
		t.Errorf("FilePath = %q, want %q", track.FilePath, "/music/queen/bohemian.flac")
	}
}

func TestNewTrack_DefaultsEmptyTitle(t *testing.T) {
	track := NewTrack(Track{
		Title: "",
	})
	if track.Title != "Unknown" {
		t.Errorf("Title = %q, want %q", track.Title, "Unknown")
	}
}

func TestNewTrack_PreservesNonEmptyTitle(t *testing.T) {
	track := NewTrack(Track{
		Title: "Daydreaming",
	})
	if track.Title != "Daydreaming" {
		t.Errorf("Title = %q, want %q", track.Title, "Daydreaming")
	}
}

func TestTrack_RoundTripJSON(t *testing.T) {
	track := Track{
		ID:          "track-001",
		Title:       "Bohemian Rhapsody",
		Artist:      "Queen",
		Album:       "A Night at the Opera",
		AlbumArtist: "Queen",
		Genre:       "Rock",
		Year:        1975,
		TrackNumber: 11,
		Duration:    354.5,
		Format:      "flac",
		BitRate:     960,
		ISRC:        "GBUM71029604",
		FilePath:    "/music/queen/bohemian.flac",
	}

	b, err := json.Marshal(track)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Track
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != track.ID {
		t.Errorf("round-trip ID = %q, want %q", decoded.ID, track.ID)
	}
	if decoded.Title != track.Title {
		t.Errorf("round-trip Title = %q, want %q", decoded.Title, track.Title)
	}
	if decoded.Artist != track.Artist {
		t.Errorf("round-trip Artist = %q, want %q", decoded.Artist, track.Artist)
	}
	if decoded.Album != track.Album {
		t.Errorf("round-trip Album = %q, want %q", decoded.Album, track.Album)
	}
	if decoded.AlbumArtist != track.AlbumArtist {
		t.Errorf("round-trip AlbumArtist = %q, want %q", decoded.AlbumArtist, track.AlbumArtist)
	}
	if decoded.Genre != track.Genre {
		t.Errorf("round-trip Genre = %q, want %q", decoded.Genre, track.Genre)
	}
	if decoded.Year != track.Year {
		t.Errorf("round-trip Year = %d, want %d", decoded.Year, track.Year)
	}
	if decoded.TrackNumber != track.TrackNumber {
		t.Errorf("round-trip TrackNumber = %d, want %d", decoded.TrackNumber, track.TrackNumber)
	}
	if decoded.Duration != track.Duration {
		t.Errorf("round-trip Duration = %f, want %f", decoded.Duration, track.Duration)
	}
	if decoded.Format != track.Format {
		t.Errorf("round-trip Format = %q, want %q", decoded.Format, track.Format)
	}
	if decoded.BitRate != track.BitRate {
		t.Errorf("round-trip BitRate = %d, want %d", decoded.BitRate, track.BitRate)
	}
	if decoded.ISRC != track.ISRC {
		t.Errorf("round-trip ISRC = %q, want %q", decoded.ISRC, track.ISRC)
	}
	if decoded.FilePath != track.FilePath {
		t.Errorf("round-trip FilePath = %q, want %q", decoded.FilePath, track.FilePath)
	}
}

func TestTrack_Unmarshal_TypeMismatch_ReturnsError(t *testing.T) {
	// Sending a string for the numeric Duration field
	badJSON := `{"duration": "not-a-number"}`
	var track Track
	err := json.Unmarshal([]byte(badJSON), &track)
	if err == nil {
		t.Error("expected unmarshal error for type mismatch, got nil")
	}
}

func TestTrack_Unmarshal_MalformedJSON_ReturnsError(t *testing.T) {
	badJSON := `{broken`
	var track Track
	err := json.Unmarshal([]byte(badJSON), &track)
	if err == nil {
		t.Error("expected unmarshal error for malformed JSON, got nil")
	}
}

// ---------------------------------------------------------------------------
// TrackFeatures
// ---------------------------------------------------------------------------

func TestTrackFeatures_RoundTripJSON(t *testing.T) {
	tf := TrackFeatures{
		BPM:              120.5,
		Key:              "A minor",
		Energy:           0.85,
		Danceability:     0.72,
		Acousticness:     0.15,
		SpectralCentroid: 1044.2,
		Chroma:           [12]float64{1.0, 0.0, 0.5, 0.0, 0.0, 0.3, 0.0, 0.0, 0.0, 0.0, 0.1, 0.0},
		MFCCs:            [13]float64{12.3, 4.1, -2.0, 0.5, 0.3, 0.1, 0.0, -0.1, -0.2, -0.1, 0.0, 0.1, 0.0},
		ZCR:              0.15,
	}

	b, err := json.Marshal(tf)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded TrackFeatures
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.BPM != tf.BPM {
		t.Errorf("BPM = %f, want %f", decoded.BPM, tf.BPM)
	}
	if decoded.Key != tf.Key {
		t.Errorf("Key = %q, want %q", decoded.Key, tf.Key)
	}
	if decoded.Energy != tf.Energy {
		t.Errorf("Energy = %f, want %f", decoded.Energy, tf.Energy)
	}
	if decoded.Danceability != tf.Danceability {
		t.Errorf("Danceability = %f, want %f", decoded.Danceability, tf.Danceability)
	}
	if decoded.Acousticness != tf.Acousticness {
		t.Errorf("Acousticness = %f, want %f", decoded.Acousticness, tf.Acousticness)
	}
	if decoded.SpectralCentroid != tf.SpectralCentroid {
		t.Errorf("SpectralCentroid = %f, want %f", decoded.SpectralCentroid, tf.SpectralCentroid)
	}
	if decoded.Chroma != tf.Chroma {
		t.Errorf("Chroma = %v, want %v", decoded.Chroma, tf.Chroma)
	}
	if decoded.MFCCs != tf.MFCCs {
		t.Errorf("MFCCs = %v, want %v", decoded.MFCCs, tf.MFCCs)
	}
	if decoded.ZCR != tf.ZCR {
		t.Errorf("ZCR = %f, want %f", decoded.ZCR, tf.ZCR)
	}
}

func TestTrackFeatures_MissingOptional_MarshalAsZero(t *testing.T) {
	tf := TrackFeatures{
		// Key empty, BPM 0 — intentional
		Energy: 0.5,
	}

	b, err := json.Marshal(tf)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal into raw: %v", err)
	}

	// Key must be present and empty
	if k, ok := raw["key"]; !ok {
		t.Error("key field missing from JSON output")
	} else if k != "" {
		t.Errorf("key = %v, want empty string", k)
	}

	// BPM must be present as 0
	if b, ok := raw["bpm"]; !ok {
		t.Error("bpm field missing from JSON output")
	} else if b != float64(0) {
		t.Errorf("bpm = %v, want 0", b)
	}
}

// ---------------------------------------------------------------------------
// Collection
// ---------------------------------------------------------------------------

func TestCollection_HoldsMultipleTracks(t *testing.T) {
	tracks := []Track{
		{ID: "t1", Title: "One"},
		{ID: "t2", Title: "Two"},
		{ID: "t3", Title: "Three"},
	}

	now := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	coll := Collection{
		Tracks:    tracks,
		Name:      "My Playlist",
		CreatedAt: now,
	}

	if len(coll.Tracks) != 3 {
		t.Fatalf("len(Tracks) = %d, want 3", len(coll.Tracks))
	}
	if coll.Tracks[0].ID != "t1" {
		t.Errorf("Tracks[0].ID = %q, want %q", coll.Tracks[0].ID, "t1")
	}
	if coll.Tracks[1].ID != "t2" {
		t.Errorf("Tracks[1].ID = %q, want %q", coll.Tracks[1].ID, "t2")
	}
	if coll.Tracks[2].ID != "t3" {
		t.Errorf("Tracks[2].ID = %q, want %q", coll.Tracks[2].ID, "t3")
	}
	if coll.Name != "My Playlist" {
		t.Errorf("Name = %q, want %q", coll.Name, "My Playlist")
	}
	if !coll.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", coll.CreatedAt, now)
	}
}

func TestCollection_Empty_SerializesToEmptyArray(t *testing.T) {
	coll := Collection{
		Tracks: []Track{},
		Name:   "Empty List",
	}

	b, err := json.Marshal(coll)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal into raw: %v", err)
	}

	tracks, ok := raw["tracks"]
	if !ok {
		t.Fatal("tracks field missing from JSON output")
	}

	// Must be [], not null
	tracksSlice, ok := tracks.([]interface{})
	if !ok {
		t.Fatalf("tracks is not an array, got %T", tracks)
	}
	if len(tracksSlice) != 0 {
		t.Errorf("tracks array length = %d, want 0", len(tracksSlice))
	}
}
