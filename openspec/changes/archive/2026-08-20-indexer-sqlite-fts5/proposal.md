## Why

The indexer currently emits linny.vim-compatible JSON, but the server needs a
queryable index for the MCP read tools: ranked full-text search, taxonomy/term
lookups, and — crucially for the authorization epic — the ability to **filter in SQL
at query time** rather than post-filtering. The briefing fixes the storage: SQLite +
**FTS5**, via `modernc.org/sqlite` (pure Go, no cgo), living in `stateDir` as a
**disposable cache** that is never committed and can always be rebuilt from the corpus.

## What Changes

- Add `modernc.org/sqlite` (pure Go). FTS5 is compiled in (verified).
- Add `internal/index.Store`: opens/creates a SQLite database in `stateDir`, holds the
  schema (docs, taxonomies, terms, membership, and a `docs_fts` FTS5 virtual table),
  and persists a built `Graph` transactionally.
- **Rebuild-safe**: `Populate` recreates the schema inside a transaction every run, so
  re-indexing an existing DB is always clean and deleting the DB and rebuilding is a
  valid, documented recovery step.
- Add query methods that become the substrate for the read tools (E0501):
  `Search` (FTS5 `MATCH`, ranked by `bm25`, with snippets), `ListTaxonomies`,
  `TermsForTaxonomy`, `DocsByTerm`, and `GetDoc`.
- Wire `lindexer build --state-dir <dir>` to also persist the SQLite store (JSON
  emission via `--index` is unchanged), and add `lindexer search --state-dir <dir>
  <query>` to demonstrate FTS5 from the CLI.
- Tune the connection for a disposable cache (single writer, `synchronous=OFF`,
  in-memory journal) so a full rebuild stays in single-digit seconds.

## Capabilities

### New Capabilities
- `index-store`: SQLite+FTS5 persistence of the taxonomy graph and full text, with
  ranked search and taxonomy/term/document queries, as a rebuildable disposable cache.

### Modified Capabilities

## Impact

- New external dependency `modernc.org/sqlite`; `nix/package.nix` `vendorHash` updated.
- New: `internal/index/store.go`, `internal/index/query.go`; `lindexer build
  --state-dir` and `lindexer search` wired.
- Foundation for E0501 (read tools) and E0401 (SQL-level scope filtering).
