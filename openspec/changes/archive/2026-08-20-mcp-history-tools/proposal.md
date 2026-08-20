## Why

Git ships the corpus's full history, but the read tools only expose the current
state. The briefing (§8) lists history tools as the feature that "recovers git's main
advantage" — letting an agent see when and how a note changed. These are read
operations, so they must obey the same rules as the other read tools: scope-filtered
(you cannot see the history of a document you cannot read) and redacted (a diff can
contain a credential that a current-state read would have stripped).

## What Changes

- Add `internal/gitsafe` history helpers that inspect the corpus git tree read-only:
  `History` (commit list for a path), `Diff` (a ref-vs-working diff for a path), and
  `ChangedSince` (paths changed since a date). Ref/since inputs are validated to not
  begin with `-` so they can never be interpreted as git flags (args are passed
  without a shell, so there is no shell-injection surface).
- Add MCP tools `history(slug, limit?)`, `diff(slug, ref)`, `changed_since(since)`.
  Each resolves the slug to a document, enforces the caller's scope (denied ⇒
  not-found), and pipes free-text output (commit subjects, diff hunks) through the
  egress redactor.
- `changed_since` returns only the changed documents the caller may read.
- Attach the notebook's corpus path to the server so tools can run git.
- Record the tools in `docs/tools.md`.

## Capabilities

### New Capabilities
- `mcp-history-tools`: git-history navigation over MCP, scope-aware and redacted.

### Modified Capabilities
- `git-safety`: adds read-only history/diff/changed-since helpers.
- `mcp-read-tools`: registers the three history tools alongside the read tools.

## Impact

- New: `internal/gitsafe/history.go`. Modified: `internal/mcp` (tools + Server
  gains `CorpusPath`), `cmd/linny-mcp` serve (passes corpus path), `docs/tools.md`.
- No new dependency (git binary invoked via os/exec, as already done for unmerged
  detection). Tools return an error if the git binary is unavailable.
