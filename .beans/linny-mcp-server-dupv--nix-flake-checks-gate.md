---
# linny-mcp-server-dupv
title: nix flake checks gate
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T20:23:38Z
parent: linny-mcp-server-xwyf
blocked_by:
    - linny-mcp-server-0gdv
---

checks.{system}: go test + golangci-lint so 'nix flake check' gates this build's own output. Must pass end-to-end.

**OpenSpec change:** `nix-flake-checks`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/nix-flake-checks/tasks.md`. Ships with tests._

## Summary of Changes

Added a coverage gate to nix flake check and documented the testing agreement. checks.<system>.coverage runs the whole suite with -coverpkg=./... (so e2e tests count toward the packages they exercise) and FAILS below 70%; git is on PATH in that check so the git-backed history/gitsafe tests run rather than skip, and the coverage profile is kept as a build output. Added CLI e2e tests (cmd/lindexer build+search, cmd/linny-mcp serve validation) for headroom. openspec/project.md now states what unit vs e2e mean here and the 70% floor.

**Measured total statement coverage: 82.0%** (floor 70%). nix flake check now runs three checks — gotest + lint + coverage — all green. Inherently-untestable glue (main() wrappers, the long-running serve listen loop) is knowingly excluded from expectations.
