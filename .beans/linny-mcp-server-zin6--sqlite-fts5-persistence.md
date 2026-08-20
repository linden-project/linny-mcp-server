---
# linny-mcp-server-zin6
title: SQLite + FTS5 persistence
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T19:29:48Z
parent: linny-mcp-server-gupt
blocked_by:
    - linny-mcp-server-r7gl
---

Persist the graph + full text to SQLite+FTS5 in stateDir using modernc.org/sqlite (no cgo). Disposable cache; deleting stateDir and rebuilding must always be a valid, documented recovery step. Full rebuild single-digit seconds for ~5k notes.

**OpenSpec change:** `indexer-sqlite-fts5`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/indexer-sqlite-fts5/tasks.md`. Ships with tests._

## Summary of Changes

Added internal/index.Store: SQLite+FTS5 persistence via modernc.org/sqlite (pure Go, no cgo; FTS5 verified). Schema: docs, taxonomies, terms, membership, and a docs_fts FTS5 virtual table. Populate() rebuilds transactionally (drop+recreate) so it is idempotent and delete-and-rebuild always recovers; the connection is tuned for a disposable single-writer cache (synchronous=OFF, journal_mode=MEMORY). Query API: Search (FTS5 MATCH ranked by bm25 with snippets), ListTaxonomies, TermsForTaxonomy, DocsByTerm, GetDoc (title/props/body) — the substrate for the read tools (E0501) and SQL-level scope filtering (E0401). lindexer build --state-dir persists the store alongside JSON emission; lindexer search runs FTS from the CLI.

Verified: unit tests (open/empty, populate+query, missing doc, ranked search + empty result, idempotent re-populate, delete-and-rebuild recovery) and a live CLI run (build --state-dir persisted a 56-record store; search returned bm25-ranked hits with snippet highlighting). nix vendorHash updated for the modernc tree; nix flake check: all checks passed.
