package model

import "time"

// Track represents an audio track with metadata.
type Track struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Album       string  `json:"album"`
	AlbumArtist string  `json:"album_artist"`
	Genre       string  `json:"genre"`
	Year        int     `json:"year"`
	TrackNumber int     `json:"track_number"`
	Duration    float64 `json:"duration"`
	Format      string  `json:"format"`
	BitRate     int     `json:"bit_rate"`
	ISRC        string  `json:"isrc"`
	FilePath    string  `json:"file_path"`
}

// NewTrack creates a Track, applying defaults for empty fields.
func NewTrack(t Track) Track {
	if t.Title == "" {
		t.Title = "Unknown"
	}
	return t
}

// TrackFeatures holds extracted acoustic features for a track.
type TrackFeatures struct {
	BPM              float64     `json:"bpm"`
	Key              string      `json:"key"`
	Energy           float64     `json:"energy"`
	Danceability     float64     `json:"danceability"`
	Acousticness     float64     `json:"acousticness"`
	SpectralCentroid float64     `json:"spectral_centroid"`
	Chroma           [12]float64 `json:"chroma"`
	MFCCs            [13]float64 `json:"mfccs"`
	ZCR              float64     `json:"zcr"`
}

// Collection aggregates tracks into a named collection.
type Collection struct {
	Tracks    []Track   `json:"tracks"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
