# persistence Specification

## Purpose

Define the SQLite-based persistence layer for storing and retrieving `Track` entries with their associated `TrackFeatures`, including schema migration and error conditions.

## Requirements

### Requirement: Track Persistence

The repository MUST store a `Track` with its associated `TrackFeatures` in a single transactional operation.

#### Scenario: Save a track with features

- GIVEN a Track and TrackFeatures with valid values
- WHEN `Save` is called on the repository
- THEN the track and features are persisted to SQLite
- AND no error is returned

#### Scenario: Duplicate track ID returns conflict error

- GIVEN a Track with ID "abc-123" is already persisted
- WHEN `Save` is called again with the same ID
- THEN a conflict error is returned
- AND the error message includes the duplicate ID

### Requirement: Track Retrieval

The repository MUST retrieve a `Track` by unique ID, including its full `TrackFeatures`.

#### Scenario: Retrieve existing track by ID

- GIVEN a Track with ID "abc-123" was previously saved
- WHEN `GetByID` is called with "abc-123"
- THEN the returned Track matches the original
- AND its TrackFeatures match the originally stored values

#### Scenario: Non-existent ID returns not-found error

- GIVEN no track with ID "nonexistent" exists
- WHEN `GetByID` is called with "nonexistent"
- THEN a "not found" error is returned
- AND the error can be distinguished from other errors via `errors.Is`

### Requirement: Track Listing

The repository MUST list all stored tracks, ordered by insertion time.

#### Scenario: List returns all tracks

- GIVEN three tracks have been saved
- WHEN `List` is called
- THEN all three are returned
- AND they are ordered by insertion time ascending

#### Scenario: Empty database returns empty slice

- GIVEN no tracks have been saved
- WHEN `List` is called
- THEN an empty slice is returned (not nil)
- AND no error is returned

### Requirement: Schema Auto-Migration

The repository SHALL auto-create the SQLite schema on first use, creating `tracks` and `track_features` tables with appropriate columns, types, and a foreign key from features to tracks.

#### Scenario: First save creates schema

- GIVEN an empty SQLite database file
- WHEN the first `Save` is called
- THEN the `tracks` and `track_features` tables exist
- AND the save succeeds

#### Scenario: Re-opening database preserves data

- GIVEN a database file with persisted tracks
- WHEN a new repository instance opens the same file
- THEN all previously saved tracks are retrievable
- AND no migration runs (schema already matches)

### Requirement: Error Resilience

The repository MUST return structured errors for all failure modes and SHALL NOT panic on any storage operation.

#### Scenario: Corrupt database file

- GIVEN the SQLite file is corrupted (random bytes)
- WHEN the repository opens it
- THEN an error is returned during open or first operation
- AND no panic occurs

### Requirement: Fingerprint Storage

The repository MUST store a unique content fingerprint per track.

#### Scenario: Save stores fingerprint

- GIVEN a track with a fingerprint
- WHEN saved
- THEN the fingerprint persists

#### Scenario: Duplicate fingerprint

- GIVEN a track whose fingerprint exists
- WHEN saved
- THEN it is skipped as a duplicate

### Requirement: Fingerprint Dedupe

The repository MUST skip re-saving a track when its fingerprint already exists.

#### Scenario: Already-analyzed skipped

- GIVEN a present fingerprint
- WHEN the batch reaches it
- THEN it is skipped, no new row written

### Requirement: Concurrent Save Safety

Concurrent saves MUST NOT fail with `SQLITE_BUSY`.

#### Scenario: Concurrent saves succeed

- GIVEN multiple tracks saved concurrently
- WHEN all complete
- THEN no save fails busy

### Requirement: Idempotent Re-run

Re-running analysis SHALL leave analyzed tracks unchanged.

#### Scenario: Second run changes nothing

- GIVEN an already-analyzed library
- WHEN the batch runs again
- THEN no duplicate rows, tracks unchanged
