---
# linny-mcp-server-sac5
title: Synthetic corpus generator
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:21:29Z
parent: linny-mcp-server-tq36
blocked_by:
    - linny-mcp-server-0gdv
---

Generate a few thousand realistic flat notes with multi-taxonomy YAML front matter, dates, booleans, $FORMAT/$INCLUDE, and deliberate edge cases (unicode, huge front matter, empty bodies, malformed YAML, committed conflict markers, embedded fake secrets). NEVER touch the real secondbrain repo.

**OpenSpec change:** `synthetic-corpus-generator`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/synthetic-corpus-generator/tasks.md`. Ships with tests._

## Summary of Changes

Added internal/corpus: a deterministic synthetic Linny notebook generator (same seed => byte-identical tree, no wall clock). Writes a flat content/ dir of realistic records (plural taxonomy keys tags/projects/customer/type/subject, crdate, starred, task lists), a consistent lindenConfig/ (L1/L2 configs with some starred), and a Hugo config.yaml so the reference indexer can build the same corpus for verify. Edge cases: unicode, long front matter, empty body, malformed YAML, committed conflict markers, fake secrets (AWS/PEM/token/IBAN shaped), and a work+health doc for scope-intersection tests. cmd/gen-corpus materializes it. Tests cover determinism, flat structure, config consistency, and edge-case presence.

NEVER derived from real data. nix flake check: all checks passed.
