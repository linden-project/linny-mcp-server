## Why

The briefing (§12) is unambiguous: "A tested backup/restore path. If the server can
delete, a verified restore is mandatory, not optional." The write tools mutate the
corpus (create, set/unset front matter, archive), so a verified snapshot/restore is
required data-safety — independent of git-sync, and testable.

## What Changes

- Add `internal/backup`: `Backup(corpusRoot, w)` writes a `tar.gz` snapshot of the
  source-of-truth data (the content dir and `lindenConfig`), and `Restore(r,
  corpusRoot)` extracts it back, with path sanitization so no entry can escape the
  target directory (no `..`/absolute-path traversal). The disposable index/state and
  VCS dirs are excluded — they are rebuildable.
- Add `linny-mcp backup --corpus <c> --out <file.tar.gz>` and
  `linny-mcp restore --in <file.tar.gz> --corpus <c>`.
- The verified round-trip is the deliverable: a test backs up, deletes/mutates a
  record, restores, and asserts the original content is recovered byte-for-byte.

## Capabilities

### New Capabilities
- `backup-restore`: a verified corpus snapshot/restore.

### Modified Capabilities

## Impact

- New: `internal/backup/**`; `linny-mcp backup`/`restore` wired. Standard library only
  (archive/tar + compress/gzip).
