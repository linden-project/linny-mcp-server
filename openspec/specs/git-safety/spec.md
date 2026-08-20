# git-safety Specification

## Purpose
TBD - created by archiving change git-safety-degraded-mode. Update Purpose after archive.
## Requirements
### Requirement: Working-tree inspection

The server SHALL inspect a notebook's git working tree for the conditions that make
it unsafe to write: committed conflict markers in tracked files, an in-progress
merge/rebase/cherry-pick/revert, unmerged index entries, and a detached HEAD. It
SHALL NOT modify git state or invoke the external git-sync.

#### Scenario: Clean tree is writable

- **WHEN** the working tree has no conflict markers, no in-progress operation, and a
  normal branch checkout
- **THEN** inspection reports the tree as clean and writable

#### Scenario: Merge in progress detected

- **WHEN** a `MERGE_HEAD` (or rebase/cherry-pick/revert state) exists in the git dir
- **THEN** inspection reports an in-progress operation and the tree as not writable

#### Scenario: Committed conflict markers detected

- **WHEN** a tracked file contains a line beginning with `<<<<<<<` or `>>>>>>>`
- **THEN** inspection reports the tree conflicted and lists the offending path(s)

### Requirement: Degraded read-only mode gates writes

The server SHALL refuse every write while the tree is not clean-and-merged, returning
a clear, **retryable** error. It SHALL enter and leave degraded mode purely as a
function of the live tree state (no sticky flag), so it recovers automatically once
the tree is clean again.

#### Scenario: Write refused while degraded

- **WHEN** a write is attempted against a conflicted or in-progress tree
- **THEN** it fails with a retryable degraded-mode error and no file is modified

#### Scenario: Automatic recovery

- **WHEN** the tree becomes clean again after having been conflicted
- **THEN** a subsequent write is permitted without any manual reset

#### Scenario: Forced read-only

- **WHEN** the server is started with read-only forced on
- **THEN** all writes are refused regardless of tree state

### Requirement: Atomic writes

Every write to the corpus SHALL be atomic: write to a temporary file in the same
directory, `fsync` it, then `rename` it into place (and fsync the directory).
Consumers (Hugo, `linny.vim`) SHALL never observe a partially written note.

#### Scenario: No partial file on completion

- **WHEN** a document is written
- **THEN** the destination path only ever contains complete content, and no temp file
  is left behind on success

### Requirement: Optimistic concurrency

Writes SHALL support optimistic concurrency: a caller records a content hash at read
time, and a write SHALL fail with a retryable stale-write error if the file changed
underneath (its current hash differs from the expected hash).

#### Scenario: Stale write rejected

- **WHEN** a write supplies an expected hash that no longer matches the file on disk
- **THEN** the write fails with a retryable stale-write error and the file is unchanged

#### Scenario: Fresh write accepted

- **WHEN** a write supplies the current hash (or the empty hash for a new file)
- **THEN** the write succeeds atomically

### Requirement: Health reflects degraded state

`/healthz` SHALL report the live guard state: a `degraded` flag, a `conflicted`
flag, and the conflicted paths when any.

#### Scenario: Healthz shows conflict

- **WHEN** the tree is conflicted and `/healthz` is requested
- **THEN** the response has `degraded: true`, `conflicted: true`, and lists the paths

