# Delta for project-scaffold

## ADDED Requirements

### Requirement: Go Module Initialization

The project MUST be initialized as a Go module targeting Go 1.26 with a valid `go.mod` and package declaration.

#### Scenario: Fresh initialization creates go.mod

- GIVEN an empty project root
- WHEN `go mod init` is executed
- THEN `go.mod` exists with `go 1.26` directive

#### Scenario: Re-initialization is idempotent

- GIVEN an existing `go.mod`
- WHEN scaffold runs again
- THEN the module file SHALL NOT be overwritten unless forced

### Requirement: Directory Layout

The project MUST create package directories: `model/`, `audio/`, `metadata/`, `storage/`, and `testdata/`.

#### Scenario: Directories created by scaffold

- GIVEN a fresh Go module
- WHEN the scaffold step executes
- THEN all five directories exist with at least a Go source file each

#### Scenario: Scaffold is idempotent with existing directories

- GIVEN directories already exist with source files
- WHEN scaffold runs again
- THEN existing files SHALL NOT be overwritten

### Requirement: Tooling Configuration

The project SHALL include `.golangci.yml` and a CI baseline where `go test ./... -cover` runs without failure.

#### Scenario: Linter config present and valid

- GIVEN the scaffold completed
- WHEN `golangci-lint run` is invoked
- THEN the configuration file exists and is parseable

#### Scenario: Test suite runs on empty packages

- GIVEN packages exist with only placeholder source
- WHEN `go test ./...` runs
- THEN the command exits with status 0

### Requirement: Python Artifact Cleanup

All Python artifacts SHALL be removed: `src/`, `tests/`, `requirements*.txt`, and `pytest.ini`.

#### Scenario: Python files removed

- GIVEN Python artifacts exist at standard paths
- WHEN the cleanup step executes
- THEN `src/`, `tests/`, and requirements files are deleted

#### Scenario: No Python artifacts to remove

- GIVEN the repository has no Python files
- WHEN the cleanup step executes
- THEN the step succeeds with no errors or side effects
