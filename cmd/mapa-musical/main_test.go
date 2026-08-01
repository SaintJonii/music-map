package main

import (
	"context"
	"testing"

	"github.com/SaintJonii/music-map/storage"
)

func TestPipeline_WAVFixture(t *testing.T) {
	repo, err := storage.NewRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	p := newPipeline(repo)
	ctx := context.Background()

	// Process the 16-bit stereo WAV fixture.
	track, features, err := p.processFile(ctx, "../../testdata/wav/stereo_44100_16bit.wav")
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Verify track metadata.
	if track.FilePath != "../../testdata/wav/stereo_44100_16bit.wav" {
		t.Errorf("expected file path, got %q", track.FilePath)
	}
	if track.Duration <= 0 {
		t.Errorf("expected positive duration, got %.2fs", track.Duration)
	}

	// Verify features were extracted (non-zero values expected for real audio).
	if features.BPM <= 0 {
		t.Errorf("expected positive BPM, got %.1f", features.BPM)
	}
	if features.Key == "" {
		t.Error("expected non-empty Key")
	}
	if features.Energy <= 0 {
		t.Errorf("expected positive Energy, got %.3f", features.Energy)
	}
	if features.Danceability < 0 || features.Danceability > 1 {
		t.Errorf("expected Danceability in [0,1], got %.3f", features.Danceability)
	}
	if features.Acousticness < 0 || features.Acousticness > 1 {
		t.Errorf("expected Acousticness in [0,1], got %.3f", features.Acousticness)
	}
	if features.SpectralCentroid <= 0 {
		t.Errorf("expected positive SpectralCentroid, got %.1f Hz", features.SpectralCentroid)
	}
	if features.ZCR <= 0 {
		t.Errorf("expected positive ZCR, got %.4f", features.ZCR)
	}

	// Check Chroma has values.
	var chromaSum float64
	for _, v := range features.Chroma {
		chromaSum += v
	}
	if chromaSum <= 0 {
		t.Errorf("expected positive chroma values, got sum=%.4f", chromaSum)
	}

	// Check MFCCs have values.
	var mfccSum float64
	for _, v := range features.MFCCs {
		mfccSum += v
	}
	// MFCC0 is typically large, others near zero.
	if mfccSum == 0 {
		t.Errorf("expected non-zero MFCC values, got sum=%.4f", mfccSum)
	}

	// Verify persistence: retrieve from DB.
	retrievedTrack, retrievedFeatures, err := repo.GetByID(ctx, track.ID)
	if err != nil {
		t.Fatalf("GetByID after save failed: %v", err)
	}

	if retrievedTrack.Duration != track.Duration {
		t.Errorf("persisted duration mismatch: expected %.4f, got %.4f", track.Duration, retrievedTrack.Duration)
	}
	if retrievedFeatures.BPM != features.BPM {
		t.Errorf("persisted BPM mismatch: expected %.1f, got %.1f", features.BPM, retrievedFeatures.BPM)
	}
	if retrievedFeatures.Key != features.Key {
		t.Errorf("persisted Key mismatch: expected %q, got %q", features.Key, retrievedFeatures.Key)
	}
	if retrievedFeatures.Energy != features.Energy {
		t.Errorf("persisted Energy mismatch: expected %.3f, got %.3f", features.Energy, retrievedFeatures.Energy)
	}

	t.Logf("Track ID: %s", retrievedTrack.ID)
	t.Logf("Duration: %.2fs", retrievedTrack.Duration)
	t.Logf("BPM: %.1f", retrievedFeatures.BPM)
	t.Logf("Key: %s", retrievedFeatures.Key)
	t.Logf("Energy: %.3f", retrievedFeatures.Energy)
	t.Logf("Danceability: %.3f", retrievedFeatures.Danceability)
	t.Logf("Acousticness: %.3f", retrievedFeatures.Acousticness)
	t.Logf("ZCR: %.4f", retrievedFeatures.ZCR)
	t.Logf("Spectral Centroid: %.1f Hz", retrievedFeatures.SpectralCentroid)
}

func TestPipeline_NonExistentFile(t *testing.T) {
	repo, err := storage.NewRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	defer repo.Close()

	p := newPipeline(repo)
	ctx := context.Background()

	_, _, err = p.processFile(ctx, "nonexistent.wav")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}
