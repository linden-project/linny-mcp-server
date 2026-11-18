# nix-packaging Specification

## MODIFIED Requirements

### Requirement: Package built via buildGoModule

The flake SHALL build the MCP server via `buildGoModule` so cross-compilation and
packaging stay cgo-free (consistent with the `modernc.org/sqlite` decision). The
package SHALL contain the `linny-mcp` binary only; `lindexer` ships from the
`linden-index` flake.

#### Scenario: Package builds

- **WHEN** `nix build .#linny-mcp` is run on a supported system
- **THEN** it produces a `result/bin/linny-mcp` executable
- **AND** `result/bin/lindexer` is absent

### Requirement: Development shell

The flake SHALL export `devShells.default` providing the Go toolchain,
`golangci-lint`, and `git` (the test suite shells out to git). Hugo SHALL NOT be
required, because the Hugo reference comparison moved to `linden-index`.

#### Scenario: Dev shell provides tools

- **WHEN** `nix develop` is entered
- **THEN** `go`, `golangci-lint`, and `git` are all on `PATH`

### Requirement: Flake checks gate the build

The flake SHALL export `checks.<system>` that run `go test ./...` with the coverage
gate and `golangci-lint`, so `nix flake check` fails when tests, coverage, or lint
fail. The Hugo round-trip check SHALL NOT run here.

#### Scenario: Checks run tests and lint

- **WHEN** `nix flake check` is run
- **THEN** it executes the Go test suite with the coverage gate and the linter, and
  fails if either fails
