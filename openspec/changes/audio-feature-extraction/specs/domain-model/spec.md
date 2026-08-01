# Delta for domain-model

## ADDED Requirements

### Requirement: Track Struct Definition

The system MUST provide a `Track` struct with fields: `ID`, `Title`, `Artist`, `Album`, `AlbumArtist`, `Genre`, `Year`, `TrackNumber`, `Duration` (float64), `Format`, `BitRate` (int), `ISRC`, and `FilePath`.

#### Scenario: Track holds all fields

- GIVEN a Track is constructed with all field values
- WHEN fields are read back
- THEN every field matches the input value

#### Scenario: Empty Title defaults to "Unknown"

- GIVEN a Track is created with an empty Title string
- WHEN Title is read
- THEN it SHALL default to `"Unknown"`

### Requirement: TrackFeatures Struct Definition

The system MUST provide a `TrackFeatures` struct with fields: `BPM`, `Key`, `Energy` ([0,1]), `Danceability` ([0,1]), `Acousticness` ([0,1]), `SpectralCentroid`, `Chroma` ([12]float64), `MFCCs` ([13]float64), and `ZCR`.

#### Scenario: Full features round-trip through JSON

- GIVEN a TrackFeatures with valid values
- WHEN it is marshaled to JSON and back
- THEN all fields are preserved within float precision

#### Scenario: Missing optional features marshal as zero/empty

- GIVEN TrackFeatures where Key is empty and BPM is 0
- WHEN marshaled to JSON
- THEN the JSON contains `"key": ""` and `"bpm": 0`

### Requirement: Collection Struct

The system MUST provide a `Collection` struct aggregating `[]Track` with `Name` and `CreatedAt`.

#### Scenario: Collection holds multiple tracks

- GIVEN a Collection with three Track entries
- WHEN tracks are enumerated
- THEN all three are present in order

#### Scenario: Empty collection serializes to empty array

- GIVEN a Collection with zero tracks
- WHEN serialized to JSON
- THEN the tracks field is `[]`, not `null`

### Requirement: JSON Serialization Contract

All domain structs MUST survive a full marshal-unmarshal cycle without data loss.

#### Scenario: Track round-trips through JSON

- GIVEN a fully populated Track
- WHEN `json.Marshal` then `json.Unmarshal` into a new Track
- THEN the original and deserialized structs are deeply equal

#### Scenario: Malformed JSON returns error

- GIVEN an invalid JSON payload with type mismatch
- WHEN `json.Unmarshal` is called
- THEN a descriptive error is returned
