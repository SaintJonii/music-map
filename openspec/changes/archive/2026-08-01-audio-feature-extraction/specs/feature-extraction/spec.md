# Delta for feature-extraction

## ADDED Requirements

### Requirement: RMS Energy Extractor

The system MUST compute RMS energy from PCM samples, returning a float64 in [0, 1].

#### Scenario: 1 kHz sine at half amplitude

- GIVEN a 44100 Hz PCM buffer of a 1 kHz sine at amplitude 0.5
- WHEN RMS energy is computed
- THEN the result is approximately 0.5 (within ±0.01)

#### Scenario: Silence returns zero energy

- GIVEN a buffer of all-zero samples
- WHEN RMS energy is computed
- THEN the result is exactly 0.0

### Requirement: Zero-Crossing Rate

The system MUST count zero crossings per second, handling DC-offset signals via threshold crossings.

#### Scenario: 440 Hz sine wave

- GIVEN a 1-second 440 Hz sine at 44100 Hz
- WHEN ZCR is computed
- THEN the result is approximately 880 crossings/second (within ±5%)

#### Scenario: DC offset signal

- GIVEN a signal shifted by +0.5 DC
- WHEN ZCR uses threshold-based detection
- THEN the result is 0 crossings per second

### Requirement: Spectral Centroid

The system MUST compute the spectral centroid as the weighted mean frequency of the magnitude spectrum.

#### Scenario: Low-frequency sine

- GIVEN a 100 Hz sine wave
- WHEN spectral centroid is computed via FFT
- THEN the centroid is approximately 100 Hz (within ±10 Hz)

### Requirement: BPM Detection

The system MUST estimate tempo using onset-detection and autocorrelation, achieving ±5% accuracy on synthetic click tracks.

#### Scenario: 120 BPM click track

- GIVEN a 10-second synthetic click track at 120 BPM
- WHEN BPM detection runs
- THEN the result is between 114 and 126 BPM

#### Scenario: Very short input

- GIVEN a PCM buffer shorter than 1 second
- WHEN BPM detection runs
- THEN either a best-effort estimate with warning is returned OR an error indicates insufficient duration

### Requirement: Chroma Vector

The system MUST compute a 12-bin chroma vector normalized to sum to 1.0.

#### Scenario: C-major chord

- GIVEN PCM of a C-major triad (C, E, G)
- WHEN chroma is computed
- THEN bins for C (0), E (4), G (7) have the highest values

### Requirement: MFCC Coefficients

The system MUST compute 13 Mel-frequency cepstral coefficients per analysis frame.

#### Scenario: Valid PCM input

- GIVEN a 1024-sample PCM frame
- WHEN MFCCs are computed
- THEN 13 coefficients are returned; MFCC-0 represents overall energy

#### Scenario: Empty input

- GIVEN an empty PCM buffer
- WHEN MFCCs are computed
- THEN a zero-valued [13]float64 vector is returned

### Requirement: Key Detection

The system MUST classify the detected key as major or minor with a tonic note label.

#### Scenario: A-minor chord progression

- GIVEN PCM of an A-minor triad
- WHEN key detection runs
- THEN the result is "A minor"

### Requirement: Danceability Estimation

The system SHALL compute danceability as a weighted combination of BPM, energy, and ZCR, returning a float64 in [0, 1].

#### Scenario: High-energy 120 BPM track

- GIVEN BPM=120, energy=0.9, and ZCR=0.3
- WHEN danceability is computed
- THEN the result is above 0.7

#### Scenario: Slow ambient track

- GIVEN BPM=60, energy=0.1, and ZCR=0.05
- WHEN danceability is computed
- THEN the result is below 0.3

### Requirement: Acousticness Estimation

The system SHALL estimate acousticness from spectral centroid and energy, returning float64 in [0, 1]. Lower centroid + lower energy → higher acousticness.

#### Scenario: Acoustic guitar recording

- GIVEN a low spectral centroid (~800 Hz) and moderate energy (0.3)
- WHEN acousticness is computed
- THEN the result is above 0.6

#### Scenario: Electronic track

- GIVEN a high spectral centroid (~2000 Hz) and high energy (0.9)
- WHEN acousticness is computed
- THEN the result is below 0.3
