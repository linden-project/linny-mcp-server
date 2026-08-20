## Why

`indexer-verify` shipped the diff engine and a `verify --reference <dir>` that consumes
a pre-built reference index. The briefing's full intent (§4/§10) is to actually **run
Hugo's indexer** and diff ours against it — the round-trip that keeps drift visible.
This change makes `verify` build the Hugo reference itself.

## What Changes

- Vendor the reference Hugo site (`layouts/` + `config/_default/config.yaml`) from
  `linny-notebook-template` under `internal/hugoref/site`, embedded with `go:embed`.
- Add `internal/hugoref.BuildReference(corpusRoot)`: assemble a temp Hugo site from the
  embedded layouts/config plus the corpus's `content/` and `lindenConfig/` (copied, so
  the real corpus is never mutated with layouts), run `hugo`, and return the produced
  index directory. Requires the `hugo` binary (in the dev shell); returns a clear
  "hugo not found" error otherwise.
- Extend the differ: `VerifyDirsWithOpts(ours, ref, {IgnoreReferenceOnly})` compares
  only the files our indexer emits — skipping Hugo's vestigial per-page
  `<slug>/index.json` outputs (which are also invalid JSON) — and normalizes away
  Hugo's injected built-in params (`draft`, `iscjklanguage`) in `docs_with_props`.
- Add `lindexer verify --hugo`: build the reference via Hugo instead of `--reference`.

## Capabilities

### New Capabilities
- `verify-hugo`: run the Hugo reference indexer and diff ours against it.

### Modified Capabilities
- `index-verify`: subset comparison ignoring reference-only files + built-in-param
  normalization.

## Impact

- New: `internal/hugoref/**` (embedded reference site). Modified:
  `internal/index/verify.go`, `cmd/lindexer` verify. Test skips when `hugo` is absent.
- Documents the known synthetic-corpus divergences (Hugo emits `{}` L1 term configs
  and adds built-in params) as reported drift — the tool's whole point.
