## 1. CHANGELOG

- [x] 1.1 Add CHANGELOG.md (Keep a Changelog) with [Unreleased], backfilled

## 2. Scripts

- [x] 2.1 scripts/ship-change.sh: stage → gate → archive → commit (jj, Pim) → push
- [x] 2.2 scripts/release.sh: promote [Unreleased] → dated version

## 3. Gate

- [x] 3.1 flake coverage check: overall ≥70% AND each core package ≥80%
- [x] 3.2 Raise internal/gitsafe coverage above 80%

## 4. Backfill + QA

- [x] 4.1 Backfill optional-quarantine/design.md (security trade-off)
- [x] 4.2 Run code-review skill over the diff; fix findings
- [x] 4.3 Run security-review skill over the diff; fix findings

## 5. Ship

- [x] 5.1 Dogfood scripts/ship-change.sh to ship this change; nix flake check green
