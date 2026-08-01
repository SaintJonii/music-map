package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/SaintJonii/music-map/audio"
	"github.com/SaintJonii/music-map/metadata"
	"github.com/SaintJonii/music-map/model"
	"github.com/SaintJonii/music-map/storage"
)

// pipeline orchestrates the full audio processing flow.
type pipeline struct {
	tagReader metadata.TagReader
	extractor audio.FeatureExtractor
	repo      storage.Repository
}

// newPipeline creates the default processing pipeline.
func newPipeline(repo storage.Repository) *pipeline {
	return &pipeline{
		tagReader: metadata.NewTagReader(),
		extractor: &audio.DefaultExtractor{},
		repo:      repo,
	}
}

// processFile runs the full pipeline on a single audio file.
func (p *pipeline) processFile(ctx context.Context, filePath string) (model.Track, model.TrackFeatures, error) {
	// 1. Open file.
	f, err := os.Open(filePath)
	if err != nil {
		return model.Track{}, model.TrackFeatures{}, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// 2. Read tags (tags-first strategy).
	track, err := p.tagReader.ReadTags(f)
	if err != nil {
		return model.Track{}, model.TrackFeatures{}, fmt.Errorf("read tags: %w", err)
	}
	track.FilePath = filePath

	// Set ID from file path if not present.
	if track.ID == "" {
		track.ID = fileID(filePath)
	}

	// Reset file position for decoding.
	if _, err := f.Seek(0, 0); err != nil {
		return model.Track{}, model.TrackFeatures{}, fmt.Errorf("seek: %w", err)
	}

	// 3. Detect format and decode.
	decoder, err := audio.DetectFormat(f)
	if err != nil {
		return model.Track{}, model.TrackFeatures{}, fmt.Errorf("detect format: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return model.Track{}, model.TrackFeatures{}, fmt.Errorf("seek: %w", err)
	}
	samples, sampleRate, channels, err := decoder.Decode(f)
	if err != nil {
		return model.Track{}, model.TrackFeatures{}, fmt.Errorf("decode: %w", err)
	}

	track.Duration = float64(len(samples)) / float64(sampleRate*channels)
	track.Format, _ = detectFormatName(samples, sampleRate, channels)

	// 4. Extract features.
	energy := p.extractor.ExtractRMS(samples)
	zcr := p.extractor.ExtractZCR(samples, sampleRate)
	centroid, _ := p.extractor.ExtractSpectralCentroid(samples, sampleRate)
	bpm, _ := p.extractor.ExtractBPM(samples, sampleRate)
	chroma, _ := p.extractor.ExtractChroma(samples, sampleRate)
	mfccs, _ := p.extractor.ExtractMFCCs(samples)
	key := p.extractor.ExtractKey(chroma)
	danceability := p.extractor.ExtractDanceability(bpm, energy, zcr)
	acousticness := p.extractor.ExtractAcousticness(centroid, energy)

	features := model.TrackFeatures{
		BPM:              bpm,
		Key:              key,
		Energy:           energy,
		Danceability:     danceability,
		Acousticness:     acousticness,
		SpectralCentroid: centroid,
		Chroma:           chroma,
		MFCCs:            mfccs,
		ZCR:              zcr,
	}

	// 5. Persist.
	if err := p.repo.Save(ctx, track, features); err != nil {
		return model.Track{}, model.TrackFeatures{}, fmt.Errorf("persist: %w", err)
	}

	return track, features, nil
}

// detectFormatName detects the audio format name.
func detectFormatName(_ []float64, _ int, _ int) (string, error) {
	// Format detection would normally come from the decoder.
	// For now, return empty — the format is set by the CLI or from file extension.
	return "", nil
}

// fileID generates a deterministic ID from a file path.
func fileID(path string) string {
	h := sha256.Sum256([]byte(path))
	return fmt.Sprintf("%x", h[:16])
}

func main() {
	ctx := context.Background()

	// Use in-memory DB for simple CLI runs.
	repo, err := storage.NewRepository(":memory:")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize storage: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	p := newPipeline(repo)

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <audio-file>\n", os.Args[0])
		os.Exit(1)
	}

	filePath := os.Args[1]
	track, features, err := p.processFile(ctx, filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", filePath, err)
		os.Exit(1)
	}

	// Print results.
	fmt.Printf("Track: %s\n", track.Title)
	fmt.Printf("Artist: %s\n", track.Artist)
	fmt.Printf("Duration: %.2fs\n", track.Duration)
	fmt.Printf("BPM: %.1f\n", features.BPM)
	fmt.Printf("Key: %s\n", features.Key)
	fmt.Printf("Energy: %.3f\n", features.Energy)
	fmt.Printf("Danceability: %.3f\n", features.Danceability)
	fmt.Printf("Acousticness: %.3f\n", features.Acousticness)
	fmt.Printf("ZCR: %.4f\n", features.ZCR)
	fmt.Printf("Spectral Centroid: %.1f Hz\n", features.SpectralCentroid)
	fmt.Println("Processing complete.")
}
