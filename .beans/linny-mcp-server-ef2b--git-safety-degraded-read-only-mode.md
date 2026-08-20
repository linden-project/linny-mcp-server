---
# linny-mcp-server-ef2b
title: Git safety & degraded read-only mode
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T18:49:42Z
parent: linny-mcp-server-vptn
blocked_by:
    - linny-mcp-server-gvc5
---

Inspect working tree before every write (unmerged entries, conflict markers, rebase/detached). Enter degraded read-only mode on unclean tree; refuse writes with retryable error; auto-exit when clean. Atomic writes (temp+fsync+rename). Optimistic concurrency via read-time content hash. Do NOT reimplement git-sync.

**OpenSpec change:** `git-safety-degraded-mode`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/git-safety-degraded-mode/tasks.md`. Ships with tests._

## Summary of Changes

Delivered internal/gitsafe: live git working-tree inspection + degraded read-only gating + safe writes. inspect() reads the git dir directly (MERGE_HEAD, rebase-merge/, rebase-apply/, CHERRY_PICK_HEAD, REVERT_HEAD, detached HEAD) and scans tracked *.md for committed conflict markers (<<<<<<< />>>>>>> only, matching the indexer); optional git ls-files -u catches unmerged index entries when git is on PATH. Guard.EnsureWritable() re-inspects every call (no sticky state) so degraded mode is entered and left automatically; DegradedError and StaleWriteError are typed and Retryable(). AtomicWrite does temp-in-same-dir -> fsync -> rename -> fsync-dir; WriteIfUnchanged adds optimistic concurrency (empty hash = create-if-absent). serve builds a Guard from --corpus/--read-only and /healthz now reports live degraded/conflicted/paths.

Verified with unit tests (clean/merge/rebase/detached/conflict, forced read-only, auto-recovery, atomic-no-temp, stale-write rejection) and a live server test: a corpus with committed markers reports degraded+conflicted with the path, and recovers to status:ok after resolution with no manual reset. nix flake check: all checks passed. Ahead/behind, sync_status() tool, and ntfy alerting are E0303.
