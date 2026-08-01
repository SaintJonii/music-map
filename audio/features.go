package audio

import (
	"errors"
	"math"
)

// ErrInsufficientDuration is returned when audio is too short for feature extraction.
var ErrInsufficientDuration = errors.New("audio: insufficient duration for BPM detection")

// BPMSensor estimates tempo from PCM samples.
type BPMSensor interface {
	ExtractBPM(samples []float64, sampleRate int) (float64, error)
}

// KeyDetector classifies musical key from a chroma vector.
type KeyDetector interface {
	ExtractKey(chroma [12]float64) string
}

// EnergyAnalyzer computes RMS energy and zero-crossing rate.
type EnergyAnalyzer interface {
	ExtractRMS(samples []float64) float64
	ExtractZCR(samples []float64, sampleRate int) float64
}

// SpectralAnalyzer computes frequency-domain features.
type SpectralAnalyzer interface {
	ExtractSpectralCentroid(samples []float64, sampleRate int) (float64, error)
	ExtractChroma(samples []float64, sampleRate int) ([12]float64, error)
	ExtractMFCCs(frame []float64) ([13]float64, error)
}

// VibeEstimator estimates perceptual features from derived values.
type VibeEstimator interface {
	ExtractDanceability(bpm, energy float64, zcr float64) float64
	ExtractAcousticness(spectralCentroid float64, energy float64) float64
}

// FeatureExtractor composes all feature extraction capabilities.
type FeatureExtractor interface {
	BPMSensor
	KeyDetector
	EnergyAnalyzer
	SpectralAnalyzer
	VibeEstimator
}

// DefaultExtractor is the default implementation of FeatureExtractor
// using go-dsp FFT and standard DSP formulas.
type DefaultExtractor struct{}

// ExtractRMS computes the root mean square energy of the samples.
// Returns a value in [0, 1] representing average signal power.
func (e *DefaultExtractor) ExtractRMS(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sumSq float64
	for _, s := range samples {
		sumSq += s * s
	}
	return math.Sqrt(sumSq / float64(len(samples)))
}

// ExtractZCR computes the zero-crossing rate using a mean-centered threshold.
// Returns crossings per second. Handles DC-offset signals by centering.
func (e *DefaultExtractor) ExtractZCR(samples []float64, sampleRate int) float64 {
	if len(samples) < 2 {
		return 0
	}

	// Compute mean to handle DC offset.
	var sum float64
	for _, s := range samples {
		sum += s
	}
	mean := sum / float64(len(samples))

	// If all samples are equal (pure DC), there are no zero crossings.
	// Check by verifying no variation around the mean.
	var hasVariation bool
	for _, s := range samples {
		if math.Abs(s-mean) > 1e-15 {
			hasVariation = true
			break
		}
	}
	if !hasVariation {
		return 0
	}

	// Count sign changes in mean-centered samples.
	var crossings int
	prev := samples[0] - mean
	for i := 1; i < len(samples); i++ {
		curr := samples[i] - mean
		if (prev < 0 && curr >= 0) || (prev >= 0 && curr < 0) {
			crossings++
		}
		prev = curr
	}

	durationSec := float64(len(samples)) / float64(sampleRate)
	return float64(crossings) / durationSec
}

// ExtractSpectralCentroid computes the weighted mean frequency of the magnitude spectrum.
// Returns frequency in Hz.
func (e *DefaultExtractor) ExtractSpectralCentroid(samples []float64, sampleRate int) (float64, error) {
	if len(samples) == 0 {
		return 0, nil
	}

	mags := magnitudeSpectrum(samples)
	N := len(mags)*2 - 2 // approximate original FFT length for frequency mapping
	if N <= 0 {
		N = 2
	}

	var weightedSum, magSum float64
	for i, mag := range mags {
		freq := float64(i) * float64(sampleRate) / float64(N)
		weightedSum += freq * mag
		magSum += mag
	}

	if magSum == 0 {
		return 0, nil
	}

	return weightedSum / magSum, nil
}

// ——— 3.2 Feature Extractors ———

