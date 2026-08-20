## Context

Greenfield repo with only `README.md`, `LICENSE`, and `.gitignore`. The briefing
(§9) fixes the packaging shape: a Nix flake with `packages`, `nixosModules`,
`overlays`, `devShells`, `checks`, built via `buildGoModule`, and — a hard user
constraint — **plain nix, no flake-utils**, enumerating systems explicitly. The
SQLite decision (`modernc.org/sqlite`, pure Go) means the build stays cgo-free, so
`buildGoModule` needs no special handling.

## Goals / Non-Goals

**Goals:**
- A module that builds and tests green with zero external dependencies, so
  `vendorHash` can be `null` until a real dependency is added.
- A flake whose `checks` already run `go test` + lint, so the DoD's
  `nix flake check` gate is live from the first change.
- The internal package layout carved out so later epics slot in without churn.
- README carries the secret-hygiene rule before any auth code exists.

**Non-Goals:**
- No MCP SDK, SQLite, or fsnotify dependency yet (added by their epics).
- No real `nixosModules.linny-mcp` implementation — a stub only (milestone 06).
- No `gen-token` logic — the subcommand exists but is a stub.

## Decisions

- **Plain-nix system iteration.** Define `systems = [ "x86_64-linux"
  "aarch64-linux" ]` and a local `forAllSystems = f: nixpkgs.lib.genAttrs systems
  (system: f system)` helper. `nixpkgsFor.<system> = import nixpkgs { inherit
  system; overlays = [ self.overlays.default ]; }`. This is the idiomatic
  flake-utils-free pattern and keeps the two systems explicit and greppable.
- **`vendorHash = null`.** With no external modules, buildGoModule needs no vendor
  hash. The first change that adds a dependency updates this to the real hash — a
  documented, expected step, not a surprise.
- **`ldflags` version stamping.** Inject the version via
  `-X main.version=<v>` from a single `version` var read by both binaries through a
  shared `internal/buildinfo` value, so `version` output is consistent.
- **Checks reuse the package build.** `checks.<system>.gotest` runs `go test ./...`
  in a `buildGoModule`-style derivation; `checks.<system>.lint` runs
  `golangci-lint run`. Both are also runnable from `nix develop` for the local loop.
- **golangci-lint config** kept minimal and non-pedantic at scaffold time
  (`govet`, `staticcheck`, `errcheck`, `ineffassign`, `unused`) so it does not block
  legitimate stubs; tightened per-epic as real code lands.

## Risks / Trade-offs

- [Offline `nix flake check`] → nixpkgs must be fetchable to evaluate the flake. In
  a sandbox without network this can fail for environmental reasons unrelated to the
  code. Mitigation: keep `go build`/`go test` runnable directly as the primary local
  gate; treat `nix flake check` as the CI/reproducibility gate.
- [Stub packages tripping `unused`/lint] → each `internal/*` package ships a
  `doc.go` with only a package clause and doc comment (no unused symbols), avoiding
  lint noise.
- [`vendorHash = null` drift] → the moment a dep is added, the null hash breaks the
  build loudly, which is the desired fail-fast; documented in the flake comment.

## Open Questions

- Final module path is assumed `github.com/linden-project/linny-mcp-server` (matches
  the git remote). Confirm if the canonical import path should be the shorter
  `linny-mcp` suggested in the briefing.
