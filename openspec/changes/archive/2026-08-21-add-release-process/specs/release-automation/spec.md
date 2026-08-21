## ADDED Requirements

### Requirement: Manual gum-driven release script

`scripts/release.sh` SHALL cut a release manually: it prompts for the bump level with a
`gum` chooser (`major`, `minor`, `patch`), computes the next version from `VERSION` by
semantic versioning, and requires confirmation before making any change. `gum` and `gh`
SHALL be available in the flake devShell.

#### Scenario: Choosing a bump computes the next version

- **WHEN** the maintainer runs `scripts/release.sh` and selects a bump level
- **THEN** the next version is computed from `VERSION` (`patch` → z+1, `minor` → y+1.0,
  `major` → x+1.0.0) and shown for confirmation

### Requirement: Release preflight gate

`scripts/release.sh` SHALL abort before making any change unless the preflight passes: a
clean working copy on an up-to-date `main`, and a green `nix flake check`.

#### Scenario: A failing gate aborts the release

- **WHEN** the working copy is dirty, not on `main`, or `nix flake check` fails
- **THEN** the script exits non-zero without editing `VERSION`/`CHANGELOG.md` or creating
  a tag or release

### Requirement: Release bumps, tags the release commit, and publishes

On confirmation, `scripts/release.sh` SHALL update `VERSION`, roll the `[Unreleased]`
CHANGELOG section into a dated section for the new version (leaving a fresh
`[Unreleased]`), commit and push with `jj`, then create and push the `vX.Y.Z` tag
through the colocated git, and create the GitHub release with `gh release create` using
the new CHANGELOG section as the notes. There SHALL be no GitHub Actions release
workflow and no uploaded build artifacts.

#### Scenario: A release is cut end to end

- **WHEN** the maintainer confirms
- **THEN** `VERSION` and `CHANGELOG.md` are updated, a `jj` commit is pushed, a `vX.Y.Z`
  tag is pushed via git, and a GitHub release is created from the CHANGELOG section

#### Scenario: The tag points at the version-bump commit

- **WHEN** the tag is created
- **THEN** it points at the commit that bumped `VERSION`, so a build at the tag reports
  that version (the bump+commit happens before the tag)

#### Scenario: The tag is pushed through git, not jj

- **WHEN** the release commit exists
- **THEN** the tag is created and pushed via the colocated git (`git tag` /
  `git push origin vX.Y.Z`), since `jj git push` does not push tags
