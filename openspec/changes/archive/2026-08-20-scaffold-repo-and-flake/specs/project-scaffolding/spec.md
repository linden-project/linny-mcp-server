## ADDED Requirements

### Requirement: Buildable Go module

The repository SHALL contain a Go module named
`github.com/linden-project/linny-mcp-server` targeting Go 1.26 that builds with no
external dependencies at scaffold time, so that later changes add dependencies
incrementally.

#### Scenario: Module compiles clean

- **WHEN** `go build ./...` is run at the repository root
- **THEN** it exits zero with no errors

#### Scenario: Tests run green

- **WHEN** `go test ./...` is run at the repository root
- **THEN** it exits zero

### Requirement: Command entrypoints exist

The module SHALL provide two command entrypoints: `cmd/linny-mcp` (the MCP server
binary, host of the future `gen-token` helper) and `cmd/lindexer` (the standalone
indexer CLI). Each SHALL build to a runnable binary and print its own name and a
version string when invoked with `version` or `--version`.

#### Scenario: Server binary reports version

- **WHEN** the `linny-mcp` binary is run with `version`
- **THEN** it prints a version string and exits zero

#### Scenario: Indexer binary reports version

- **WHEN** the `lindexer` binary is run with `version`
- **THEN** it prints a version string and exits zero

### Requirement: Internal package layout

The module SHALL establish the internal package layout
(`internal/{index,auth,authz,gitsafe,redact,mcp,config}`) as documented stubs so
later epics have a stable home. Each package SHALL contain at least a `doc.go`
describing its responsibility.

#### Scenario: Packages are importable

- **WHEN** `go vet ./...` is run
- **THEN** every internal package compiles and vets clean

### Requirement: README states the secret-hygiene rule

`README.md` SHALL prominently state that no token value may ever appear in a Nix
option (Nix options land world-readable in `/nix/store`) and that the NixOS module
takes a `tokensFile` path sourced from `age.secrets`.

#### Scenario: Secret-hygiene rule present

- **WHEN** `README.md` is read
- **THEN** it contains the rule that tokens are provided via a `tokensFile` path,
  never as a Nix option value, with the `/nix/store` rationale

### Requirement: Disposable index is git-ignored

`.gitignore` SHALL exclude the disposable index/state directories and Nix/direnv
build artifacts so the index is never committed.

#### Scenario: State dir ignored

- **WHEN** a file is created under the configured state directory (e.g.
  `.linny_temp/` or `result/`, `.direnv/`)
- **THEN** `git status` does not list it as tracked or untracked-to-commit
