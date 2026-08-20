## Context

The in-memory `Graph` already holds everything (records with props/body/tasks,
membership, L1/L2 config, starred sets). This change gives it a persistent, queryable
form. FTS5 availability in `modernc.org/sqlite` was verified (`snippet()` + `bm25()`
work).

## Goals / Non-Goals

**Goals:** a rebuildable SQLite+FTS5 cache in `stateDir`; ranked search; taxonomy/term/
doc queries that later filter in SQL. Fast full rebuild.

**Non-Goals:** no scope filtering or redaction yet (E0401/E0402); no MCP tools (E0501);
no incremental/`fsnotify` updates (E0204) — this change does full-rebuild persistence.

## Decisions

- **`database/sql` with `MaxOpenConns(1)`.** The store is written by a single indexer
  process; one connection avoids `SQLITE_BUSY` and keeps FTS writes simple. Read-heavy
  concurrency can be revisited when the server serves many tools.
- **Disposable-cache PRAGMAs.** `synchronous=OFF` and `journal_mode=MEMORY`: a crash
  mid-rebuild just means "rebuild again", which is already the documented recovery
  path, so durability is not worth the write cost. Keeps rebuild in single-digit
  seconds.
- **Transactional recreate on `Populate`.** Each populate drops and recreates tables
  inside one transaction, then bulk-inserts. This makes re-indexing idempotent and
  tolerant of schema changes across versions — no migration machinery for a cache.
- **Schema.** `docs(filename PK, title, props_json, tasks_open/closed/total,
  starred)`; `taxonomies(name PK)`; `terms(taxonomy, term, config_json, starred, PK)`;
  `membership(taxonomy, term, filename, PK)` indexed for term→docs and doc→terms;
  `docs_fts` FTS5 `(filename UNINDEXED, title, body)`. Body lives only in FTS to avoid
  duplicating ~20 MB of text; `GetDoc` reads body from `docs_fts` and props from
  `docs`.
- **bm25 ordering.** FTS5 `bm25()` returns lower = more relevant; `ORDER BY
  bm25(docs_fts)` ascending yields best-first. Snippets via `snippet()`.
- **Query hygiene.** The user query is passed as a bound parameter to `MATCH`; callers
  get an empty result (not an error) for a query that matches nothing.

## Risks / Trade-offs

- [FTS5 query-syntax errors from raw user input] → malformed MATCH syntax can error;
  the read-tool epic will sanitize/quote user queries. For now Search surfaces the
  error to the caller.
- [Body stored only in FTS] → relies on FTS5 columns being retrievable (they are, when
  not `UNINDEXED`). Verified.
- [`synchronous=OFF`] → acceptable only because the DB is a cache; documented.

## Open Questions

- Whether to add `sqlite-vec` for local embeddings later (out of scope; v1 rejects
  cloud embeddings and defers local ones).
