# dev-workflow Specification

## Purpose
TBD - created by archiving change dev-workflow-tooling. Update Purpose after archive.
## Requirements
### Requirement: CHANGELOG with an Unreleased section

The repository SHALL maintain a `CHANGELOG.md` in Keep a Changelog format with a
`## [Unreleased]` section that shipping a change appends to.

#### Scenario: Unreleased section present

- **WHEN** `CHANGELOG.md` is read
- **THEN** it contains an `## [Unreleased]` heading with `### Added`/`### Changed`/
  `### Fixed` subsections as needed

### Requirement: One gated ship script

`scripts/ship-change.sh <change> "<subject>"` SHALL stage the tree, run the gate
(`nix flake check`), archive the named OpenSpec change, commit as Pim Snel with no
self-promoting trailers, and push `main`. If the gate fails it SHALL exit non-zero
without archiving, committing, or pushing.

#### Scenario: Gate failure aborts the ship

- **WHEN** `nix flake check` fails during a ship
- **THEN** the script exits non-zero and does not archive/commit/push

#### Scenario: Successful ship

- **WHEN** the gate passes
- **THEN** the change is archived, committed as Pim Snel, and `main` is pushed

### Requirement: Release script promotes Unreleased

`scripts/release.sh <version>` SHALL rewrite the `## [Unreleased]` section to a dated
`## [<version>]` section and leave a fresh empty `## [Unreleased]` above it.

#### Scenario: Promote unreleased

- **WHEN** `release.sh 0.2.0` is run with content under `[Unreleased]`
- **THEN** that content moves under `## [0.2.0] - <date>` and a new empty
  `[Unreleased]` remains

### Requirement: Coverage gate enforces a core floor

The `coverage` flake check SHALL fail unless total statement coverage is ≥70% AND each
core package (`internal/index`, `internal/authz`, `internal/auth`, `internal/redact`,
`internal/defense`, `internal/gitsafe`, `internal/config`, `internal/corpus`) has ≥80%
own coverage.

#### Scenario: A core package below 80% fails the gate

- **WHEN** a core package's own coverage drops below 80%
- **THEN** the coverage check fails

#### Scenario: Overall and core both satisfied passes

- **WHEN** total coverage is ≥70% and every core package is ≥80%
- **THEN** the coverage check passes

