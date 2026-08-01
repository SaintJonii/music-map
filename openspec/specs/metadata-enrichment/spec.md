# metadata-enrichment Specification

## Purpose

Define the metadata pipeline with tags-first data source priority: (1) local ID3v2/Vorbis tags → (2) Navidrome/Subsonic API (post-MVP) → (3) acoustic analysis fallback → (4) MusicBrainz enrichment via ISRC or MBID. Also define the LibrarySource interface for swappable backends.

## Requirements

### Requirement: Local Tag Reading

The system MUST extract Title, Artist, Album, AlbumArtist, Genre, Year, TrackNumber, ISRC, and MusicBrainz ID fields from ID3v2 tags (MP3) and Vorbis comments (FLAC) using the `dhowden/tag` library. Local tags are the PRIMARY data source — acoustic analysis SHALL only run when tags are absent or incomplete.

#### Scenario: MP3 with full ID3v2 tags

- GIVEN an MP3 file with ID3v2 tags containing title, artist, album, genre, year, ISRC, and a MusicBrainz recording ID
- WHEN tags are read
- THEN all fields are populated correctly
- AND no error is returned

#### Scenario: File without tags returns empty metadata

- GIVEN a WAV file with no ID3 or RIFF metadata chunks
- WHEN tags are read
- THEN all string fields are empty
- AND no error is returned (absence of tags is not a failure)

#### Scenario: Unreadable file returns error

- GIVEN a path to a non-existent file
- WHEN tags are read
- THEN a descriptive "file not found" error is returned

### Requirement: Vorbis Comment Extraction

The FLAC tag reader MUST extract the same metadata fields from Vorbis comments using standard mapping keys.

#### Scenario: FLAC with Vorbis comments

- GIVEN a FLAC file with Vorbis comment fields for TITLE, ARTIST, and MUSICBRAINZ_TRACKID
- WHEN tags are read
- THEN the Title, Artist, and MBID fields are populated

#### Scenario: FLAC with no comments

- GIVEN a FLAC file with no Vorbis comment block
- WHEN tags are read
- THEN all fields are empty
- AND no error is returned

### Requirement: MusicBrainz Enrichment

The system MUST query the MusicBrainz API via `go-musicbrainzws2` when a valid MBID or ISRC is present in local tags. Enrichment SHALL use ISRC as the preferred lookup key when available, falling back to MBID, and SHALL populate additional metadata fields from the response.

#### Scenario: Valid MBID resolves successfully

- GIVEN a Track with a known-valid MusicBrainz recording MBID
- WHEN enrichment is triggered
- THEN the response includes title, artist, and release metadata

#### Scenario: Unknown MBID returns not-found

- GIVEN a Track with a syntactically valid but non-existent MBID
- WHEN enrichment is queried
- THEN a "not found" error is returned
- AND the Track's local metadata is preserved unchanged

### Requirement: HTTP Error Handling

The MusicBrainz client MUST handle transport-level errors gracefully and SHALL NOT panic on timeout, DNS failure, or non-200 status codes.

#### Scenario: HTTP timeout during enrichment

- GIVEN the MusicBrainz API is unreachable (simulated timeout)
- WHEN enrichment is queried
- THEN a retryable error is returned within a configurable deadline
- AND the calling code can distinguish timeout from not-found errors

#### Scenario: HTTP 503 Service Unavailable

- GIVEN the MusicBrainz API returns HTTP 503
- WHEN enrichment is queried
- THEN a descriptive error is returned containing the status code
