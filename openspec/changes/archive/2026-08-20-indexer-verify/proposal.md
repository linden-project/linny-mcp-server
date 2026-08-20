## Why

The whole "own index" decision rests on a safety net: our JSON must stay compatible
with what Hugo emits, and drift must be **visible, not silent** (briefing §4, "Hugo's
current JSON output is the reference implementation… ship a `verify_index` tool that
diffs our index against Hugo's output"). The emit half shipped with `indexer-core`;
this change adds the diff.

## What Changes

- Add `index.VerifyDirs(ours, reference)` — walk both index trees and report
  discrepancies: files present on only one side, and per-file content differences.
  Comparison rules match the spec: **arrays are compared as sets** (order-insensitive,
  §3 of the spec), and `_indexer_info.json` ignores the identity fields
  (`product_name`/`product_version`/`hugo_version`) and any value Hugo leaves as the
  literal `"TODO"`.
- Add `lindexer verify --corpus <c> --reference <hugo-index-dir>`: build our index
  from the corpus into a temp dir and diff it against a reference index tree; print
  each discrepancy and exit non-zero if any.
- Producing the Hugo reference requires the notebook's Hugo layouts and the `hugo`
  binary (in the dev shell). `verify` consumes a reference tree rather than running
  Hugo itself, so it is testable and usable regardless of how the reference was built;
  the Hugo-driven path is documented in `docs/future.md`.

## Capabilities

### New Capabilities
- `index-verify`: an order-insensitive index-tree diff and the `verify` CLI.

### Modified Capabilities

## Impact

- New: `internal/index/verify.go`; `lindexer verify` wired. No new dependency.
- Delivers the drift safety net; tested on the synthetic corpus.
