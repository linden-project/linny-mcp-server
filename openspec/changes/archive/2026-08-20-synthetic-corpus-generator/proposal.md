## Why

Every test in this project must run against a corpus that looks like a real Linny
notebook but contains **no private data** — the briefing (§10) forbids ever pointing
the build at the real `secondbrain` repo. A deterministic generator gives the indexer,
auth, scope, redaction, and git-safety epics a shared, reproducible fixture that
exercises the edge cases each of them depends on.

## What Changes

- Add `internal/corpus`: a deterministic generator that writes a flat directory of
  `.md` records with realistic YAML front matter (multiple taxonomies, `crdate`,
  `starred`, task lists) plus a matching `lindenConfig/` (L1/L2 config files) and a
  Hugo `config.yaml` so the corpus is also buildable by the reference indexer.
- Cover the deliberate edge cases: unicode, very long front matter, empty bodies,
  malformed YAML, committed git conflict markers, and embedded **fake** secrets
  (to exercise the future redaction filter).
- Deterministic output: same seed → byte-identical corpus (no wall-clock, no RNG
  without a seed), so tests are stable and `verify` diffs are meaningful.
- Add `cmd/gen-corpus`: a thin CLI to materialize a corpus into `testdata-gen/` for
  manual inspection and for the Hugo `verify` path.

## Capabilities

### New Capabilities
- `synthetic-corpus`: a deterministic, edge-case-rich synthetic Linny notebook
  generator used by all downstream tests and never derived from real data.

### Modified Capabilities

## Impact

- New: `internal/corpus/**`, `cmd/gen-corpus/**`, `testdata-gen/` (git-ignored).
- Depended on by milestones 02 (indexer round-trip), 04 (redaction/scopes), 03
  (git-safety conflict simulation).
