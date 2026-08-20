---
# linny-mcp-server-dupv
title: nix flake checks gate
status: todo
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:00:53Z
parent: linny-mcp-server-xwyf
blocked_by:
    - linny-mcp-server-0gdv
---

checks.{system}: go test + golangci-lint so 'nix flake check' gates this build's own output. Must pass end-to-end.

**OpenSpec change:** `nix-flake-checks`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/nix-flake-checks/tasks.md`. Ships with tests._
