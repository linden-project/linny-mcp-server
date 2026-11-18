## MODIFIED Requirements

### Requirement: Command entrypoints exist

The module SHALL provide one command entrypoint, `cmd/linny-mcp` (the MCP server
binary, host of the `gen-token` helper). It SHALL build to a runnable binary and print
its name and a version string when invoked with `version` or `--version`. The
`lindexer` CLI is built and released from `linden-index`.

#### Scenario: Server binary reports version

- **WHEN** the `linny-mcp` binary is run with `version`
- **THEN** it prints a version string and exits zero

#### Scenario: Indexer binary reports version

- **WHEN** an `lindexer` binary installed from `linden-index` is run with `version`
- **THEN** it prints a version string and exits zero

### Requirement: Internal package layout

The module SHALL establish the internal package layout
(`internal/{auth,authz,gitsafe,redact,mcp,config,defense,backup,alert,audit,buildinfo}`)
so later epics have a stable home. Each package SHALL contain at least a `doc.go`
describing its responsibility. The index engine lives in the `linden-index` module.

#### Scenario: Packages are importable

- **WHEN** `go vet ./...` is run
- **THEN** every internal package compiles and vets clean
