## 1. git history helpers

- [x] 1.1 `History(root, relpath, limit)` via `git log --format=…`
- [x] 1.2 `Diff(root, relpath, ref)` via `git diff <ref> -- <relpath>`
- [x] 1.3 `ChangedSince(root, since)` via `git log --since --name-only`
- [x] 1.4 Reject ref/since beginning with `-`; git-absent returns an error

## 2. MCP tools

- [x] 2.1 Server gains CorpusPath; reader carries it
- [x] 2.2 `history(slug, limit?)` — scope check (denied==not-found), redact subjects
- [x] 2.3 `diff(slug, ref)` — scope check, redact diff output
- [x] 2.4 `changed_since(since)` — scope-filter changed docs to readable ones
- [x] 2.5 Register tools; update docs/tools.md

## 3. Tests & gate

- [x] 3.1 gitsafe history/diff/changed_since against a real temp git repo
- [x] 3.2 ref/since flag-rejection
- [x] 3.3 MCP: denied doc history == not-found; changed_since scope-filtered
- [x] 3.4 diff redaction
- [x] 3.5 nix flake check green
