---
# linny-mcp-server-z4gy
title: JSON index emit + verify vs Hugo
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T21:02:30Z
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

## Summary of Changes

Completed the emit+verify epic: added index.VerifyDirs(ours, reference) — walks both index trees and reports files present on only one side plus per-file content differences. Arrays are compared as SETS (order-insensitive, per spec section 3), objects key-by-key, scalars canonically; _indexer_info.json ignores the identity/environment fields (product_name/version/hugo_version and the index_dir/content_dir/config_dir paths Hugo emits as TODO). Wired lindexer verify --corpus --reference: builds our index, diffs it against a reference index tree (e.g. Hugo output), prints each DRIFT line, and exits non-zero on any discrepancy. Producing the Hugo reference needs the notebook layouts + hugo binary (dev shell) and is documented in docs/future.md; verify consumes a reference tree so it is testable and usable regardless of how the reference was built.

Verified: differ tests (identical trees clean, reordered array equal, changed membership + missing file reported, indexer_info identity/TODO ignored) and CLI tests (verify matches a freshly-built reference; requires --reference). This is the drift safety net for the own-index decision. Coverage 79.6 percent; nix flake check all passed.
