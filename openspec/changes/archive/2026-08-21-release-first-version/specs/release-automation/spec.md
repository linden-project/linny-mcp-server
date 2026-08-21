## ADDED Requirements

### Requirement: First release cuts the current version without bumping

`scripts/release.sh` SHALL allow releasing the current `VERSION` unchanged when no tag
exists for it yet (the first release), offering a `current` option alongside
`major`/`minor`/`patch`. When a tag already exists for the current `VERSION`, the
`current` option SHALL be unavailable/rejected so a version is never re-released.

#### Scenario: First release keeps the current version

- **WHEN** `VERSION` is `0.1.0`, no `v0.1.0` tag exists, and the maintainer selects
  `current`
- **THEN** the release is cut as `v0.1.0` with no version bump

#### Scenario: current is rejected once released

- **WHEN** a `vX.Y.Z` tag already exists for the current `VERSION` and `current` is
  requested
- **THEN** the script refuses (that version is already released) and no tag is created

#### Scenario: Subsequent releases still bump

- **WHEN** the current `VERSION` is already tagged and the maintainer runs a release
- **THEN** the next version is computed by the chosen bump level as before
