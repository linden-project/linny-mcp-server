## Why

The briefing (§10) makes thorough unit + e2e testing non-negotiable, and every change
so far has shipped tests gated by `nix flake check` (go test + lint). But "tests exist"
is not the same as "tests cover enough": there was no measured coverage and no gate to
stop coverage silently rotting as the surface grows. The user asked for a realistic
coverage floor (≥ 70%). Current project-wide coverage is ~82%, so a 70% gate protects
that headroom without being aspirational.

## What Changes

- Add `checks.<system>.coverage` to the flake: it runs the full suite with
  `-coverpkg=./...` (so cross-package e2e tests count toward coverage of the packages
  they exercise) and **fails the build if total statement coverage drops below 70%**.
  `git` is available in this check so the git-backed history/gitsafe tests actually run
  (rather than skipping).
- Document the testing agreement and the coverage gate in `openspec/project.md` so the
  contract is explicit: what "unit" and "e2e" mean here, and the 70% floor.
- Keep the coverage profile as a build output for inspection.

## Capabilities

### New Capabilities
- `testing-gates`: the coverage gate and the written testing agreement.

### Modified Capabilities
- `nix-packaging`: `checks` gains a coverage check.

## Impact

- Modified: `flake.nix` (new `coverage` check), `openspec/project.md` (testing
  agreement). No production-code change; adds CLI-level tests that raise coverage
  headroom (`cmd/lindexer`, `cmd/linny-mcp`).
- `nix flake check` now runs three checks: `gotest`, `lint`, `coverage`.
