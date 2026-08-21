## Why

The `/mip:ship` and `/opsx:propose` workflows assume project scaffolding that was
never created: `mip:ship` calls `scripts/ship-change.sh` and expects a `CHANGELOG.md`
with an `## [Unreleased]` section, and it gates at **≥70% overall / ≥80% core**
coverage — stricter than the current check. `opsx:propose` always generates a
`design.md`, which recent hand-rolled changes skipped. This change makes both
workflows first-class so every future change ships with the full spec + QA rigor.

## What Changes

- Add `CHANGELOG.md` (Keep a Changelog format) with an `## [Unreleased]` section,
  backfilled with the user-facing history to date.
- Add `scripts/ship-change.sh <change> "<subject>"`: stage → gate (`nix flake check`)
  → archive the OpenSpec change → commit as Pim Snel (jj) → push `main`. Add
  `scripts/release.sh <version>` to promote `[Unreleased]` to a dated version and tag.
- Tighten the coverage gate: the flake `coverage` check now enforces **≥70% overall
  AND ≥80% for each core package** (`internal/{index,authz,auth,redact,defense,gitsafe,
  config,corpus}`). Raise `internal/gitsafe` coverage above 80% with targeted tests.
- Backfill `design.md` for the archived `optional-quarantine` change (records the
  security trade-off of disabling a defense).
- Run the `code-review` and `security-review` QA skills over this work and fix
  findings before shipping.

## Capabilities

### New Capabilities
- `dev-workflow`: the changelog, ship/release scripts, and the core-coverage gate that
  back the propose/ship workflows.

### Modified Capabilities
- `testing-gates`: coverage gate adds a per-core-package ≥80% floor.

## Impact

- New: `CHANGELOG.md`, `scripts/ship-change.sh`, `scripts/release.sh`,
  `internal/gitsafe` tests, `optional-quarantine/design.md`. Modified: `flake.nix`
  (coverage check). No production-code behaviour change.
