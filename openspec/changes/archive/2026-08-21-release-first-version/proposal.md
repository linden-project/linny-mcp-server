## Why

`scripts/release.sh` always bumps from `VERSION`, so the very first run — with
`VERSION=0.1.0` and no tags — would cut `v0.1.1`/`v0.2.0` and skip `v0.1.0`. The natural
first release is the current `VERSION` as-is. Separately, the CHANGELOG's `[Unreleased]`
section — written as a quick backfill — has a split `### Added` and is missing entries
for a few shipped changes (developer tooling: CHANGELOG + ship/release scripts +
core-coverage gate; the synthetic corpus generator). Before cutting `v0.1.0` the notes
should be a complete, accurate record.

## What Changes

- `scripts/release.sh`: when no tag exists yet for the current `VERSION`, offer to
  **release the current version without bumping** (a `current (X.Y.Z)` choice, or
  `release.sh current` non-interactively). If `vX.Y.Z` already exists, `current` is
  rejected. Every subsequent release bumps as before.
- Backfill `CHANGELOG.md` `[Unreleased]`: merge the split `### Added`, add the missing
  entries, and organize it as the complete `v0.1.0` record (so `release.sh` rolls a
  faithful section).

## Capabilities

### Modified Capabilities
- `release-automation`: the release script can cut the current untagged `VERSION`
  without a bump (first release), in addition to bumping.

## Impact

- Modified: `scripts/release.sh`, `CHANGELOG.md`. No production-code change; no gate
  impact. After this ships, `scripts/release.sh` → choose `current` → cuts `v0.1.0`.
