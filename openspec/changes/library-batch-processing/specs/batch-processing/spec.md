# batch-processing Specification

## Requirements

| Requirement | Behavior |
|-------------|----------|
| Concurrent Analysis (MUST) | Analyze valid files concurrently, per-file results. |
| Failure Isolation (MUST NOT) | A corrupt file MUST NOT stop the run. |
| Run Summary (MUST) | Counts and per-file failure details. |
| Cancellation (MUST) | Honor cancellation, return fast. |
| Deterministic Results (MUST) | Same result set at any worker count. |

### Requirement: Concurrent Analysis

#### Scenario: Batch of valid files

- GIVEN a batch of valid files
- WHEN processed
- THEN every file succeeds

### Requirement: Failure Isolation

#### Scenario: One corrupt file among valid

- GIVEN a corrupt file among valid files
- WHEN processed
- THEN valid files succeed and the corrupt one is reported as a failure

### Requirement: Run Summary

#### Scenario: Mixed run summary

- GIVEN a run with successes and failures
- WHEN complete
- THEN the summary lists each failure, file, and error

### Requirement: Cancellation

#### Scenario: Cancel mid-run

- GIVEN a run in progress
- WHEN cancelled
- THEN the runner stops and returns promptly

### Requirement: Deterministic Results

#### Scenario: Different worker counts

- GIVEN one batch at varied worker counts
- WHEN both runs complete
- THEN success and failure sets are identical
