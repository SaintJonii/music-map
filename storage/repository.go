package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/SaintJonii/music-map/model"
	_ "modernc.org/sqlite"
)

// Sentinel errors for the storage layer.
var (
	ErrNotFound = errors.New("storage: track not found")
	ErrConflict = errors.New("storage: track already exists")
)

// Repository defines the persistence interface for Tracks and TrackFeatures.
type Repository interface {
	Save(ctx context.Context, track model.Track, features model.TrackFeatures) error
	GetByID(ctx context.Context, id string) (model.Track, model.TrackFeatures, error)
	List(ctx context.Context) ([]model.Track, error)
	Close() error
}

// sqliteRepo implements Repository backed by SQLite.
type sqliteRepo struct {
	db *sql.DB
}

// NewRepository creates a new SQLite-backed Repository.
// If the path is ":memory:", an in-memory database is used.
func NewRepository(path string) (Repository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("storage: open %s: %w", path, err)
	}

	repo := &sqliteRepo{db: db}
	if err := repo.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("storage: migrate: %w", err)
	}

	return repo, nil
}

// migrate creates the schema if it doesn't exist.
func (r *sqliteRepo) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tracks (
		id           TEXT PRIMARY KEY,
		title        TEXT NOT NULL DEFAULT '',
		artist       TEXT NOT NULL DEFAULT '',
		album        TEXT NOT NULL DEFAULT '',
		album_artist TEXT NOT NULL DEFAULT '',
		genre        TEXT NOT NULL DEFAULT '',
		year         INTEGER NOT NULL DEFAULT 0,
		track_number INTEGER NOT NULL DEFAULT 0,
		duration     REAL    NOT NULL DEFAULT 0.0,
		format       TEXT    NOT NULL DEFAULT '',
		bit_rate     INTEGER NOT NULL DEFAULT 0,
		isrc         TEXT    NOT NULL DEFAULT '',
		file_path    TEXT    NOT NULL DEFAULT '',
		created_at   TEXT    NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS track_features (
		track_id          TEXT PRIMARY KEY REFERENCES tracks(id) ON DELETE CASCADE,
		bpm               REAL    NOT NULL DEFAULT 0.0,
		key               TEXT    NOT NULL DEFAULT '',
		energy            REAL    NOT NULL DEFAULT 0.0,
		danceability      REAL    NOT NULL DEFAULT 0.0,
		acousticness      REAL    NOT NULL DEFAULT 0.0,
		spectral_centroid REAL    NOT NULL DEFAULT 0.0,
		chroma_c          REAL    NOT NULL DEFAULT 0.0,
		chroma_c_sharp    REAL    NOT NULL DEFAULT 0.0,
		chroma_d          REAL    NOT NULL DEFAULT 0.0,
		chroma_d_sharp    REAL    NOT NULL DEFAULT 0.0,
		chroma_e          REAL    NOT NULL DEFAULT 0.0,
		chroma_f          REAL    NOT NULL DEFAULT 0.0,
		chroma_f_sharp    REAL    NOT NULL DEFAULT 0.0,
		chroma_g          REAL    NOT NULL DEFAULT 0.0,
		chroma_g_sharp    REAL    NOT NULL DEFAULT 0.0,
		chroma_a          REAL    NOT NULL DEFAULT 0.0,
		chroma_a_sharp    REAL    NOT NULL DEFAULT 0.0,
		chroma_b          REAL    NOT NULL DEFAULT 0.0,
		mfcc_0            REAL    NOT NULL DEFAULT 0.0,
		mfcc_1            REAL    NOT NULL DEFAULT 0.0,
		mfcc_2            REAL    NOT NULL DEFAULT 0.0,
		mfcc_3            REAL    NOT NULL DEFAULT 0.0,
		mfcc_4            REAL    NOT NULL DEFAULT 0.0,
		mfcc_5            REAL    NOT NULL DEFAULT 0.0,
		mfcc_6            REAL    NOT NULL DEFAULT 0.0,
		mfcc_7            REAL    NOT NULL DEFAULT 0.0,
		mfcc_8            REAL    NOT NULL DEFAULT 0.0,
		mfcc_9            REAL    NOT NULL DEFAULT 0.0,
		mfcc_10           REAL    NOT NULL DEFAULT 0.0,
		mfcc_11           REAL    NOT NULL DEFAULT 0.0,
		mfcc_12           REAL    NOT NULL DEFAULT 0.0,
		zcr               REAL    NOT NULL DEFAULT 0.0
	);
	`
	_, err := r.db.Exec(schema)
	return err
}

// Save persists a Track and its TrackFeatures in a single transaction.
func (r *sqliteRepo) Save(ctx context.Context, track model.Track, features model.TrackFeatures) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("storage: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Insert track.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO tracks (id, title, artist, album, album_artist, genre, year,
		                    track_number, duration, format, bit_rate, isrc, file_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		track.ID, track.Title, track.Artist, track.Album, track.AlbumArtist,
		track.Genre, track.Year, track.TrackNumber, track.Duration,
		track.Format, track.BitRate, track.ISRC, track.FilePath,
	)
	if err != nil {
		if isUniqueConstraint(err) {
			return fmt.Errorf("storage: save track %s: %w", track.ID, ErrConflict)
		}
		return fmt.Errorf("storage: insert track: %w", err)
	}

	// Insert features.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO track_features (track_id, bpm, key, energy, danceability, acousticness,
		                            spectral_centroid,
		                            chroma_c, chroma_c_sharp, chroma_d, chroma_d_sharp,
		                            chroma_e, chroma_f, chroma_f_sharp, chroma_g, chroma_g_sharp,
		                            chroma_a, chroma_a_sharp, chroma_b,
		                            mfcc_0, mfcc_1, mfcc_2, mfcc_3, mfcc_4, mfcc_5, mfcc_6,
		                            mfcc_7, mfcc_8, mfcc_9, mfcc_10, mfcc_11, mfcc_12,
		                            zcr)
		VALUES (?, ?, ?, ?, ?, ?, ?,
		        ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		        ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		        ?)
	`,
		track.ID,
		features.BPM, features.Key, features.Energy, features.Danceability, features.Acousticness,
		features.SpectralCentroid,
		features.Chroma[0], features.Chroma[1], features.Chroma[2], features.Chroma[3],
		features.Chroma[4], features.Chroma[5], features.Chroma[6], features.Chroma[7],
		features.Chroma[8], features.Chroma[9], features.Chroma[10], features.Chroma[11],
		features.MFCCs[0], features.MFCCs[1], features.MFCCs[2], features.MFCCs[3],
		features.MFCCs[4], features.MFCCs[5], features.MFCCs[6], features.MFCCs[7],
		features.MFCCs[8], features.MFCCs[9], features.MFCCs[10], features.MFCCs[11],
		features.MFCCs[12],
		features.ZCR,
	)
	if err != nil {
		return fmt.Errorf("storage: insert features: %w", err)
	}

	return tx.Commit()
}

