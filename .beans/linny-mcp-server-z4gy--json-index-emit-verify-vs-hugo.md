---
# linny-mcp-server-z4gy
title: JSON index emit + verify vs Hugo
status: todo
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:29:01Z
parent: linny-mcp-server-gupt
blocked_by:
    - linny-mcp-server-zin6
    - linny-mcp-server-sac5
---

Emit the JSON index files from the spec, byte-compatible enough that linny.vim works unmodified. 'verify' subcommand runs Hugo + ours and diffs outputs, reporting drift. Round-trip test on synthetic corpus.

**OpenSpec change:** `indexer-json-emit-verify`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/indexer-json-emit-verify/tasks.md`. Ships with tests._

## Note (2026-08-20)

JSON index emission (the emit half of this epic) was delivered in the `indexer-core` change together with E0201. What remains for this epic: the `verify` subcommand that runs Hugo + ours and diffs outputs. SQLite/FTS5 (E0202) is not required for emit — the emitter reads the in-memory graph directly.
