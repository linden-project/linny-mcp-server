---
# linny-mcp-server-j0ee
title: Tested backup/restore path
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:29Z
updated_at: 2026-08-20T21:06:32Z
parent: linny-mcp-server-xwyf
blocked_by:
    - linny-mcp-server-fe16
---

Because the server can delete, ship a VERIFIED restore path: backup command/tool + restore, with a test proving a delete can be recovered. Mandatory, not optional.

**OpenSpec change:** `backup-restore`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/backup-restore/tasks.md`. Ships with tests._

## Summary of Changes

Added internal/backup: Backup(root, w) writes a tar.gz snapshot of the source-of-truth data (content/ + lindenConfig/), excluding the disposable index/state and VCS dirs; Restore(r, root) extracts it back with path sanitization (rejects absolute paths and ../ traversal so a malicious archive cannot escape the target). Wired linny-mcp backup --corpus --out and linny-mcp restore --in --corpus.

Verified (the DoD-mandated tested restore): backup contains content+config and NOT .git/lindenIndex; a delete-then-restore recovers the record byte-for-byte; a mutated record is restored to the backed-up version; a crafted ../ traversal entry is rejected and writes nothing outside the target. Coverage 77.5 percent; nix flake check all passed.