// ExtractBPM estimates tempo in beats per minute using onset detection and autocorrelation.
// Requires at least 1 second of audio.
func (e *DefaultExtractor) ExtractBPM(samples []float64, sampleRate int) (float64, error) {
	minSamples := sampleRate // 1 second
	if len(samples) < minSamples {
		return 0, ErrInsufficientDuration
	}

	// Frame the signal into ~256-sample windows with 50% overlap.
	frameSize := 256
	hopSize := frameSize / 2
	var energies []float64
	for start := 0; start+frameSize <= len(samples); start += hopSize {
		frame := samples[start : start+frameSize]
		var sumSq float64
		for _, s := range frame {
			sumSq += s * s
		}
		energies = append(energies, math.Sqrt(sumSq/float64(frameSize)))
	}

	if len(energies) < 4 {
		return 0, ErrInsufficientDuration
	}

	// Onset detection: positive first difference of energy.
	var onsets []float64
	for i := 1; i < len(energies); i++ {
		diff := energies[i] - energies[i-1]
		if diff > 0 {
			onsets = append(onsets, diff)
		} else {
			onsets = append(onsets, 0)
		}
	}

	// Autocorrelation of onset signal.
	corr := autocorrelate(onsets)

	// Search for peaks in the BPM range (60–200 BPM).
	// The frame rate for the onset signal: sampleRate / hopSize.
	onsetRate := float64(sampleRate) / float64(hopSize)

	// Convert BPM range to lag indices.
	// BPM = 60 * onsetRate / lag
	// lag = 60 * onsetRate / BPM
	minLag := int(60.0 * onsetRate / 200.0)
	maxLag := int(60.0 * onsetRate / 60.0)
	if minLag < 1 {
		minLag = 1
	}
	if maxLag >= len(corr) {
		maxLag = len(corr) - 1
	}

	// Find the lag with maximum correlation in the BPM range (skip lag=0).
	bestLag := 0
	bestVal := 0.0
	for lag := minLag; lag <= maxLag; lag++ {
		if corr[lag] > bestVal {
			bestVal = corr[lag]
			bestLag = lag
		}
	}

	if bestLag == 0 {
		return 0, errors.New("audio: no dominant BPM detected")
	}

	bpm := 60.0 * onsetRate / float64(bestLag)
	return bpm, nil
}

// ExtractChroma computes a 12-bin chroma vector normalized to sum to 1.0.
// Maps FFT frequency bins to pitch classes.
func (e *DefaultExtractor) ExtractChroma(samples []float64, sampleRate int) ([12]float64, error) {
	var chroma [12]float64
	if len(samples) == 0 {
		return chroma, nil
	}

	mags := magnitudeSpectrum(samples)
	N := len(mags)*2 - 2
	if N <= 0 {
		N = 2
	}

	// Map each positive frequency bin to a chroma bin.
	// chromaBin = round(12 * log2(f / C0)) % 12
	for i := 1; i < len(mags); i++ {
		freq := float64(i) * float64(sampleRate) / float64(N)
		bin := int(math.Round(12.0*math.Log2(freq/C0RefFreq))) % 12
		if bin < 0 {
			bin += 12
		}
		chroma[bin] += mags[i]
	}

	// Normalize to sum to 1.0.
	var total float64
	for _, v := range chroma {
		total += v
	}
	if total > 0 {
		for i := range chroma {
			chroma[i] /= total
		}
	}

	return chroma, nil
}

// ExtractMFCCs computes 13 Mel-frequency cepstral coefficients from a PCM frame.
func (e *DefaultExtractor) ExtractMFCCs(frame []float64) ([13]float64, error) {
	var zero [13]float64
	if len(frame) == 0 {
		return zero, nil
	}

	// Compute power spectrum via shared helper.
	powerSpectrum := powerSpectrumFunc(frame)
	halfN := len(powerSpectrum)
	if halfN <= 1 {
		return zero, nil
	}
	// Use frame length as effective sample rate for frequency mapping.
	nyquist := float64(len(frame)) / 2.0

	// Create mel filterbank.
	melMax := hzToMel(nyquist)
	melMin := hzToMel(melMinFreq)
	melPoints := make([]float64, melNumFilters+2)
	for i := 0; i < melNumFilters+2; i++ {
		melPoints[i] = melMin + float64(i)*(melMax-melMin)/float64(melNumFilters+1)
	}

	// Apply filterbank.
	melEnergies := make([]float64, melNumFilters)
	for m := 0; m < melNumFilters; m++ {
		var energy float64
		fStart := melToHz(melPoints[m])
		fCenter := melToHz(melPoints[m+1])
		fEnd := melToHz(melPoints[m+2])

		for i := 0; i < halfN; i++ {
			freq := float64(i) * nyquist / float64(halfN-1)
			if halfN <= 1 {
				freq = 0
			}
			weight := triangularFilter(freq, fStart, fCenter, fEnd)
			energy += weight * powerSpectrum[i]
		}
		// Add small epsilon to avoid log(0).
		melEnergies[m] = math.Log(math.Max(energy, 1e-10))
	}

	// DCT-II on mel energies, keep first 13 coefficients.
	for k := 0; k < 13; k++ {
		var sum float64
		for n := 0; n < melNumFilters; n++ {
			sum += melEnergies[n] * math.Cos(math.Pi*float64(k)*(float64(n)+0.5)/float64(melNumFilters))
		}
		zero[k] = sum * math.Sqrt(2.0/float64(melNumFilters))
	}

	return zero, nil
}

