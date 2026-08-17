package audio

import (
	"math"
	"math/cmplx"

	"github.com/madelynnblue/go-dsp/fft"
)

// C0RefFreq is the reference frequency for MIDI note 0 (C0 ≈ 8.176 Hz).
// Used for mapping frequencies to chroma bins.
const C0RefFreq = 8.1757989156

// mel constants for MFCC computation.
const (
	melNumFilters = 24
	melMinFreq    = 20.0
)

// Spectrum holds the result of a single forward FFT on windowed real samples.
// Mags contains the magnitude of each positive-frequency bin (indices 0 to
// N/2 inclusive); Power contains the squared magnitude (power) of those same
// bins; N is the original FFT length (len(samples)), used to map bin indices
// back to frequencies with freq = i * sampleRate / N.
type Spectrum struct {
	Mags  []float64
	Power []float64
	N     int
}

// HannMagnitudeSpectrum computes a single Hann-windowed forward FFT of
// real-valued samples and returns the magnitude and power for each
// positive-frequency bin. Callers needing both the magnitude and its power
// share one FFT instead of running it twice.
func HannMagnitudeSpectrum(samples []float64) Spectrum {
	windowed := make([]float64, len(samples))
	copy(windowed, samples)
	applyWindow(windowed, hannWindow(len(samples)))
	spectrum := fft.FFTReal(windowed)
	N := len(spectrum)
	halfN := N/2 + 1
	s := Spectrum{
		Mags:  make([]float64, halfN),
		Power: make([]float64, halfN),
		N:     N,
	}
	for i := range halfN {
		mag := cmplx.Abs(spectrum[i])
		s.Mags[i] = mag
		s.Power[i] = mag * mag
	}
	return s
}

// powerSpectrum computes the forward FFT and returns the squared magnitude
// (power) for each positive-frequency bin.
func powerSpectrumFunc(samples []float64) []float64 {
	windowed := make([]float64, len(samples))
	copy(windowed, samples)
	applyWindow(windowed, hammingWindow(len(samples)))
	spectrum := fft.FFTReal(windowed)
	N := len(spectrum)
	halfN := N/2 + 1
	pows := make([]float64, halfN)
	for i := range halfN {
		mag := cmplx.Abs(spectrum[i])
		pows[i] = mag * mag
	}
	return pows
}

// hammingWindow generates a Hamming window of length n.
func hammingWindow(n int) []float64 {
	w := make([]float64, n)
	for i := range n {
		w[i] = 0.54 - 0.46*math.Cos(2.0*math.Pi*float64(i)/float64(n-1))
	}
	return w
}

// hannWindow generates a Hann (Hanning) window of length n.
func hannWindow(n int) []float64 {
	w := make([]float64, n)
	for i := range n {
		w[i] = 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(i)/float64(n-1)))
	}
	return w
}

// applyWindow applies a window function to the samples in-place.
func applyWindow(samples []float64, window []float64) {
	n := min(len(samples), len(window))
	for i := range n {
		samples[i] *= window[i]
	}
}

// hzToMel converts frequency in Hz to mel scale.
func hzToMel(hz float64) float64 {
	return 2595.0 * math.Log10(1.0+hz/700.0)
}

// melToHz converts mel scale to frequency in Hz.
func melToHz(mel float64) float64 {
	return 700.0 * (math.Pow(10.0, mel/2595.0) - 1.0)
}

// triangularFilter returns the weight of a frequency in a triangular filter.
func triangularFilter(freq, fStart, fCenter, fEnd float64) float64 {
	if freq <= fStart || freq >= fEnd {
		return 0
	}
	if freq <= fCenter {
		return (freq - fStart) / (fCenter - fStart)
	}
	return (fEnd - freq) / (fEnd - fCenter)
}

// autocorrelate returns the normalized autocorrelation of a signal.
func autocorrelate(x []float64) []float64 {
	n := len(x)
	r := make([]float64, n)
	for lag := 0; lag < n; lag++ {
		var sum float64
		for i := 0; i < n-lag; i++ {
			sum += x[i] * x[i+lag]
		}
		r[lag] = sum
	}
	// Normalize by r[0].
	if r[0] > 0 {
		for i := range r {
			r[i] /= r[0]
		}
	}
	return r
}

// sigmoid returns 1/(1 + exp(-x)).
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// clamp01 clamps value to [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
