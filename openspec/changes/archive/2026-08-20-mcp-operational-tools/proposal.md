## Why

The operational tool surface (briefing §8) is `sync_status()` and `verify_index()`.
`sync_status` shipped with the alerting epic. This change adds `verify_index` as the
operational health check it is most useful as over MCP: **is the served SQLite index
consistent with the corpus on disk right now?** (Has the corpus changed under a server
that isn't watching, leaving the index stale?)

The briefing's other sense of verify — diff our JSON against Hugo's output — already
ships as the `lindexer verify` CLI (it needs the Hugo layouts + binary, which belong
in a build/verify context, not a live request). This tool is the runtime counterpart.

## What Changes

- Add `index.Store.AllDocFilenames()` — the doc set currently in the served store.
- Add a `verify_index` MCP tool: rebuild the taxonomy graph from the corpus, compare
  its document set against the served store, and report `corpus_docs`, `store_docs`,
  `missing_from_store` (in corpus, not yet indexed), `stale_in_store` (indexed, gone
  from corpus), `conflicted` paths, and an `in_sync` flag. No Hugo needed.
- Register it as an operational tool; document it in `docs/tools.md`.

## Capabilities

### New Capabilities
- `mcp-operational-tools`: the `verify_index` served-vs-corpus consistency tool.

### Modified Capabilities
- `index-store`: adds `AllDocFilenames`.

## Impact

- Modified: `internal/index` (AllDocFilenames), `internal/mcp` (verify_index tool),
  `docs/tools.md`. No new dependency.
