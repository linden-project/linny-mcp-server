# Proposal

## Why

The standalone indexer serves linny.vim users who do not run the MCP server and do
not use Nix. Keeping it inside the server repository ties its releases to the
server, hides it behind a server name, and cannot produce downloadable binaries.

## What Changes

- **BREAKING** (build layout): remove `cmd/lindexer`, `internal/index`, and
  `internal/hugoref`; require `github.com/linden-project/linden-index` in `go.mod`.
- The index engine becomes a public package (`index`) in the new module, because Go
  forbids importing `internal/` across module boundaries.
- `VERSION` keeps stamping the server only; the indexer version is owned by
  `linden-index`.
- The flake package builds `linny-mcp` only. The dev shell and checks drop `hugo` and
  the Hugo round-trip, which move with the indexer.
- Coverage floors stay as they are: 70% overall and 80% on core packages. The new
  repository inherits the same floors.
- Move the `indexer`, `index-format`, `index-store`, `index-verify`, `index-watch`,
  `verify-hugo`, and `synthetic-corpus` specs to the new repository and remove them
  here.
- No change to MCP tools, authentication, git safety, or write behavior.

Not in this change: bootstrapping `linden-index`, its release automation, and its
own change artifacts.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `version-management`: `VERSION` stamps the server only, not the indexer.
- `nix-packaging`: the package ships only the server; hugo leaves the dev shell and
  checks.
- `dev-workflow`: the coverage gate lists only the core packages that remain.
- `project-scaffolding`: one command entrypoint remains; the index engine is external.
- `indexer`, `index-format`, `index-store`, `index-verify`, `index-watch`,
  `verify-hugo`, `synthetic-corpus`: all requirements removed; ownership moves to
  `linden-index`.

## Impact

- Go: `go.mod`, `go.sum`; deleted `cmd/lindexer`, `internal/index`,
  `internal/hugoref`; import paths in `internal/mcp`, `internal/redact`, and tests.
- Packaging: `flake.nix`, `nix/package.nix`, `RELEASING.md`.
- Docs and release: `README.md`, `CHANGELOG.md`, `docs/linden-index-spec.md` and the
  index-related docs move out.
- Downstream: mipnix gains a second flake input; linny.vim consumes released
  `lindexer` binaries.
