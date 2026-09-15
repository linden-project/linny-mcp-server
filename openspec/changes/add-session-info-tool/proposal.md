## Why

An agent connected to linny-mcp cannot ask what it is allowed to do.

`internal/mcp/server.go:73` returns the caller's scopes, and it is the only place that
does. It is the placeholder handler used when no store is attached, whose own response
says `no notebook store attached; read tools unavailable`. So the single path that
reports a caller's permissions is the path where the server does nothing useful. In a
real deployment the store is attached, `authz.Parse(id.Scopes)` compiles the scopes
into a `ScopeSet`, and no tool ever reads them back.

The cost of that is not theoretical. A caller holding `read:*` and `write:inbox` can
create quarantined drafts but cannot touch an existing document. What it sees is a
refusal, so it reports that it cannot write at all. Working out why means inspecting
the deployment: the token file, the server flags, the git working tree. The agent
cannot self-diagnose, and neither can the person asking it.

Worse, the symptom is shared by four independent causes:

```
write tools not registered   (no state dir / no audit log / forced read-only)
        OR  guard degraded   (conflict markers, mid-rebase, detached HEAD)
        OR  scopes too narrow
        OR  target not readable under scope
                    │
                    └──▶ all four surface as "I cannot write"
```

Nothing exposes which one fired until a write is attempted and refused, which also
writes a `denied` entry to the audit log for what was really a capability question.

The version half of this is smaller but real. `buildinfo.Version` reaches the MCP
handshake as `Implementation.Version`, where the client consumes it at connect. It
never becomes a tool result, so an agent cannot answer "which version is this server
running" without someone reading the service logs.

## What Changes

- Add an MCP tool **`session_info`**, taking no arguments, returning what the current
  connection is: the server, the notebook, the caller, and whether writing is possible
  right now.

```json
{
  "server":   { "name": "linny-mcp", "version": "0.1.0" },
  "notebook": { "name": "default", "quarantine": true, "writes_enabled": true },
  "caller":   { "identity": "claude-online",
                "scopes": ["read:*", "write:inbox"],
                "can_read": true, "can_create": true, "can_modify": "own-drafts" },
  "status":   { "degraded": false, "read_only": false, "reason": "" },
  "can_write_now": false,
  "reason": "token has write:inbox, not write:*; existing documents cannot be modified"
}
```

- `can_modify` is a three-state value (`all`, `own-drafts`, `none`), because
  `ensureModify` has three states and a boolean has to lie about the middle one.
- Derived permissions are computed from `ScopeSet.CanWriteAll` and
  `ScopeSet.CanWriteInbox`, the same predicates the write path enforces with, so the
  report cannot drift from the enforcement.
- `can_write_now` is the conjunction of all three gates, with `reason` naming the one
  that failed. Reporting the gates separately and leaving the agent to combine them is
  how the wrong diagnosis gets made.
- Degraded state is read through the same guard-backed accessor `sync_status` already
  uses, so the two tools cannot disagree.
- Extract the writes-enabled condition (`Guard != nil && Audit != nil &&
  !Guard.ForcedReadOnly()`) so tool registration and this report share one expression.

## Capabilities

### Modified Capabilities
- `mcp-operational-tools`: adds `session_info`.

## Impact

- Modified: `internal/mcp/server.go` (thread the caller's identity, raw scopes and the
  policy into the read-side; extract the writes-enabled predicate),
  `internal/mcp/tools.go` (result types, handler, registration), `docs/tools.md`.
- No new dependencies, no schema change, no authorization vocabulary change, and no
  change to any existing tool's behaviour or output.
- `internal/authz` gains nothing: the tool consumes predicates that already exist.

## Non-Goals

- **Filesystem paths.** `_indexer_info.json` records the content and config
  directories; this tool does not. An agent has no use for host layout, and nothing in
  the egress redactor is looking for paths.
- **Document counts and index freshness.** `verify_index` already reports them.
- **A tool-surface version.** `docs/tools.md` keeps a changelog of the tool surface
  (v1, v1.1, v1.2), but it exists only in prose. Reporting it would mean maintaining a
  constant in code that nothing enforces.
- **`can_delete`.** `authz` parses `delete:*` but no tool consumes it. Reporting it
  would advertise a capability the server does not have.
- **The no-store placeholder path.** It already reports identity and scopes; it is not
  worth aligning a degraded-mode stub with the real tool.
