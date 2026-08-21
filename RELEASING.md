# Releasing linny-mcp

A release is **a git tag**. mipnix consumes this repository as a Nix flake input
built from source, so there is nothing to compile-and-upload — no goreleaser, no
prebuilt binaries, no tag-triggered Actions.

## Version source of truth

`VERSION` (repo root) holds the current version. `flake.nix` reads it
(`version = lib.fileContents ./VERSION`) and stamps it into the binary via the
package `-ldflags` → `internal/buildinfo.Version`, which is what the `version`
subcommand, the `serve` banner, `_indexer_info.json`'s `product_version`, and the MCP
server's advertised implementation version all report. **Bump only `VERSION`** — never
hardcode a version anywhere else. (Dev builds via `go run` report `0.0.0-dev`; only Nix
builds are stamped.)

## Cutting a release

```sh
nix develop            # provides gum + gh
scripts/release.sh     # or: scripts/release.sh patch|minor|major
```

The script, in order:

1. **Preflight (gate — never bypassed):** the working copy must be clean, `main` must
   be in sync with `origin`, and `nix flake check` must pass.
2. **Bump:** `gum choose major/minor/patch`, computes the next semver from `VERSION`,
   asks to confirm.
3. **Record:** writes the new `VERSION` and rolls the CHANGELOG `[Unreleased]` section
   into a dated `[X.Y.Z]` section (leaving a fresh `[Unreleased]`).
4. **Commit + tag:** `jj commit` + `jj git push` the bump, then `git tag vX.Y.Z` on
   that commit and `git push origin vX.Y.Z` (tags go through the colocated git —
   `jj git push` does not push tags). The tag points at the bump commit, so a build at
   the tag reports the right version.
5. **Publish:** `gh release create vX.Y.Z` with the new CHANGELOG section as the notes
   (needs local `gh` auth).

`CHANGELOG.md` is kept current by `scripts/ship-change.sh` as each change ships (it
appends to `[Unreleased]`), so the release notes are ready at release time.

## Semver before 1.0

The bump chooser maps literally (`patch 0.1.0→0.1.1`, `minor →0.2.0`, `major →1.0.0`),
but while the project is pre-1.0 (alpha): **breaking changes ride a `minor` bump**, and
`major` is the deliberate "declare stable → 1.0.0" choice. Alpha status is communicated
here and in the README, not with a `-alpha` version suffix (a prerelease suffix would
break the bump math).

## Consuming a release in mipnix

```sh
# in mipnix, after a new tag exists:
nix flake update linny-mcp        # repins flake.lock to the new tag/commit
nixos-rebuild switch --flake .#dapperehaan
```

Watch `/healthz` (and ntfy) after the switch. Secret hygiene is unchanged: device
tokens come from `age.secrets` as a `tokensFile` **path** — never a token value in a
Nix option.
