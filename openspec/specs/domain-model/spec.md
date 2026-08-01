# domain-model Specification

## Purpose

Define the core domain types — `Track`, `TrackFeatures`, and `Collection` — their fields, constraints, and JSON serialization contract.

## Requirements

### Requirement: Track Struct Definition

The system MUST provide a `Track` struct with fields: `ID`, `Title`, `Artist`, `Album`, `AlbumArtist`, `Genre`, `Year`, `TrackNumber`, `Duration` (seconds, float64), `Format` (string), `BitRate` (int), `ISRC` (string), and `FilePath`.

#### Scenario: Track holds all fields

- GIVEN a Track is constructed with all field values
- WHEN fields are read back
- THEN every field matches the input value

#### Scenario: Empty Title defaults

- GIVEN a Track is created with an empty Title string
- WHEN Title is read
- THEN it SHALL default to `"Unknown"`

### Requirement: TrackFeatures Struct Definition

The system MUST provide a `TrackFeatures` struct with fields: `BPM` (float64), `Key` (string), `Energy` (float64 in [0,1]), `Danceability` (float64 in [0,1]), `Acousticness` (float64 in [0,1]), `SpectralCentroid` (float64), `Chroma` ([12]float64), `MFCCs` ([13]float64), and `ZCR` (float64).

#### Scenario: Full features round-trip through JSON

- GIVEN a TrackFeatures with valid values
- WHEN it is marshaled to JSON and back
- THEN all field values are preserved within float precision

#### Scenario: Missing optional features marshal as zero/empty

- GIVEN TrackFeatures where Key is empty and BPM is 0
- WHEN marshaled to JSON
- THEN the JSON contains `"key": ""` and `"bpm": 0`
- AND no field is omitted

### Requirement: Collection Struct

The system MUST provide a `Collection` struct that aggregates `[]Track` results with an optional `Name` and `CreatedAt` timestamp.

#### Scenario: Collection holds multiple tracks

- GIVEN a Collection with three Track entries
- WHEN tracks are enumerated
- THEN all three are present in order

#### Scenario: Empty collection serializes to empty array

- GIVEN a Collection with zero tracks
- WHEN serialized to JSON
- THEN the tracks field is `[]`, not `null`

### Requirement: JSON Serialization Contract

All domain structs MUST implement `json.Marshaler` and `json.Unmarshaler` (via struct tags) and SHALL survive a full marshal-unmarshal cycle without data loss.

#### Scenario: Track round-trips through JSON

- GIVEN a fully populated Track
- WHEN `json.Marshal` then `json.Unmarshal` into a new Track
- THEN the original and deserialized structs are deeply equal

#### Scenario: Malformed JSON returns error

- GIVEN an invalid JSON payload (missing required field type mismatch)
- WHEN `json.Unmarshal` is called
- THEN an error is returned with a descriptive message
