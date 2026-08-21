---
# linny-mcp-server-xdrx
title: 'Dev workflow tooling: CHANGELOG, ship/release scripts, core-coverage gate, QA'
status: completed
type: epic
priority: normal
created_at: 2026-08-21T11:26:22Z
updated_at: 2026-08-21T11:35:54Z
parent: linny-mcp-server-xwyf
---

Make /opsx:propose and /mip:ship first-class: add CHANGELOG.md (Unreleased), scripts/ship-change.sh + scripts/release.sh, tighten the gate to >=70% overall AND >=80% core, raise gitsafe coverage, backfill a design.md, and run code-review + security-review. **OpenSpec change:** `dev-workflow-tooling`

## Summary of Changes

Made /opsx:propose and /mip:ship first-class. Added CHANGELOG.md (Keep a Changelog, [Unreleased] backfilled); scripts/ship-change.sh (stage -> nix flake check gate -> openspec archive -> jj commit as Pim -> push; never bypasses the gate) and scripts/release.sh (promote [Unreleased] -> dated version). Tightened the flake coverage check to enforce >=70% overall AND >=80% for each core package (index/authz/auth/redact/defense/gitsafe/config/corpus); raised internal/gitsafe to ~83% with targeted tests (gitdir-file resolution, cherry-pick/revert/rebase detection, empty-ref guard, empty history). Backfilled design.md for the archived optional-quarantine change. Ran code-review (clean) and security-review (no findings; tooling/tests/docs only, no runtime attack surface). Dogfooded ship-change.sh to ship this change. nix flake check all green.
