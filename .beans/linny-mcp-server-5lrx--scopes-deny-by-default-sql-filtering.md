---
# linny-mcp-server-5lrx
title: 'Scopes: deny-by-default + SQL filtering'
status: todo
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:00:53Z
parent: linny-mcp-server-fo8d
blocked_by:
    - linny-mcp-server-zin6
    - linny-mcp-server-gvc5
---

Scope vocabulary read:taxonomy:<name>, deny:taxonomy:<name>, write:inbox, write:*, delete:*. Deny by default. Intersection semantics: a doc tagged work+health is excluded when health denied (evaluate deny across ALL terms). Filter in SQL at query time, never post-filter; never leak existence. One token per client.

**OpenSpec change:** `authz-scopes`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/authz-scopes/tasks.md`. Ships with tests._
