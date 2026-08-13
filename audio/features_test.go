package audio

import (
	"math"
	"strings"
	"testing"
)

// --- Test signal generators ---

// generateSine creates a sine wave at the given frequency with the given peak amplitude.
// sampleRate and durationSec control the size of the output buffer.
// Samples are in [-amplitude, +amplitude].
func generateSine(freqHz float64, amplitude float64, sampleRate int, durationSec float64) []float64 {
	n := int(float64(sampleRate) * durationSec)
	samples := make([]float64, n)
	for i := range n {
		t := float64(i) / float64(sampleRate)
		samples[i] = amplitude * math.Sin(2.0*math.Pi*freqHz*t)
	}
	return samples
}

// --- 3.1 RED: RMS Energy ---

func TestExtractRMS_SineHalfAmplitude(t *testing.T) {
	// Generate a sine whose RMS ≈ 0.5.
	// For a sine wave: RMS = peak / √2 → peak = RMS * √2 = 0.5 * √2 ≈ 0.7071.
	// Per spec: "sine 0.5→0.5±0.01" — amplitude 0.5 in RMS terms.
	const (
		sampleRate    = 44100
		durationSec   = 1.0
		targetRMS     = 0.5
		peakForTarget = targetRMS * math.Sqrt2 // ≈ 0.7071
		tolerance     = 0.01
	)
	samples := generateSine(1000.0, peakForTarget, sampleRate, durationSec)

	ext := &DefaultExtractor{}
	got := ext.ExtractRMS(samples)

	if math.Abs(got-targetRMS) > tolerance {
		t.Errorf("ExtractRMS(sine) = %.6f, want ≈ %.6f (±%.4f)", got, targetRMS, tolerance)
	}
}

func TestExtractRMS_SilenceReturnsZero(t *testing.T) {
	samples := make([]float64, 44100)

	ext := &DefaultExtractor{}
	got := ext.ExtractRMS(samples)

	if got != 0.0 {
		t.Errorf("ExtractRMS(silence) = %.6f, want 0.0", got)
	}
}

func TestExtractRMS_QuarterAmplitude(t *testing.T) {
	// Triangulation: different amplitude ensures implementation is not hardcoded.
	// Sine with RMS-target 0.25 → peak = 0.25 * √2 ≈ 0.3536.
	const (
		sampleRate    = 44100
		durationSec   = 1.0
		targetRMS     = 0.25
		peakForTarget = targetRMS * math.Sqrt2
		tolerance     = 0.01
	)
	samples := generateSine(440.0, peakForTarget, sampleRate, durationSec)

	ext := &DefaultExtractor{}
	got := ext.ExtractRMS(samples)

	if math.Abs(got-targetRMS) > tolerance {
		t.Errorf("ExtractRMS(quarter-amplitude) = %.6f, want ≈ %.6f (±%.4f)", got, targetRMS, tolerance)
	}
}

func TestExtractRMS_EmptyInputReturnsZero(t *testing.T) {
	ext := &DefaultExtractor{}
	got := ext.ExtractRMS(nil)

	if got != 0.0 {
		t.Errorf("ExtractRMS(empty) = %.6f, want 0.0", got)
	}
}

// --- 3.1 RED: Zero-Crossing Rate ---

func TestExtractZCR_Sine440Hz(t *testing.T) {
	// 440 Hz sine at 44100 Hz sample rate, 1 second.
	// Each cycle crosses zero twice → 440 * 2 = 880 crossings/sec.
	// Per spec: "440 Hz→880±5%".
	const (
		sampleRate  = 44100
		durationSec = 1.0
		expectedZCR = 880.0
		tolerance   = 0.05 * expectedZCR // ±5%
	)
	samples := generateSine(440.0, 1.0, sampleRate, durationSec)

	ext := &DefaultExtractor{}
	got := ext.ExtractZCR(samples, sampleRate)

	if math.Abs(got-expectedZCR) > tolerance {
		t.Errorf("ExtractZCR(440Hz) = %.2f, want ≈ %.2f (±5%%)", got, expectedZCR)
	}
}

func TestExtractZCR_Sine220Hz(t *testing.T) {
	// Triangulation: different frequency — ZCR should be 2 * freq.
	const (
		sampleRate  = 44100
		durationSec = 1.0
		freqHz      = 220.0
		expectedZCR = 2.0 * freqHz // 440
		tolerance   = 0.05 * expectedZCR
	)
	samples := generateSine(freqHz, 1.0, sampleRate, durationSec)

	ext := &DefaultExtractor{}
	got := ext.ExtractZCR(samples, sampleRate)

	if math.Abs(got-expectedZCR) > tolerance {
		t.Errorf("ExtractZCR(220Hz) = %.2f, want ≈ %.2f (±5%%)", got, expectedZCR)
	}
}

