# Design

## Context

`internal/index`, `internal/hugoref`, and `cmd/lindexer` form a self-contained
index engine that the MCP server imports. The engine has no MCP dependency, but
`internal/` packages cannot be imported by other modules and the repository name
marks it as a server. Splitting it out gives `lindexer` its own release cadence and
a name that matches what it does.

## Goals and Non-Goals

Goals:

- Remove the index engine, the Hugo reference, and the `lindexer` CLI from this
  repository.
- Keep the MCP server behaviour identical.
- Keep the coverage floors.

Non-goals:

- Bootstrapping `linden-index`, its flake, or its release automation. Those are
  planned in that repository.
- Adding, removing, or changing any MCP tool.
- Changing the index format.

## Decisions

### Direction of the dependency

`linny-mcp-server` imports `github.com/linden-project/linden-index/index`. The new
repository does not import the server. This keeps the engine free of MCP types and
lets `lindexer` build standalone with `CGO_ENABLED=0`.

Alternative considered: a shared third module both repositories import. That adds a
repository and a version bump between two projects that are released together in
practice. One direction is simpler.

### Public package layout

`internal/index` becomes `index` at the module root of the new repository, so
importers write `linden-index/index`. `internal/hugoref` becomes `hugoref` and stays
internal in intent, but must be importable by its own `cmd/lindexer`; Go allows that
within one module.

### Version stamping

`VERSION` keeps stamping only `linny-mcp`. `lindexer` reports the version of the
`linden-index` module that was built, which the new repository supplies through its
own ldflags or build info. The server's version and the engine's version can
therefore diverge, which is intended.

### Coverage floors

The existing floors stay: 70 percent overall, 80 percent for core packages. The new
repository inherits both. Splitting the code lowers this repository's denominator,
so the gates are re-run after the deletion and the core package list is updated to
what remains.

### Check changes

The Hugo round-trip check depends on `internal/hugoref` and leaves with it. The dev
shell drops `hugo`. The package builds `cmd/linny-mcp` only.

### Downstream

`mipnix` consumes the flake for the server. Users who need `lindexer` install it from
the `linden-index` release. The NixOS module is untouched.

## Risks and Trade-offs

- A breaking build layout for anyone importing this repository directly. It is a
  pre-1.0 project and the import path was never public.
- Two repositories to release. They version independently, which is the point, but
  the server must bump its `linden-index` dependency to pick up engine fixes.
- Removing `internal/hugoref` removes the only automated proof that the engine
  matches Hugo in this repository. That proof moves to the new repository's checks.

## Migration

1. Land the `linden-index` extraction and release a tagged version.
2. Bump `go.mod` here and delete the moved code in one change.
3. Run `go build ./...`, `go test ./...` with coverage, and `nix flake check`.
