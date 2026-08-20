## Why

The briefing (§1.3) says the real operations on a Linny notebook are front-matter
state transitions, not file edits — and §8 defines the write tool surface. With the
git-safety atomic-write path, the quarantine policy, the audit log, and the store all
in place, the write tools can finally be assembled safely: quarantine-by-default,
degraded-mode-aware, optimistic-concurrent, validated, reindexed, and audited.

## What Changes

- Add MCP write tools: `create_doc`, `append_to_doc`, `set_front_matter`,
  `unset_front_matter`, `archive`. Each runs the full safe-write pipeline:
  1. **Scope check** — `create_doc` needs `write:inbox` (or `write:*`); modifying an
     existing doc needs `write:*`, or `write:inbox` when the doc is a quarantined
     draft. A doc the caller cannot read is reported as not-found.
  2. **Degraded gate** — `guard.EnsureWritable()`; refused with a retryable error when
     the working tree is conflicted / mid-operation.
  3. **Atomic + optimistic write** — `create_doc` requires the slug not to exist;
     edits use the read-time content hash (`WriteIfUnchanged`), so a concurrent change
     fails as a retryable stale-write.
  4. **Quarantine by default** — `create_doc` places the new document in the
     quarantine term (`status: agent-draft`).
  5. **Reindex, then return resulting term membership** so the agent sees what its
     write actually did.
  6. **Audit** — every attempt (ok/denied/error) is appended to the external
     append-only audit log with the resulting content as the diff.
- Front-matter edits (`set`/`unset`/`archive`) are **surgical**: the YAML mapping is
  edited as a `yaml.Node`, preserving key order and comments (avoiding the
  key-reordering corruption the briefing warns about). `create_doc` renders fresh.
- Register the write tools only when writes are enabled (a guard + audit log present
  and the guard is not forced read-only). `serve` opens the audit log in the state
  dir and enables writes unless `--read-only`.

## Capabilities

### New Capabilities
- `mcp-write-tools`: quarantine-default, degraded-aware, audited write tools that
  return resulting term membership.

### Modified Capabilities
- `server-runtime`: `serve` opens the audit log and enables write tools.

## Impact

- New: `internal/mcp/write.go`. Modified: `internal/mcp/server.go` (Guard/Audit/Policy
  + write registration), `internal/authz` (write-scope predicates),
  `internal/index` (`TermsOfDoc`), `cmd/linny-mcp` serve (audit log). `docs/tools.md`
  updated. Uses `gopkg.in/yaml.v3` (already a dependency).
- Known PoC limitation: full `lindenConfig` term validation and a fully fred-style
  surgical editor for all value types are follow-ups (see `docs/future.md`).
