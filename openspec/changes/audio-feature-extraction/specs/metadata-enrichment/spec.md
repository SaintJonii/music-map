# Delta for metadata-enrichment

## ADDED Requirements

### Requirement: Local Tag Reading

The system MUST extract Title, Artist, Album, AlbumArtist, Genre, Year, ISRC, and MusicBrainz ID from ID3v2 tags (MP3) and Vorbis comments (FLAC) using `dhowden/tag`. Tags-first strategy: local tags are the PRIMARY data source — acoustic analysis only runs when features are missing.

#### Scenario: MP3 with full ID3v2 tags

- GIVEN an MP3 file with ID3v2 tags containing title, artist, album, and MBID
- WHEN tags are read
- THEN all fields are populated correctly with no error

#### Scenario: File without tags returns empty metadata

- GIVEN a WAV file with no metadata chunks
- WHEN tags are read
- THEN all string fields are empty and no error is returned

#### Scenario: Unreadable file returns error

- GIVEN a path to a non-existent file
- WHEN tags are read
- THEN a "file not found" error is returned

### Requirement: Vorbis Comment Extraction

The FLAC tag reader MUST extract metadata fields from Vorbis comments using standard mapping keys.

#### Scenario: FLAC with Vorbis comments

- GIVEN a FLAC file with TITLE, ARTIST, and MUSICBRAINZ_TRACKID comments
- WHEN tags are read
- THEN Title, Artist, and MBID fields are populated

#### Scenario: FLAC with no comments

- GIVEN a FLAC file with no Vorbis comment block
- WHEN tags are read
- THEN all fields are empty and no error is returned

### Requirement: MusicBrainz Enrichment

The system MUST query the MusicBrainz API via `go-musicbrainzws2` when a valid ISRC or MBID is present in local tags. ISRC is the preferred lookup key; MBID is the fallback.

#### Scenario: Valid MBID resolves successfully

- GIVEN a Track with a known-valid MusicBrainz recording MBID
- WHEN enrichment is triggered
- THEN the response includes title, artist, and release metadata

#### Scenario: Unknown MBID returns not-found

- GIVEN a Track with a non-existent MBID
- WHEN enrichment is queried
- THEN a "not found" error is returned; local metadata is preserved

### Requirement: HTTP Error Handling

The MusicBrainz client MUST handle transport errors gracefully without panicking.

#### Scenario: HTTP timeout during enrichment

- GIVEN the MusicBrainz API is unreachable (simulated timeout)
- WHEN enrichment is queried
- THEN a retryable error is returned within the configured deadline

#### Scenario: HTTP 503 Service Unavailable

- GIVEN the MusicBrainz API returns HTTP 503
- WHEN enrichment is queried
- THEN a descriptive error containing the status code is returned