func TestExtractZCR_ShortInputReturnsZero(t *testing.T) {
	// Less than 2 samples → no crossings possible.
	ext := &DefaultExtractor{}
	got := ext.ExtractZCR([]float64{0.5}, 44100)

	if got != 0.0 {
		t.Errorf("ExtractZCR(single sample) = %.2f, want 0.0", got)
	}
}

func TestExtractZCR_DCOffsetReturnsZero(t *testing.T) {
	// Pure DC signal — all values the same, no zero crossings.
	const sampleRate = 44100
	samples := make([]float64, sampleRate)
	for i := range samples {
		samples[i] = 0.5
	}

	ext := &DefaultExtractor{}
	got := ext.ExtractZCR(samples, sampleRate)

	if got != 0.0 {
		t.Errorf("ExtractZCR(DC offset) = %.2f, want 0.0", got)
	}
}

// --- 3.1 RED: Spectral Centroid ---

func TestExtractSpectralCentroid_Sine100Hz(t *testing.T) {
	// A 100 Hz sine should have its spectral centroid near 100 Hz.
	// Use 1 second at 44100 Hz for decent FFT resolution.
	const (
		sampleRate  = 44100
		durationSec = 1.0
		freqHz      = 100.0
		expectedHz  = 100.0
		toleranceHz = 10.0
	)
	samples := generateSine(freqHz, 0.5, sampleRate, durationSec)

	ext := &DefaultExtractor{}
	got, err := ext.ExtractSpectralCentroid(samples, sampleRate)
	if err != nil {
		t.Fatalf("ExtractSpectralCentroid() returned unexpected error: %v", err)
	}

	if math.Abs(got-expectedHz) > toleranceHz {
		t.Errorf("ExtractSpectralCentroid(100Hz) = %.2f Hz, want ≈ %.2f Hz (±%.1f Hz)", got, expectedHz, toleranceHz)
	}
}

