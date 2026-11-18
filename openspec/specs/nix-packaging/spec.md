# nix-packaging Specification

## Purpose
TBD - created by archiving change scaffold-repo-and-flake. Update Purpose after archive.

## Requirements

### Requirement: Plain-nix flake without flake-utils

The repository SHALL provide a `flake.nix` that enumerates supported systems
explicitly as a plain Nix list (`[ "x86_64-linux" "aarch64-linux" ]`) and maps over
them with a local helper. It SHALL NOT depend on `flake-utils` or any equivalent
iteration library.

#### Scenario: No flake-utils input

- **WHEN** `flake.nix` inputs are inspected
- **THEN** the only inputs are `nixpkgs` (and optionally pinned helpers that are not
  `flake-utils`)

#### Scenario: Both systems present in outputs

- **WHEN** `nix flake show` is evaluated
- **THEN** `packages.x86_64-linux` and `packages.aarch64-linux` both expose the
  `linny-mcp` (and default) package

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

### Requirement: Overlay export

The flake SHALL export `overlays.default` adding `linny-mcp` to a nixpkgs package
set, so downstream (mipnix) can consume it as an overlay.

#### Scenario: Overlay adds package

- **WHEN** the overlay is applied to a nixpkgs instance
- **THEN** `pkgs.linny-mcp` resolves to the built package

### Requirement: Hardened systemd unit

The NixOS module's systemd unit SHALL apply a hardened sandbox: `ProtectSystem=strict`
with `ReadWritePaths` limited to the notebook corpus and state directories,
`PrivateTmp`, `NoNewPrivileges`, `RestrictAddressFamilies` limited to
`AF_INET AF_INET6 AF_UNIX`, `SystemCallFilter=@system-service`, an empty
`CapabilityBoundingSet`, `LockPersonality`, and `MemoryDenyWriteExecute`.

#### Scenario: Hardening directives present

- **WHEN** the module is evaluated into a NixOS system
- **THEN** the `linny-mcp` unit's serviceConfig includes the hardening directives
  above, with `ReadWritePaths` covering each notebook's corpus and state dir

### Requirement: ntfy option surfaced without secrets

The module SHALL expose an `ntfyTopicURL` option and pass it through the generated
config. Secrets SHALL NOT appear in any option: `tokensFile` remains a path.

#### Scenario: ntfy flows through config

- **WHEN** `ntfyTopicURL` is set and the module is evaluated
- **THEN** the generated config JSON carries that URL and no token value appears in
  any option
