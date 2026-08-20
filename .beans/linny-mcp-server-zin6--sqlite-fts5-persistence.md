---
# linny-mcp-server-zin6
title: SQLite + FTS5 persistence
status: todo
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:00:53Z
parent: linny-mcp-server-gupt
blocked_by:
    - linny-mcp-server-r7gl
---

Persist the graph + full text to SQLite+FTS5 in stateDir using modernc.org/sqlite (no cgo). Disposable cache; deleting stateDir and rebuilding must always be a valid, documented recovery step. Full rebuild single-digit seconds for ~5k notes.

**OpenSpec change:** `indexer-sqlite-fts5`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/indexer-sqlite-fts5/tasks.md`. Ships with tests._
