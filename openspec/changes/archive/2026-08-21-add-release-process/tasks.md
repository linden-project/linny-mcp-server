## 1. Version source of truth

- [x] 1.1 Add `VERSION` file (`0.1.0`)
- [x] 1.2 `flake.nix` reads it (`version = lib.fileContents ./VERSION`, trimmed)
- [x] 1.3 Confirm the built binary reports it (`version` cmd, `_indexer_info`, MCP handshake)

## 2. Release script

- [x] 2.1 Preflight: clean working copy on up-to-date `main` + `nix flake check`
- [x] 2.2 `gum choose major/minor/patch`; compute next semver from `VERSION`; confirm
- [x] 2.3 Bump `VERSION`; roll `[Unreleased]` → dated section (+ fresh `[Unreleased]`)
- [x] 2.4 `jj` commit + push; then `git tag vX.Y.Z <rev>` + `git push origin` (tag the bump commit)
- [x] 2.5 `gh release create vX.Y.Z` with the new CHANGELOG section as notes
- [x] 2.6 Add `gum` + `gh` to the flake devShell

## 3. Docs

- [x] 3.1 `RELEASING.md`: VERSION contract, release command, pre-1.0 semver note, mipnix pin, secret-hygiene reminder

## 4. Verify

- [x] 4.1 `nix flake check` green (VERSION wiring + devShell change); `nix build` binary reports VERSION
- [x] 4.2 Dry-run the preflight + bump math (shellcheck + bash -n + bump-math check)

## First use (maintainer, after this ships — not part of this change)

Cut `v0.1.0` by running `scripts/release.sh`. This is an outward-facing action (a real
git tag + `gh release`), run deliberately by the maintainer, so it is intentionally not
a task of this change.

## Out of scope (design D4/D5)

- No goreleaser / uploaded artifacts; no tag-triggered `release.yml`.
- CI `ci.yml` (nix flake check on PRs) is a separate follow-up change.
