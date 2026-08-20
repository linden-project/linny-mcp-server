---
# linny-mcp-server-gex4
title: sync_status, /healthz & ntfy alerting
status: todo
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:00:53Z
parent: linny-mcp-server-vptn
blocked_by:
    - linny-mcp-server-ef2b
---

sync_status() MCP tool + /healthz endpoint: last sync, ahead/behind, conflicted bool + paths, degraded flag. Alert via self-hosted ntfy on Hetzner to phone. Never write alerts into the corpus.

**OpenSpec change:** `sync-status-healthz-alerts`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/sync-status-healthz-alerts/tasks.md`. Ships with tests._