// GetByID retrieves a Track and its TrackFeatures by ID.
func (r *sqliteRepo) GetByID(ctx context.Context, id string) (model.Track, model.TrackFeatures, error) {
	var track model.Track
	var features model.TrackFeatures

	err := r.db.QueryRowContext(ctx, `
		SELECT t.id, t.title, t.artist, t.album, t.album_artist, t.genre,
		       t.year, t.track_number, t.duration, t.format, t.bit_rate,
		       t.isrc, t.file_path,
		       f.bpm, f.key, f.energy, f.danceability, f.acousticness,
		       f.spectral_centroid,
		       f.chroma_c, f.chroma_c_sharp, f.chroma_d, f.chroma_d_sharp,
		       f.chroma_e, f.chroma_f, f.chroma_f_sharp, f.chroma_g, f.chroma_g_sharp,
		       f.chroma_a, f.chroma_a_sharp, f.chroma_b,
		       f.mfcc_0, f.mfcc_1, f.mfcc_2, f.mfcc_3, f.mfcc_4, f.mfcc_5, f.mfcc_6,
		       f.mfcc_7, f.mfcc_8, f.mfcc_9, f.mfcc_10, f.mfcc_11, f.mfcc_12,
		       f.zcr
		FROM tracks t
		LEFT JOIN track_features f ON t.id = f.track_id
		WHERE t.id = ?
	`, id).Scan(
		&track.ID, &track.Title, &track.Artist, &track.Album, &track.AlbumArtist,
		&track.Genre, &track.Year, &track.TrackNumber, &track.Duration,
		&track.Format, &track.BitRate, &track.ISRC, &track.FilePath,
		&features.BPM, &features.Key, &features.Energy, &features.Danceability,
		&features.Acousticness, &features.SpectralCentroid,
		&features.Chroma[0], &features.Chroma[1], &features.Chroma[2], &features.Chroma[3],
		&features.Chroma[4], &features.Chroma[5], &features.Chroma[6], &features.Chroma[7],
		&features.Chroma[8], &features.Chroma[9], &features.Chroma[10], &features.Chroma[11],
		&features.MFCCs[0], &features.MFCCs[1], &features.MFCCs[2], &features.MFCCs[3],
		&features.MFCCs[4], &features.MFCCs[5], &features.MFCCs[6], &features.MFCCs[7],
		&features.MFCCs[8], &features.MFCCs[9], &features.MFCCs[10], &features.MFCCs[11],
		&features.MFCCs[12],
		&features.ZCR,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Track{}, model.TrackFeatures{}, ErrNotFound
		}
		return model.Track{}, model.TrackFeatures{}, fmt.Errorf("storage: get track %s: %w", id, err)
	}

	return track, features, nil
}

// List returns all tracks ordered by insertion time ascending.
func (r *sqliteRepo) List(ctx context.Context) ([]model.Track, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, artist, album, album_artist, genre, year,
		       track_number, duration, format, bit_rate, isrc, file_path
		FROM tracks
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("storage: list tracks: %w", err)
	}
	defer rows.Close()

	var tracks []model.Track
	for rows.Next() {
		var t model.Track
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Artist, &t.Album, &t.AlbumArtist,
			&t.Genre, &t.Year, &t.TrackNumber, &t.Duration,
			&t.Format, &t.BitRate, &t.ISRC, &t.FilePath,
		); err != nil {
			return nil, fmt.Errorf("storage: scan track: %w", err)
		}
		tracks = append(tracks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: rows iteration: %w", err)
	}

	// Return non-nil empty slice.
	if tracks == nil {
		tracks = []model.Track{}
	}
	return tracks, nil
}

// Close closes the underlying database connection.
func (r *sqliteRepo) Close() error {
	return r.db.Close()
}

// isUniqueConstraint checks if an error is a SQLite UNIQUE constraint violation.
func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
