## Context

The corpus is a git working tree kept in sync by an external, out-of-scope git-sync
script running on every machine. The server must detect an unhealthy tree and stop
writing, but must never take ownership of git. Detection has to work while the user
is unreachable, so it favours false-positive safety (degrade when unsure) over
availability.

## Goals / Non-Goals

**Goals:** decide "is it safe to write right now?" from the live tree; make every
write atomic and concurrency-safe; recover automatically.

**Non-Goals:** no committing, merging, pulling, or conflict resolution; no ahead/behind
counts, `sync_status()` tool, or ntfy alerting (that is E0303).

## Decisions

- **File-based inspection first, git binary optional.** The decisive signals are
  read from the git dir directly (`MERGE_HEAD`, `rebase-merge/`, `rebase-apply/`,
  `CHERRY_PICK_HEAD`, `REVERT_HEAD`, `HEAD` for detached) plus a conflict-marker scan
  of tracked `*.md`. This keeps the guard dependency-free and testable. If a `git`
  binary is on PATH, `git ls-files -u` additionally catches unmerged index entries
  that have no committed markers. Absence of git degrades capability gracefully, not
  safety — markers and merge-state files still trip the guard.
- **Detached HEAD is treated as degraded.** A synced notebook is expected to sit on a
  branch; a detached HEAD usually means git-sync is mid-operation. The briefing lists
  it explicitly. Cheap and safe to block.
- **No sticky degraded flag.** `EnsureWritable()` re-inspects every time, so recovery
  is automatic and there is no stuck state to reset. The tiny inspection cost is
  acceptable at write frequency.
- **Conflict-marker scan matches the indexer.** Only lines beginning `<<<<<<<` or
  `>>>>>>>` count (never bare `=======`), so Markdown setext headings are not false
  positives — consistent with `internal/index`.
- **Atomic write = temp-in-same-dir → fsync → rename → fsync dir.** Same-directory
  temp guarantees `rename` is atomic on the same filesystem; the directory fsync makes
  the rename durable. Temp files use a `.tmp-*` pattern and are removed on error.
- **Optimistic concurrency via content hash, not mtime.** `HashFile` returns the empty
  string for a missing file, so "create only if absent" is expressed as
  `expectedHash == ""`.

## Risks / Trade-offs

- [Marker scan cost on large corpora] → scan is O(bytes) and only runs on write /
  health; acceptable at ~5k×4KB. Can be cached against mtimes later if needed.
- [git-binary detection variance] → treated as best-effort enrichment; the guard is
  safe without it.
- [Detached-HEAD blocking legitimate flows] → acceptable for a single-user notebook
  server; documented, and forced read-only / clean checkout clears it.

## Open Questions

- Should unmerged-index detection be mandatory (require git) rather than best-effort?
  Deferred until we know whether the target always ships git (it runs git-sync, so
  almost certainly yes).
