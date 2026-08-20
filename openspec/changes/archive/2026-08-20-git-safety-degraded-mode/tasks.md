## 1. Tree inspection

- [x] 1.1 Resolve the git dir (handle `.git` dir and `.git` gitdir-file)
- [x] 1.2 Detect in-progress op: MERGE_HEAD, rebase-merge/, rebase-apply/, CHERRY_PICK_HEAD, REVERT_HEAD
- [x] 1.3 Detect detached HEAD (HEAD not a symbolic ref)
- [x] 1.4 Scan tracked *.md for committed conflict markers
- [x] 1.5 Detect unmerged index entries via `git ls-files -u` when the git binary is present (optional)
- [x] 1.6 `TreeState` with Clean/Conflicted/ConflictedPaths/InProgress/Detached/Reason

## 2. Guard & errors

- [x] 2.1 `Guard(root, forceReadOnly)`; `State()`; `EnsureWritable()`
- [x] 2.2 Typed retryable errors: `DegradedError`, `StaleWriteError`

## 3. Safe writes

- [x] 3.1 `AtomicWrite` (temp in same dir → fsync → rename → fsync dir)
- [x] 3.2 `HashFile` (sha256; empty hash for a nonexistent file)
- [x] 3.3 `WriteIfUnchanged(path, data, expectedHash)`; gate via `EnsureWritable`

## 4. Server wiring

- [x] 4.1 `serve` builds a Guard from --corpus/--read-only
- [x] 4.2 `/healthz` reports degraded/conflicted/paths from the guard

## 5. Tests

- [x] 5.1 Clean tree writable; merge-in-progress blocks; conflict markers block + list paths
- [x] 5.2 Detached HEAD blocks; forced read-only blocks
- [x] 5.3 Automatic recovery once the tree is clean
- [x] 5.4 Atomic write leaves no temp; content is all-or-nothing
- [x] 5.5 Stale-hash write rejected (retryable); fresh write accepted
- [x] 5.6 /healthz reflects conflicted state
- [x] 5.7 nix flake check green
