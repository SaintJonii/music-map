# Delta for persistence

## ADDED Requirements

| Requirement | Behavior |
|-------------|----------|
| Fingerprint Storage (MUST) | Store a unique content fingerprint per track. |
| Fingerprint Dedupe (MUST) | Skip re-save when fingerprint exists. |
| Concurrent Save Safety (MUST NOT) | Concurrent saves MUST NOT fail with `SQLITE_BUSY`. |
| Idempotent Re-run (SHALL) | Re-running leaves analyzed tracks unchanged. |

### Requirement: Fingerprint Storage

#### Scenario: Save stores fingerprint

- GIVEN a track with a fingerprint
- WHEN saved
- THEN the fingerprint persists

#### Scenario: Duplicate fingerprint

- GIVEN a track whose fingerprint exists
- WHEN saved
- THEN it is skipped as a duplicate

### Requirement: Fingerprint Dedupe

#### Scenario: Already-analyzed skipped

- GIVEN a present fingerprint
- WHEN the batch reaches it
- THEN it is skipped, no new row written

### Requirement: Concurrent Save Safety

#### Scenario: Concurrent saves succeed

- GIVEN multiple tracks saved concurrently
- WHEN all complete
- THEN no save fails busy

### Requirement: Idempotent Re-run

#### Scenario: Second run changes nothing

- GIVEN an already-analyzed library
- WHEN the batch runs again
- THEN no duplicate rows, tracks unchanged
