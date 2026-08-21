## Context

Lineage: **huphop** (Go binary — goreleaser, `go:embed`, `vendorHash`, cross-compiled
artifacts, tag-triggered `release.yml`) → **beans.nvim** (pure-Lua plugin — no build, a
manual `gum` script cuts a `git tag` + `gh release`) → **linny-mcp**. linny-mcp is a
hybrid: a Go binary (keep huphop's binary version-stamping) that is distributed as a Nix
**flake input** built from source by mipnix (so adopt beans' "a release is a git tag").

Current state (grounded): the version already flows `flake.nix version` → package
`-ldflags -X …buildinfo.Version` → `buildinfo.Version` (default `0.0.0-dev`) → the
`version` subcommand, the `serve` banner, `_indexer_info.product_version`, and the MCP
`Implementation.Version`. `CHANGELOG.md` with `[Unreleased]` and `scripts/ship-change.sh`
already exist. Missing: a `VERSION` file, `RELEASING.md`, and a real `release.sh`.

## Goals / Non-Goals

**Goals:** one repeatable, gated, manual command to cut a release; a single version
source legible in-repo and in every product surface; the tag drives mipnix's flake pin.

**Non-Goals:** goreleaser / uploaded binaries (consumers build from source via the
flake); a tag-triggered `release.yml`; a CI workflow (separate future change); any
runtime behaviour change (the version surfaces already exist).

## Decisions

- **D1 — `VERSION` is the source of truth.** `flake.nix` uses
  `version = lib.strings.trim (lib.fileContents ./VERSION)` (or `fileContents`, which
  already drops the trailing newline). The existing ldflags path then stamps it into
  `buildinfo.Version`; no code change is needed for the CLI/index/MCP surfaces.
- **D2 — First value `0.1.0`, tag `v0.1.0` (drop `-alpha`).** A semver prerelease suffix
  breaks `gum`'s next-version math and ordering; alpha status is communicated in the
  README/CHANGELOG, not the number.
- **D3 — Keep `gh release create`.** mipnix only needs the git tag (it pins a ref), but a
  GitHub release gives human-readable notes and a visible version list at low cost.
  Requires local `gh` auth (documented in RELEASING.md). Revisable — a plain tag would
  also suffice.
- **D4 — No artifacts, no `release.yml`.** Distribution is source-via-flake; there is
  nothing to compile-and-upload and no on-tag Actions job.
- **D5 — CI is a separate change.** A `.github/workflows/ci.yml` running `nix flake
  check` on PRs is valuable but orthogonal to releasing; the release preflight already
  runs the gate locally.
- **Preflight = one gate.** Unlike beans (stylua + luacheck + smoke + helptags), linny's
  preflight collapses to a clean-working-copy-on-`main` check plus `nix flake check`
  (build + tests + lint + coverage floors + Hugo zero-drift round-trip).
- **Tags go through the colocated git.** `jj git push` does not push tags, so after the
  `jj` commit/push the script runs `git tag vX.Y.Z <rev>` + `git push origin vX.Y.Z`.

## Risks / Trade-offs

- **VERSION/tag ordering hazard (load-bearing).** mipnix builds the flake *at the tag*,
  which reads `VERSION`, which stamps `buildinfo.Version`. The script MUST bump+commit
  `VERSION` first and then tag *that* commit; reversed, `linny-mcp version` at `v0.1.0`
  would report the previous number. Encoded as a spec scenario.
- **Local `gh` auth required** to cut a release (D3). Documented; the tag itself is what
  mipnix consumes, so a failed `gh release` does not block deployment.
- **CHANGELOG discipline is manual** — release notes are only as good as the
  `[Unreleased]` section that `ship-change.sh` accumulates; RELEASING.md makes keeping it
  current part of the flow.

## Open Questions

- Add `ci.yml` (nix flake check on PRs) as the immediate follow-up? (D5 defers it.)
- Publish to any registry beyond the git tag (e.g. a flake registry entry)? Not needed
  for mipnix; deferred.
