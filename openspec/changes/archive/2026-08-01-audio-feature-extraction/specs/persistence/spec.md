# Delta for persistence

## ADDED Requirements

### Requirement: Track Persistence

The repository MUST store a `Track` with its `TrackFeatures` in a single transactional operation.

#### Scenario: Save a track with features

- GIVEN a Track and TrackFeatures with valid values
- WHEN `Save` is called
- THEN the track and features are persisted to SQLite with no error

#### Scenario: Duplicate track ID returns conflict error

- GIVEN a Track with ID "abc-123" is already persisted
- WHEN `Save` is called again with the same ID
- THEN a conflict error including the duplicate ID is returned

### Requirement: Track Retrieval

The repository MUST retrieve a `Track` by unique ID, including its `TrackFeatures`.

#### Scenario: Retrieve existing track by ID

- GIVEN a Track with ID "abc-123" was previously saved
- WHEN `GetByID` is called with "abc-123"
- THEN the returned Track and its TrackFeatures match the original

#### Scenario: Non-existent ID returns not-found error

- GIVEN no track with ID "nonexistent" exists
- WHEN `GetByID` is called with "nonexistent"
- THEN a "not found" error distinguishable via `errors.Is` is returned

### Requirement: Track Listing

The repository MUST list all stored tracks ordered by insertion time.

#### Scenario: List returns all tracks

- GIVEN three tracks have been saved
- WHEN `List` is called
- THEN all three are returned in insertion order

#### Scenario: Empty database returns empty slice

- GIVEN no tracks have been saved
- WHEN `List` is called
- THEN an empty slice (not nil) is returned with no error

### Requirement: Schema Auto-Migration

The repository SHALL auto-create the SQLite schema on first use with `tracks` and `track_features` tables and a foreign key.

#### Scenario: First save creates schema

- GIVEN an empty SQLite database file
- WHEN the first `Save` is called
- THEN tables exist and the save succeeds

#### Scenario: Re-opening database preserves data

- GIVEN a database file with persisted tracks
- WHEN a new repository instance opens the same file
- THEN all previously saved tracks are retrievable

### Requirement: Error Resilience

The repository MUST return structured errors for all failure modes and SHALL NOT panic on storage operations.

#### Scenario: Corrupt database file

- GIVEN the SQLite file is corrupted
- WHEN the repository opens it
- THEN an error is returned and no panic occurs
