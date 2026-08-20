## Why

The single most valuable safety behaviour in the briefing (§6): while the user is
away from his phone, the server must **not** keep writing into a broken git working
tree. An external git-sync runs on every machine and owns git; the server must never
reimplement or fight it. Instead it inspects the tree itself and, whenever the tree
is not clean-and-merged, drops into **degraded read-only mode** — refusing writes with
a clear, retryable error — and returns to normal automatically once the tree is clean
again. This is a DoD item: "a conflicted tree puts the server read-only, and
`sync_status()` says so."

## What Changes

- Add `internal/gitsafe`: inspect a notebook's git working tree for the conditions
  that must block writes — committed conflict markers in tracked files, an in-progress
  merge/rebase/cherry-pick/revert, unmerged index entries, and detached HEAD.
- `Guard`: given a corpus root (and an optional forced read-only flag), report tree
  state and gate writes via `EnsureWritable()`. Degraded mode is entered/left purely
  as a function of the live tree — no sticky state to get stuck in.
- **Atomic writes**: temp file in the same directory → `fsync` → `rename` → fsync the
  directory, so Hugo and `linny.vim` never observe a half-written note.
- **Optimistic concurrency**: `HashFile` records a content hash at read time;
  `WriteIfUnchanged` fails with a retryable stale-write error if the file changed
  underneath.
- Errors are typed and marked **retryable** (degraded, stale-write) so callers/agents
  surface a "try again" rather than a crash.
- Wire the running server: `linny-mcp serve` builds a `Guard` from `--corpus` and
  `--read-only`, and `/healthz` now reports real `degraded`, `conflicted`, and
  conflicted-paths fields derived from the guard.

The `sync_status()` MCP tool, ahead/behind counts, and ntfy alerting are the next
epic (E0303); this change delivers the detection + gate + safe-write primitives and
the health surfacing.

## Capabilities

### New Capabilities
- `git-safety`: working-tree inspection, degraded read-only gating, atomic writes,
  and optimistic-concurrency writes for the notebook corpus.

### Modified Capabilities
- `server-runtime`: `/healthz` reports live degraded/conflicted state from the guard.

## Impact

- New: `internal/gitsafe/**` (replaces the doc.go stub).
- Modified: `cmd/linny-mcp` `serve` (build guard, wire health), `internal/mcp` health
  derivation. Standard library only.
