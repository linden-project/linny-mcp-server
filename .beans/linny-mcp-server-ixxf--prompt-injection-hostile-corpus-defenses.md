---
# linny-mcp-server-ixxf
title: Prompt-injection & hostile-corpus defenses
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T20:31:23Z
parent: linny-mcp-server-fo8d
blocked_by:
    - linny-mcp-server-gvc5
    - linny-mcp-server-ef2b
---

Agent writes land in quarantine taxonomy (inbox/agent-draft) by default; promotion is separate. No tool widens its own scope. Delete/bulk-retag require out-of-band confirmation. Every write logged with diff to an append-only audit log OUTSIDE the corpus. Wrap returned bodies in explicit data delimiters.

**OpenSpec change:** `hostile-corpus-defenses`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/hostile-corpus-defenses/tasks.md`. Ships with tests._

## Summary of Changes

Landed the hostile-corpus guardrails ahead of the write tools. internal/audit: an append-only audit log kept OUTSIDE the corpus (in stateDir), one JSON line per write (time/identity/tool/slug/diff/outcome), opened O_APPEND and never rewritten; operator-facing so NOT redacted. internal/defense: a quarantine Policy (default taxonomy status, term agent-draft) with ApplyQuarantine (scalar/list merge, idempotent) + IsQuarantined, and RequiresConfirmation for destructive ops (delete, bulk_retag); plus Delimit which wraps returned bodies in explicit data-delimiter markers and strips any forged marker first so framing cannot be broken. get_doc now returns its (redacted) body wrapped in those delimiters. The scope invariant (no tool widens its own scope) is documented: a request scope is fixed at construction from the token.

Verified: audit append-only across reopen + path-outside-corpus; quarantine apply/detect for missing/scalar/list and confirmation flags; delimit wrap + forged-marker neutralization; and get_doc returns a delimited body. Coverage 81.9
