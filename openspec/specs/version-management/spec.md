# version-management Specification

## Purpose
TBD - created by archiving change add-release-process. Update Purpose after archive.
## Requirements
### Requirement: VERSION file is the single source of truth

A `VERSION` file at the repository root SHALL hold the current version, and `flake.nix`
SHALL read it (rather than a hardcoded literal) so the packaged binary is stamped from
it via the existing ldflags path.

#### Scenario: Flake reads VERSION

- **WHEN** the flake package is built
- **THEN** the built binary's `buildinfo.Version` equals the contents of `VERSION`
  (trailing newline trimmed)

#### Scenario: One place to bump

- **WHEN** the version changes
- **THEN** only `VERSION` is edited, and no other file hardcodes the version string

### Requirement: Version is reported on every product surface

The version from `VERSION` SHALL be reported by the `linny-mcp version` and `lindexer
version` subcommands, the `serve` startup banner, `_indexer_info.json`'s
`product_version`, and the MCP server's advertised implementation version.

#### Scenario: CLI reports the version

- **WHEN** `linny-mcp version` is run from a build stamped from `VERSION`
- **THEN** it prints that version

#### Scenario: Indexer info reports the version

- **WHEN** the index is emitted
- **THEN** `_indexer_info.json`'s `product_version` is that version

### Requirement: RELEASING.md documents the process

The repository SHALL include `RELEASING.md` describing the `VERSION` contract, how to
cut a release with `scripts/release.sh`, the pre-1.0 semver convention (before 1.0,
breaking changes ride a minor bump; `major` is the deliberate go-to-1.0.0 choice), and
how mipnix pins the resulting tag.

#### Scenario: Maintainer docs present

- **WHEN** `RELEASING.md` is read
- **THEN** it documents the release command, the version source, and the pre-1.0 semver
  note

