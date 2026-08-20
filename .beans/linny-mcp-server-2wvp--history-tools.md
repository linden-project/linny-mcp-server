---
# linny-mcp-server-2wvp
title: History tools
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T20:09:12Z
parent: linny-mcp-server-s9mf
blocked_by:
    - linny-mcp-server-gvc5
---

history(doc), diff(doc, ref), changed_since(date). Recovers git's main advantage over a bare index.

**OpenSpec change:** `mcp-history-tools`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/mcp-history-tools/tasks.md`. Ships with tests._

## Summary of Changes

Added git-history MCP tools. internal/gitsafe/history.go: read-only History (git log), Diff (git diff <ref> -- path), ChangedSince (git log --since --name-only); ref/since inputs that begin with "-" are rejected so they can never be read as git flags (args run without a shell). internal/mcp registers history(slug, limit?), diff(slug, ref), changed_since(since); the Server carries CorpusPath and the reader uses it. Each tool resolves the slug via the scoped store (denied==not-found), and free-text output (commit subjects, diff hunks) is passed through the egress redactor — important because a diff can reintroduce a secret a current-state read would have stripped. changed_since returns only readable changed docs. docs/tools.md updated (history tools shipped in v1).

Verified: gitsafe unit tests on a real temp git repo (2-commit history order, diff shows added line, changed_since lists the path, flag-like ref/since rejected) and MCP tool tests on a git-backed synthetic corpus (health-denied doc history==not-found, diff of a working-tree AWS key redacted to placeholder, changed_since excludes the denied doc). nix flake check: all checks passed.
