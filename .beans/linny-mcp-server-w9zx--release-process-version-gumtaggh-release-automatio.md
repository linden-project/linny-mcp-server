---
# linny-mcp-server-w9zx
title: 'Release process: VERSION + gum/tag/gh release automation'
status: in-progress
type: epic
priority: normal
created_at: 2026-08-21T12:32:58Z
updated_at: 2026-08-21T12:40:15Z
parent: linny-mcp-server-xwyf
---

Adopt the beans/huphop release pattern for a Go+flake tool consumed as a flake input: VERSION source of truth read by the flake, RELEASING.md, and an upgraded scripts/release.sh (preflight nix flake check -> gum bump -> jj commit -> git tag -> gh release). Non-goals: goreleaser/artifacts, release.yml, CI workflow (separate). **OpenSpec change:** `add-release-process`
