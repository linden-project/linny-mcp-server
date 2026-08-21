---
# linny-mcp-server-w9zx
title: 'Release process: VERSION + gum/tag/gh release automation'
status: completed
type: epic
priority: normal
created_at: 2026-08-21T12:32:58Z
updated_at: 2026-08-21T12:47:05Z
parent: linny-mcp-server-xwyf
---

Adopt the beans/huphop release pattern for a Go+flake tool consumed as a flake input: VERSION source of truth read by the flake, RELEASING.md, and an upgraded scripts/release.sh (preflight nix flake check -> gum bump -> jj commit -> git tag -> gh release). Non-goals: goreleaser/artifacts, release.yml, CI workflow (separate). **OpenSpec change:** `add-release-process`

## Summary of Changes

Added the release process (adapted from the beans/huphop pattern for a Go+flake tool consumed as a flake input). VERSION file is now the single source of truth: flake.nix reads it (lib.fileContents) and the existing ldflags path stamps it into buildinfo.Version, verified by nix build (linny-mcp 0.1.0 / lindexer 0.1.0). Upgraded scripts/release.sh to the full manual flow: preflight (clean working copy on up-to-date main + nix flake check) -> gum choose major/minor/patch -> compute next semver -> bump VERSION + roll CHANGELOG [Unreleased] -> jj commit/push -> git tag vX.Y.Z on the bump commit + git push (tags via colocated git) -> gh release create with the CHANGELOG section as notes. Added RELEASING.md and gum+gh to the devShell. No goreleaser/artifacts, no release.yml (distribution is source-via-flake); CI workflow deferred.

QA: shellcheck clean on both scripts, bash -n OK, bump-math verified (0.1.0 -> patch 0.1.1 / minor 0.2.0 / major 1.0.0), and the VERSION->flake->binary wiring confirmed by a real nix build. Shipped via scripts/ship-change.sh; nix flake check all green. First release (v0.1.0) is a deliberate maintainer action: run scripts/release.sh.
