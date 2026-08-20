---
# linny-mcp-server-2es7
title: Operational tools
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T21:17:12Z
parent: linny-mcp-server-s9mf
blocked_by:
    - linny-mcp-server-z4gy
    - linny-mcp-server-gex4
---

sync_status() and verify_index() surfaced as MCP tools. Wire to the indexer verify path and git-safety status.

**OpenSpec change:** `mcp-operational-tools`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/mcp-operational-tools/tasks.md`. Ships with tests._

## Summary of Changes

Added the verify_index operational MCP tool, completing the operational surface (sync_status shipped earlier). Added index.Store.AllDocFilenames(). verify_index rebuilds the taxonomy graph from the corpus and compares its document set to the served store, returning in_sync plus corpus_docs, store_docs, missing_from_store (in corpus, not indexed), stale_in_store (indexed, gone from corpus), and conflicted paths — a runtime "is the served index fresh?" check that needs no Hugo. (The Hugo-reference JSON diff remains the lindexer verify CLI.) docs/tools.md updated.

Verified: in-sync store reports in_sync with the conflict-marker path listed; adding an unindexed record reports it under missing_from_store and in_sync=false. Coverage 77.0 percent; nix flake check all passed.
