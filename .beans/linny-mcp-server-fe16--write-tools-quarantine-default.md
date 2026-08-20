---
# linny-mcp-server-fe16
title: Write tools (quarantine default)
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T20:42:16Z
parent: linny-mcp-server-s9mf
blocked_by:
    - linny-mcp-server-ef2b
    - linny-mcp-server-ixxf
---

create_doc (enforce slug convention), set_front_matter (mirror fred semantics; validate against lindenConfig), unset_front_matter, append_to_doc, archive. Validate -> write atomically -> reindex -> return resulting term membership.

**OpenSpec change:** `mcp-write-tools`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/mcp-write-tools/tasks.md`. Ships with tests._

## Summary of Changes

Added the MCP write tools (create_doc, append_to_doc, set_front_matter, unset_front_matter, archive) in internal/mcp/write.go, each running the safe-write pipeline: scope check (create needs write:inbox/write:*; edits need write:* or write:inbox for a quarantined draft; unreadable target == not-found); git-safety degraded gate; atomic + optimistic-concurrent write (create must-not-exist, edits by read-time hash); quarantine-by-default on create (status agent-draft); reindex then return resulting term membership; and an entry in the external append-only audit log for every attempt. Front-matter edits are surgical via yaml.Node (order/comment preserving). authz gained CanWriteAll/CanWriteInbox; index gained TermsOfDoc; serve opens the audit log in the state dir and enables writes unless read-only. docs/tools.md updated.

Verified: handler tests (create quarantined + membership + file + audit ok, forbidden without write scope, refused when degraded, set_front_matter updates membership + file, archive sets flag, modify forbidden with read-only) and a protocol-level e2e (create_doc over the SDK client then read back). Coverage 80.6 percent; nix flake check gotest+lint+coverage all passed. Known follow-ups (docs/future): full lindenConfig term validation and a complete fred-style editor.
