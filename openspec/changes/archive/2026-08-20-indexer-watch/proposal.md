## Why

The briefing (§4) lists "Incremental updates via `fsnotify`" so the index stays fresh
as the corpus changes (git-sync pulls, edits from `linny.vim`) without a manual
rebuild. This adds a `watch` mode.

## What Changes

- Add `internal/index.Watcher`: watches the content dir and `lindenConfig` via
  `fsnotify`, coalesces bursts of events with a **debounce**, and invokes a rebuild
  callback once per quiet period. (A debounced *full* rebuild — cheap at PoC scale;
  true per-file incremental diffing is noted in `docs/future.md`.)
- Add `lindexer watch --corpus <c> --state-dir <s> [--index <dir>]`: build once, then
  watch; on each change batch, rebuild the graph and repopulate the SQLite store (and
  re-emit JSON when `--index` is given), logging each refresh.
- The debounce is a small standalone helper, tested deterministically; the fsnotify
  wiring is thin and covered by an integration test.

## Capabilities

### New Capabilities
- `index-watch`: debounced fsnotify-driven index refresh and the `watch` CLI.

### Modified Capabilities

## Impact

- New dependency `github.com/fsnotify/fsnotify`; `vendorHash` updated. New:
  `internal/index/watch.go`; `lindexer watch` wired.