func TestExtractSpectralCentroid_SilenceReturnsZero(t *testing.T) {
	// Triangulation: silent signal should give centroid 0.
	samples := make([]float64, 1024)

	ext := &DefaultExtractor{}
	got, err := ext.ExtractSpectralCentroid(samples, 44100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != 0.0 {
		t.Errorf("ExtractSpectralCentroid(silence) = %.2f, want 0.0", got)
	}
}

func TestExtractSpectralCentroid_Sine200Hz(t *testing.T) {
	// Triangulation: different frequency to rule out hardcoded return.
	const (
		sampleRate  = 44100
		durationSec = 1.0
		freqHz      = 200.0
		expectedHz  = 200.0
		toleranceHz = 15.0
	)
	samples := generateSine(freqHz, 0.5, sampleRate, durationSec)

	ext := &DefaultExtractor{}
	got, err := ext.ExtractSpectralCentroid(samples, sampleRate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(got-expectedHz) > toleranceHz {
		t.Errorf("ExtractSpectralCentroid(200Hz) = %.2f Hz, want ≈ %.2f Hz (±%.1f Hz)", got, expectedHz, toleranceHz)
	}
}

// === 3.2 RED: BPM Detection ===

// generateClickTrack creates a click track at the given BPM.
// Each beat is a short impulse (dirac-like sample).
func generateClickTrack(bpm float64, sampleRate int, durationSec float64) []float64 {
	n := int(float64(sampleRate) * durationSec)
	samples := make([]float64, n)
	samplesPerBeat := float64(sampleRate) * 60.0 / bpm
	for i := range n {
		if i%int(samplesPerBeat) < 5 {
			samples[i] = 1.0
		}
	}
	return samples
}

func TestExtractBPM_ClickTrack120BPM(t *testing.T) {
	const (
		sampleRate  = 44100
		durationSec = 10.0
		bpm         = 120.0
		tolerance   = 0.05 * bpm // ±5%
	)
	samples := generateClickTrack(bpm, sampleRate, durationSec)

	ext := &DefaultExtractor{}
	got, err := ext.ExtractBPM(samples, sampleRate)
	if err != nil {
		t.Fatalf("ExtractBPM() returned unexpected error: %v", err)
	}

	if math.Abs(got-bpm) > tolerance {
		t.Errorf("ExtractBPM(120BPM click) = %.2f, want ≈ %.2f (±5%%)", got, bpm)
	}
}

func TestExtractBPM_ShortInputReturnsError(t *testing.T) {
	// Less than 1 second of audio should return an error.
	samples := generateSine(440.0, 0.5, 44100, 0.5)

	ext := &DefaultExtractor{}
	_, err := ext.ExtractBPM(samples, 44100)
	if err == nil {
		t.Error("ExtractBPM(short input) should return an error")
	}
}

func TestExtractBPM_ClickTrack100BPM(t *testing.T) {
	// Triangulation: different tempo to rule out hardcoded response.
	const (
		sampleRate  = 44100
		durationSec = 10.0
		bpm         = 100.0
		tolerance   = 0.05 * bpm
	)
	samples := generateClickTrack(bpm, sampleRate, durationSec)

	ext := &DefaultExtractor{}
	got, err := ext.ExtractBPM(samples, sampleRate)
	if err != nil {
		t.Fatalf("ExtractBPM() returned unexpected error: %v", err)
	}

	if math.Abs(got-bpm) > tolerance {
		t.Errorf("ExtractBPM(100BPM click) = %.2f, want ≈ %.2f (±5%%)", got, bpm)
	}
}

// === 3.2 RED: Chroma Vector ===

// noteFrequency returns the frequency of a musical note given its MIDI number.
func noteFrequency(midi int) float64 {
	return 440.0 * math.Pow(2.0, float64(midi-69)/12.0)
}

// generateToneSum creates a signal by summing sine waves at given MIDI notes.
func generateToneSum(midiNotes []int, sampleRate int, durationSec float64) []float64 {
	n := int(float64(sampleRate) * durationSec)
	samples := make([]float64, n)
	for _, midi := range midiNotes {
		freq := noteFrequency(midi)
		for i := range n {
			t := float64(i) / float64(sampleRate)
			samples[i] += 0.3 * math.Sin(2.0*math.Pi*freq*t)
		}
	}
	return samples
}

func TestExtractChroma_CMajorTriad(t *testing.T) {
	// C-major triad: C4 (60), E4 (64), G4 (67).
	// Bins for C (0), E (4), G (7) should be highest.
	const sampleRate = 44100
	midiNotes := []int{60, 64, 67} // C4, E4, G4
	samples := generateToneSum(midiNotes, sampleRate, 2.0)

	ext := &DefaultExtractor{}
	chroma, err := ext.ExtractChroma(samples, sampleRate)
	if err != nil {
		t.Fatalf("ExtractChroma() returned unexpected error: %v", err)
	}

	// Verify total sum is ~1.0 (normalized).
	var sum float64
	for _, v := range chroma {
		sum += v
	}
	if math.Abs(sum-1.0) > 0.01 {
		t.Errorf("Chroma sum = %.4f, want 1.0", sum)
	}

	// C (0), E (4), G (7) should be the three highest bins.
	type indexedVal struct {
		idx int
		val float64
	}
	var bins []indexedVal
	for i, v := range chroma {
		bins = append(bins, indexedVal{i, v})
	}
	// Sort by value descending.
	for i := range bins {
		for j := i + 1; j < len(bins); j++ {
			if bins[j].val > bins[i].val {
				bins[i], bins[j] = bins[j], bins[i]
			}
		}
	}

	topThree := make(map[int]bool)
	for i := range 3 {
		topThree[bins[i].idx] = true
	}

	assertBinInTop(t, topThree, 0, "C")
	assertBinInTop(t, topThree, 4, "E")
	assertBinInTop(t, topThree, 7, "G")
}

func assertBinInTop(t *testing.T, topThree map[int]bool, bin int, name string) {
	t.Helper()
	if !topThree[bin] {
		t.Errorf("%s bin (%d) not in top 3", name, bin)
	}
}

func TestExtractChroma_MajorThird(t *testing.T) {
	// Triangulation: only C and E (perfect third) — bins 0 and 4 should be highest.
	const sampleRate = 44100
	midiNotes := []int{60, 64} // C4, E4
	samples := generateToneSum(midiNotes, sampleRate, 2.0)

	ext := &DefaultExtractor{}
	chroma, err := ext.ExtractChroma(samples, sampleRate)
	if err != nil {
		t.Fatalf("ExtractChroma() returned unexpected error: %v", err)
	}

	// Sort bins to find top 2.
	var bins []struct {
		idx int
		val float64
	}
	for i, v := range chroma {
		bins = append(bins, struct {
			idx int
			val float64
		}{i, v})
	}
	for i := range bins {
		for j := i + 1; j < len(bins); j++ {
			if bins[j].val > bins[i].val {
				bins[i], bins[j] = bins[j], bins[i]
			}
		}
	}

	topTwo := map[int]bool{bins[0].idx: true, bins[1].idx: true}
	if !topTwo[0] || !topTwo[4] {
		t.Errorf("C-major third: expected bins 0 (C) and 4 (E) to be top 2, got: %v", topTwo)
	}
}

// === 3.2 RED: MFCC Coefficients ===

func TestExtractMFCCs_ValidFrameReturns13Coeffs(t *testing.T) {
	// 1024-sample frame should return 13 MFCCs.
	frame := generateSine(440.0, 0.5, 1024, 1.0) // 1 second at 1024 Hz = 1024 samples

	ext := &DefaultExtractor{}
	mfccs, err := ext.ExtractMFCCs(frame)
	if err != nil {
		t.Fatalf("ExtractMFCCs() returned unexpected error: %v", err)
	}

	// Should return exactly 13 coefficients.
	if len(mfccs) != 13 {
		t.Errorf("ExtractMFCCs() returned %d coefficients, want 13", len(mfccs))
	}

	// MFCC-0 should represent overall energy (non-zero for non-silent input).
	if mfccs[0] == 0.0 {
		t.Error("MFCC-0 should be non-zero for non-silent input")
	}
}

func TestExtractMFCCs_EmptyInputReturnsZeros(t *testing.T) {
	ext := &DefaultExtractor{}
	mfccs, err := ext.ExtractMFCCs(nil)
	if err != nil {
		t.Fatalf("ExtractMFCCs() returned unexpected error: %v", err)
	}

	for i, v := range mfccs {
		if v != 0.0 {
			t.Errorf("MFCC[%d] = %.6f for empty input, want 0.0", i, v)
		}
	}
}

// === 3.2 RED: Key Detection ===

func TestExtractKey_AMinorTriad(t *testing.T) {
	// A-minor triad: A4 (69), C5 (72), E5 (76).
	const sampleRate = 44100
	midiNotes := []int{69, 72, 76} // A4, C5, E5
	samples := generateToneSum(midiNotes, sampleRate, 2.0)

	ext := &DefaultExtractor{}
	chroma, err := ext.ExtractChroma(samples, sampleRate)
	if err != nil {
		t.Fatalf("ExtractChroma() for key test: %v", err)
	}

	got := ext.ExtractKey(chroma)

	// Case-insensitive comparison for "A minor".
	if !strings.EqualFold(got, "A minor") && !strings.EqualFold(got, "a minor") {
		t.Errorf("ExtractKey(A-minor triad) = %q, want \"A minor\"", got)
	}
}

func TestExtractKey_CMajorTriad(t *testing.T) {
	// Triangulation: C-major triad should return "C major".
	const sampleRate = 44100
	midiNotes := []int{60, 64, 67} // C4, E4, G4
	samples := generateToneSum(midiNotes, sampleRate, 2.0)

	ext := &DefaultExtractor{}
	chroma, err := ext.ExtractChroma(samples, sampleRate)
	if err != nil {
		t.Fatalf("ExtractChroma() for key test: %v", err)
	}

	got := ext.ExtractKey(chroma)

	if !strings.EqualFold(got, "C major") {
		t.Errorf("ExtractKey(C-major triad) = %q, want \"C major\"", got)
	}
}

// === 3.2 RED: Danceability ===

func TestExtractDanceability_HighEnergy(t *testing.T) {
	// BPM=120, energy=0.9, ZCR=0.3 → should be > 0.7.
	ext := &DefaultExtractor{}
	got := ext.ExtractDanceability(120.0, 0.9, 0.3)

	if got <= 0.7 {
		t.Errorf("ExtractDanceability(high) = %.4f, want > 0.7", got)
	}
	if got > 1.0 || got < 0.0 {
		t.Errorf("ExtractDanceability(high) = %.4f, out of [0,1] range", got)
	}
}

func TestExtractDanceability_SlowAmbient(t *testing.T) {
	// BPM=60, energy=0.1, ZCR=0.05 → should be < 0.3.
	ext := &DefaultExtractor{}
	got := ext.ExtractDanceability(60.0, 0.1, 0.05)

	if got >= 0.3 {
		t.Errorf("ExtractDanceability(slow) = %.4f, want < 0.3", got)
	}
	if got > 1.0 || got < 0.0 {
		t.Errorf("ExtractDanceability(slow) = %.4f, out of [0,1] range", got)
	}
}

// === 3.2 RED: Acousticness ===

func TestExtractAcousticness_AcousticRecording(t *testing.T) {
	// Low spectral centroid (~800 Hz) + moderate energy (0.3) → acoustic > 0.6.
	ext := &DefaultExtractor{}
	got := ext.ExtractAcousticness(800.0, 0.3)

	if got <= 0.6 {
		t.Errorf("ExtractAcousticness(acoustic) = %.4f, want > 0.6", got)
	}
	if got > 1.0 || got < 0.0 {
		t.Errorf("ExtractAcousticness(acoustic) = %.4f, out of [0,1] range", got)
	}
}

func TestExtractAcousticness_Electronic(t *testing.T) {
	// High spectral centroid (~2000 Hz) + high energy (0.9) → acoustic < 0.3.
	ext := &DefaultExtractor{}
	got := ext.ExtractAcousticness(2000.0, 0.9)

	if got >= 0.3 {
		t.Errorf("ExtractAcousticness(electronic) = %.4f, want < 0.3", got)
	}
	if got > 1.0 || got < 0.0 {
		t.Errorf("ExtractAcousticness(electronic) = %.4f, out of [0,1] range", got)
	}
}
