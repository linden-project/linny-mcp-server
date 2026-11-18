# version-management Specification

## MODIFIED Requirements

### Requirement: VERSION file is the single source of truth

A `VERSION` file at the repository root SHALL hold the current version of the MCP
server, and `flake.nix` SHALL read it (rather than a hardcoded literal) so the
packaged `linny-mcp` binary is stamped from it via the existing ldflags path. The
`lindexer` version SHALL come from the `linden-index` module, not from this file.

#### Scenario: Flake reads VERSION

- **WHEN** the flake package is built
- **THEN** the built binary's `buildinfo.Version` equals the contents of `VERSION`
  (trailing newline trimmed)

#### Scenario: One place to bump

- **WHEN** the version changes
- **THEN** only `VERSION` is edited, and no other file hardcodes the version string

#### Scenario: Indexer version is independent

- **WHEN** the `linny-mcp` version is bumped
- **THEN** the `lindexer` version reported by the dependency is unchanged
- **AND** bumping the indexer does not require editing this repository's `VERSION`

### Requirement: Version is reported on every product surface

The version from `VERSION` SHALL be reported by the `linny-mcp version` subcommand,
the `serve` startup banner, and the MCP server's advertised implementation version.
The `lindexer` version is owned by `linden-index` and is not reported by this
repository's packaging.

#### Scenario: CLI reports the version

- **WHEN** `linny-mcp version` is run from a build stamped from `VERSION`
- **THEN** it prints that version

#### Scenario: Indexer info reports the version

- **WHEN** the index is emitted
- **THEN** `_indexer_info.json`'s `product_version` is the `linden-index` version,
  which this repository neither sets nor overrides

#### Scenario: Serve banner reports the version

- **WHEN** `linny-mcp serve` starts
- **THEN** the startup banner names the version from `VERSION`
