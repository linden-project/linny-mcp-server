## Why

The Linden Index Specification (v0.3.0) now exists; per the briefing this is the
first code that implements it. The indexer is the keystone deliverable: the DoD
requires that "the indexer runs standalone and produces JSON that `linny.vim`
accepts." Everything downstream (search, the `verify_index` tool, the read tools)
depends on a built index.

This change delivers the indexer's core — front-matter parsing, the taxonomy graph,
and spec-compliant JSON emission — with the SQLite+FTS5 persistence layer following
as its own change (it carries the heavier `modernc.org/sqlite` dependency).

## What Changes

- Add `internal/index`: parse YAML front matter directly (via `gopkg.in/yaml.v3`),
  build the taxonomy graph (terms, membership), and compute task counts and starred
  sets.
- Emit every index file from the spec into the index root: the eight home-level
  `_index_*.json` files and the nested L1 `<tax>/index.json` + L2
  `<tax>/<term>/index.json`, using the spec's naming/normalization.
- Read `lindenConfig/` L1/L2 YAML to populate the L1 term-config objects and the
  starred-taxonomy / starred-term indexes.
- Scan tracked content for committed git conflict markers and report them loudly.
- Handle title-less and malformed-front-matter records per the spec (Hugo-normative:
  title-less records are excluded from the props/title indexes; malformed records are
  reported and skipped, never crash the build).
- Wire `lindexer build --corpus <dir> --index <dir>` to run a full rebuild.

## Capabilities

### New Capabilities
- `indexer`: builds the Linden index (taxonomy graph + JSON emission) from a corpus,
  conforming to the `index-format` spec, standalone via `cmd/lindexer`.

### Modified Capabilities

## Impact

- First external dependency (`gopkg.in/yaml.v3`); `nix/package.nix` `vendorHash`
  updated from `null` to the real hash.
- New: `internal/index/**`, `lindexer build` wired.
- Round-trip / shape tests run against the synthetic corpus.
