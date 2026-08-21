## 1. release.sh first-release path

- [x] 1.1 Offer `current (X.Y.Z)` when `vX.Y.Z` is untagged; support `release.sh current`
- [x] 1.2 Reject `current` when the tag already exists
- [x] 1.3 `current` → no bump (next = VERSION); bump levels unchanged

## 2. CHANGELOG backfill

- [x] 2.1 Merge the split `### Added`; add missing entries (dev tooling, corpus generator)
- [x] 2.2 Organize `[Unreleased]` as the complete, accurate v0.1.0 record

## 3. Verify & ship

- [x] 3.1 shellcheck + bash -n; dry-check the `current`/bump selection logic
- [x] 3.2 nix flake check green; ship via scripts/ship-change.sh
