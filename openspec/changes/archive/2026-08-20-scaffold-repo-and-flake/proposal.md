## Why

Nothing in this PoC can be built, tested, or packaged until the repository has a
compilable Go module, a reproducible Nix development environment, and a flake that
`buildGoModule` can produce a package from. The briefing also requires the
secret-hygiene rule (no token value in any Nix option) to be stated prominently in
the README before any auth work begins. This change lays that foundation.

## What Changes

- Create the Go module (`github.com/linden-project/linny-mcp-server`, Go 1.26) and
  the two command entrypoints as buildable stubs: `cmd/linny-mcp` (server + future
  `gen-token`) and `cmd/lindexer` (standalone indexer CLI).
- Establish the `internal/` package layout as documented package stubs so later
  epics have a home: `index`, `auth`, `authz`, `gitsafe`, `redact`, `mcp`, `config`.
- Add a **plain-nix** flake (NO flake-utils) that enumerates `x86_64-linux` and
  `aarch64-linux` explicitly and exports `packages`, `overlays`, `devShells`, and
  `checks`. `nixosModules` is stubbed here and fleshed out in milestone 06.
- `devShells.default`: Go toolchain, `golangci-lint`, and `hugo` (needed later for
  the verify path).
- `checks.<system>`: run `go test ./...` and `golangci-lint` so `nix flake check`
  gates this build's own output from day one.
- Rewrite `README.md` to state the project's purpose and the §5.4 secret-hygiene
  rule prominently.
- Extend `.gitignore` to exclude the disposable index/state dirs and Nix/direnv
  artifacts so the index is never committed.

## Capabilities

### New Capabilities
- `project-scaffolding`: the buildable Go module, command entrypoints, internal
  package layout, and the README secret-hygiene contract.
- `nix-packaging`: the plain-nix flake exporting packages, overlays, devShells, and
  checks for the two supported Linux systems.

### Modified Capabilities

## Impact

- New files: `go.mod`, `cmd/**`, `internal/**/doc.go`, `flake.nix`, `nix/**`,
  updated `README.md` and `.gitignore`.
- Introduces the toolchain contract every later change depends on. No runtime
  behaviour yet.
