---
# linny-mcp-server-ef2b
title: Git safety & degraded read-only mode
status: todo
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:00:53Z
parent: linny-mcp-server-vptn
blocked_by:
    - linny-mcp-server-gvc5
---

Inspect working tree before every write (unmerged entries, conflict markers, rebase/detached). Enter degraded read-only mode on unclean tree; refuse writes with retryable error; auto-exit when clean. Atomic writes (temp+fsync+rename). Optimistic concurrency via read-time content hash. Do NOT reimplement git-sync.

**OpenSpec change:** `git-safety-degraded-mode`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/git-safety-degraded-mode/tasks.md`. Ships with tests._
