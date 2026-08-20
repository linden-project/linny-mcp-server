---
# linny-mcp-server-mv29
title: Incremental updates & watch
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T21:10:57Z
parent: linny-mcp-server-gupt
blocked_by:
    - linny-mcp-server-zin6
---

Incremental index updates via fsnotify. build/watch/verify CLI subcommands. Keep rebuilds cheap.

**OpenSpec change:** `indexer-incremental-watch`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/indexer-incremental-watch/tasks.md`. Ships with tests._

## Summary of Changes

Added lindexer watch: internal/index.Watch watches content/ + lindenConfig via fsnotify and coalesces bursts with a debounce (default 300ms), invoking a rebuild callback once per quiet period. lindexer watch --corpus --state-dir [--index] builds once, then refreshes the SQLite store (and re-emits JSON when --index is set) on each change batch, until SIGINT/SIGTERM. Debounced FULL rebuild (cheap at PoC scale); true per-file incremental diffing is noted in docs/future.md.

Verified: deterministic debounce test (a 5-signal burst fires exactly once), an fsnotify integration test (writing a record triggers the rebuild callback within timeout), and watch requires --state-dir. Adds github.com/fsnotify/fsnotify; nix vendorHash updated. nix flake check all passed.
