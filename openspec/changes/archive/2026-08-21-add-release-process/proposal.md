## Why

The version is already stamped end-to-end (`flake.nix version` → `-ldflags` →
`internal/buildinfo.Version` → the `version` subcommand, the `serve` banner,
`_indexer_info.product_version`, and the MCP server's advertised `Implementation`),
but its source of truth is a **hardcoded string in `flake.nix`**, and `scripts/
release.sh` is a stub (takes a manual version arg and only rolls the CHANGELOG — no
version bump, no tag, no GitHub release). There is no `RELEASING.md`.

This adapts the beans.nvim `add-release-process` pattern (itself derived from the
huphop Go-binary pattern) for linny-mcp's shape: it **is** a Go binary (so keep the
version-stamping into the binary), but it is **distributed as a Nix flake input** built
from source by mipnix — so, like beans, **a release is just a git tag**. No goreleaser,
no uploaded artifacts, no tag-triggered `release.yml`.

## What Changes

- Add a `VERSION` file as the single source of truth; `flake.nix` reads it
  (`version = lib.fileContents ./VERSION`), so the existing ldflags path stamps it into
  `buildinfo.Version` and everything downstream reports it unchanged. First value
  `0.1.0` (dropping the `-alpha` suffix — status lives in the README/CHANGELOG, and a
  prerelease suffix breaks the bump math).
- Add `RELEASING.md` (maintainer docs: the VERSION contract, how to cut a release, the
  pre-1.0 semver convention, how mipnix pins the tag, and the secret-hygiene reminder).
- Upgrade `scripts/release.sh` to the full manual flow: preflight (clean working copy on
  an up-to-date `main` + `nix flake check`) → `gum choose major/minor/patch` → compute
  the next semver from `VERSION`, confirm → bump `VERSION` and roll `[Unreleased]` into a
  dated section → `jj` commit + push → **`git tag vX.Y.Z` + `git push origin` (via the
  colocated git, since `jj git push` does not push tags)** → `gh release create` with the
  new CHANGELOG section as notes.
- Add `gum` and `gh` to the flake devShell.

## Capabilities

### New Capabilities
- `version-management`: `VERSION` as the single source of truth (read by the flake,
  stamped into the binary, reported by the CLI / index / MCP handshake) and the
  `RELEASING.md` contract.
- `release-automation`: the manual, gated `gum`-driven `release.sh` (bump → jj commit →
  git tag → `gh release create`) and `gum`/`gh` in the devShell.

### Modified Capabilities

## Impact

- New: `VERSION`, `RELEASING.md`. Modified: `flake.nix` (read `VERSION`; `gum`+`gh` in
  devShell), `scripts/release.sh` (full flow). No production-code change; the version
  wiring is already in place.
- Non-goals (explicitly out): goreleaser / prebuilt artifacts, a tag-triggered
  `release.yml`, and a CI workflow — the QA preflight is the single existing
  `nix flake check`; a CI `ci.yml` running that gate is a separate future change.
- First release: `v0.1.0`, cut by running `scripts/release.sh`.
