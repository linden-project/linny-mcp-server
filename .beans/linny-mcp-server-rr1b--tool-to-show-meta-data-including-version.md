---
# linny-mcp-server-rr1b
title: tool to show meta data including version
status: completed
type: feature
priority: normal
created_at: 2026-09-15T20:22:22Z
updated_at: 2026-09-15T20:57:32Z
---


**OpenSpec change:** `add-session-info-tool`

## Summary of Changes

Added the `session_info` MCP tool: server name and version, notebook name, quarantine
and writes-enabled flags, the caller's identity and raw scopes, derived permissions,
live tree status, and a single `can_write_now` verdict with the reason it is false.

The motivating gap was narrower than the title suggests. `server.go` reported a
caller's scopes in exactly one place: the placeholder handler used when no store is
attached, whose own response says read tools are unavailable. In a real deployment the
scopes compiled into a ScopeSet that nothing read back, so an agent could not tell a
missing scope from a degraded tree from a read-only server. All three surface as "I
cannot write", which is how a `write:inbox` token cost an afternoon of deployment
archaeology.

`can_modify` is three-valued (`all`, `own-drafts`, `none`) because `ensureModify` has
three outcomes and a boolean has to lie about the middle one, which is the common
case. It is computed from `CanWriteAll`/`CanWriteInbox`, the predicates the write path
enforces with, so it cannot drift. `can_write_now` is the conjunction of registration,
guard and scope, naming the first failing gate; creating counts as writing, so a
`write:inbox` token gets true with a reason saying existing documents need `write:*`.
Two drift risks were closed rather than documented: tree state reads from the accessor
`sync_status` uses, and the writes-enabled condition was extracted from an inline
expression in `server.go` so registration and the report share one predicate.
`internal/authz` gained `CanReadAny` to mirror the write predicates.

The response describes the connection and the credential only: no counts, taxonomy or
term names, slugs, or filesystem paths, asserted by serialising it and searching for
each. No `can_delete`, since `delete:*` parses but no tool implements it.

Verified: 9 tests, most driven through the real HTTP handler so the wiring in
server.go is under test rather than a copy. The gate caught `internal/authz` dropping
to 77.9% against the 80% core floor when `CanReadAny` shipped untested; tests brought
it to 83.1%. Coverage 83.2% total, internal/mcp 85.5%. nix flake check green.
