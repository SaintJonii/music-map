# library-scanning Specification

## Requirements

| Requirement | Behavior |
|-------------|----------|
| LibrarySource Interface (MUST) | `List` → `TrackRef` entries; `Open` → readable stream. |
| Local Folder Source (MUST) | List tracks under a folder recursively. |
| Audio-Only Filtering (MUST NOT) | Ignore non-audio. |
| Track Opening (MUST) | Stream for a valid `TrackRef`; error on missing. |
| Empty Folder (SHALL) | Empty list, no error. |
| Unreadable Folder (MUST) | Descriptive error, no panic. |

### Requirement: LibrarySource Interface

#### Scenario: List and open

- GIVEN any `LibrarySource`
- WHEN `List` then `Open` are called
- THEN entries and streams return via the interface

### Requirement: Local Folder Source

#### Scenario: Nested scan

- GIVEN tracks at root and subfolders
- WHEN scanned
- THEN all audio tracks return

#### Scenario: Non-audio ignored

- GIVEN audio mixed with non-audio
- WHEN scanned
- THEN only audio appears

#### Scenario: Empty folder

- GIVEN an empty folder
- WHEN scanned
- THEN an empty list, no error

#### Scenario: Unreadable folder

- GIVEN an unreadable folder
- WHEN scanned
- THEN a descriptive error, no panic

### Requirement: Track Opening

#### Scenario: Open listed track

- GIVEN a `TrackRef` from `List`
- WHEN `Open` is called
- THEN a readable stream returns

#### Scenario: Missing file

- GIVEN a `TrackRef` whose file is gone
- WHEN `Open` is called
- THEN an error returns
