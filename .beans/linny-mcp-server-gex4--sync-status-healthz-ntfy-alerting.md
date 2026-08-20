---
# linny-mcp-server-gex4
title: sync_status, /healthz & ntfy alerting
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T20:48:51Z
parent: linny-mcp-server-vptn
blocked_by:
    - linny-mcp-server-ef2b
---

sync_status() MCP tool + /healthz endpoint: last sync, ahead/behind, conflicted bool + paths, degraded flag. Alert via self-hosted ntfy on Hetzner to phone. Never write alerts into the corpus.

**OpenSpec change:** `sync-status-healthz-alerts`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/sync-status-healthz-alerts/tasks.md`. Ships with tests._

## Summary of Changes

Added the sync_status MCP tool and ntfy degraded-mode alerting. sync_status returns the live git-safety state (degraded/conflicted/paths/in-progress/detached/read-only/reason) driven by the guard; registered for any authenticated caller. internal/alert: an Alerter interface with NtfyAlerter (HTTP POST title+body to a self-hosted topic) and NopAlerter, plus a DegradedMonitor that fires exactly one alert on the clean to degraded transition and a recovery notice on the way back (never re-alerting while degraded, never touching the corpus). config gained ntfyTopicURL (a plain POST target, safe in a Nix option); serve starts a 30s background poller with the ntfy alerter when configured else no-op.

Verified: monitor one-alert-per-degradation + recovery via a fake alerter; NtfyAlerter issues the HTTP POST (httptest); sync_status reports degraded+paths on a conflicted tree and healthy on clean. Coverage 79.8 percent; nix flake check gotest+lint+coverage all passed. Follow-ups: verify_index tool (E0504) and the ntfy NixOS option wiring (E0601).