// Krumhansl-Schmuckler key templates for major and minor keys.
// Each template is a 12-element vector of weights for each pitch class.
var majorTemplate = [12]float64{6.35, 2.23, 3.48, 2.33, 4.38, 4.09, 2.52, 5.19, 2.39, 3.66, 2.29, 2.88}
var minorTemplate = [12]float64{6.33, 2.68, 3.52, 5.38, 2.60, 3.53, 2.54, 4.75, 3.98, 2.69, 3.34, 3.17}

var keyNames = [12]string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

// ExtractKey classifies the musical key from a chroma vector using
// Krumhansl-Schmuckler key-finding.
func (e *DefaultExtractor) ExtractKey(chroma [12]float64) string {
	bestKey := ""
	bestCorr := -1.0

	// Test all 12 major keys.
	for tonic := 0; tonic < 12; tonic++ {
		corr := correlateTemplates(chroma, majorTemplate, tonic)
		if corr > bestCorr {
			bestCorr = corr
			bestKey = keyNames[tonic] + " major"
		}
	}

	// Test all 12 minor keys.
	for tonic := 0; tonic < 12; tonic++ {
		corr := correlateTemplates(chroma, minorTemplate, tonic)
		if corr > bestCorr {
			bestCorr = corr
			bestKey = keyNames[tonic] + " minor"
		}
	}

	return bestKey
}

// correlateTemplates computes the Pearson correlation between a chroma vector
// and a key template shifted to a given tonic.
func correlateTemplates(chroma [12]float64, template [12]float64, tonic int) float64 {
	var sumChrom, sumTemp, sumC2, sumT2 float64
	for i := 0; i < 12; i++ {
		c := chroma[i]
		t := template[(i-tonic+12)%12]
		sumChrom += c
		sumTemp += t
		sumC2 += c * c
		sumT2 += t * t
	}

	n := 12.0
	num := 0.0
	for i := 0; i < 12; i++ {
		c := chroma[i]
		t := template[(i-tonic+12)%12]
		num += (c - sumChrom/n) * (t - sumTemp/n)
	}

	denChrom := math.Sqrt(sumC2 - sumChrom*sumChrom/n)
	denTemp := math.Sqrt(sumT2 - sumTemp*sumTemp/n)

	if denChrom == 0 || denTemp == 0 {
		return 0
	}
	return num / (denChrom * denTemp)
}

// ExtractDanceability estimates danceability as a weighted combination of BPM, energy, and ZCR.
// Returns a value in [0, 1].
func (e *DefaultExtractor) ExtractDanceability(bpm, energy float64, zcr float64) float64 {
	// BPM component: sigmoid centered at 120 BPM.
	bpmScore := sigmoid((bpm-120.0)/30.0)

	// Energy component: linear, higher is more danceable.
	energyScore := clamp01(energy)

	// ZCR component: moderate ZCR (0.1-0.4) is ideal for dance music.
	// Use a bell curve centered at 0.25.
	zcrScore := math.Exp(-math.Pow((zcr-0.25)/0.15, 2))

	// Weighted combination: BPM 40%, Energy 35%, ZCR 25%.
	score := 0.4*bpmScore + 0.35*energyScore + 0.25*zcrScore
	return clamp01(score)
}

// ExtractAcousticness estimates acousticness from spectral centroid and energy.
// Higher centroid and energy → less acoustic.
// Returns a value in [0, 1].
func (e *DefaultExtractor) ExtractAcousticness(spectralCentroid float64, energy float64) float64 {
	// Map spectral centroid: typical acoustic < 1500 Hz, electronic > 3000 Hz.
	centroidScore := sigmoid((1500.0-spectralCentroid)/500.0)

	// Map energy: lower energy suggests acoustic.
	energyScore := sigmoid((0.4-energy)/0.2)

	// Combined: centroid 60%, energy 40%.
	score := 0.6*centroidScore + 0.4*energyScore
	return clamp01(score)
}
