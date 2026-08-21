## Context

Changes so far were shipped by hand (manual `git add`/`openspec archive`/`jj commit`/
`jj git push`), which works but drifts from the documented `/mip:ship` procedure and
skips its CHANGELOG step and its stricter coverage gate. The VCS is jj colocated with
git; the gate is `nix flake check`.

## Goals / Non-Goals

**Goals:** make `/opsx:propose` (design.md included) and `/mip:ship` runnable exactly
as written; one reproducible gated ship path; a meaningful per-core-package coverage
floor.

**Non-Goals:** no CI service wiring (the gate is `nix flake check`, run locally/by the
script); no change to runtime behaviour.

## Decisions

- **`ship-change.sh` uses jj**, matching the repo. It stages the whole tree (so the
  CHANGELOG edit and bean file ride along), runs `nix flake check` as the gate,
  archives the change with `openspec archive --yes`, then `jj commit` (author is Pim
  Snel via jj config — no trailers), `jj bookmark set main`, `jj git push`. On gate
  failure it stops non-zero and changes nothing downstream.
- **Core-coverage gate lives in the flake `coverage` check**, not the script, so
  `nix flake check` alone enforces it. It computes overall coverage (≥70%) and each
  core package's own coverage (≥80%) from one `-coverpkg=./...` profile plus per-core
  `go test -cover`. Core = the correctness/security packages that are already ≥80%
  today, with `gitsafe` raised to clear the bar; `cmd/*`, `mcp`, `alert`, `hugoref`,
  `backup`, `audit` are held only to the overall floor (they are glue, transport, or
  test-only reference code).
- **`release.sh`** rewrites `## [Unreleased]` to `## [<version>] - <date>` and adds a
  fresh empty `[Unreleased]`; tagging is left explicit. Date is passed in / read from
  the system at run time (scripts may use the clock; workflow *scripts* here are shell,
  not the JS workflow runtime).

## Risks / Trade-offs

- [Core list drifts] → the list is defined in one place (the flake check) and
  documented here; adding a package to "core" is a deliberate edit.
- [gitsafe at exactly ~80%] → add margin (target mid-80s) so incidental churn doesn't
  trip the gate.
- [Disabling quarantine weakens defense] → recorded in the backfilled
  optional-quarantine design; default stays on and startup warns.

## Open Questions

- Whether to later split "core ≥80%" per-package thresholds (e.g. redact/authz ≥90%).
  Deferred; a flat 80% core floor is enough for the alpha.
