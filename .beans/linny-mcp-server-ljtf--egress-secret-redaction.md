---
# linny-mcp-server-ljtf
title: Egress secret redaction
status: todo
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:00:53Z
parent: linny-mcp-server-fo8d
blocked_by:
    - linny-mcp-server-gvc5
    - linny-mcp-server-sac5
---

gitleaks-style redaction filter on ALL tool responses so no response can return a credential regardless of what an agent asks for. High priority. Exercised by fake secrets in the synthetic corpus.

**OpenSpec change:** `egress-redaction`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/egress-redaction/tasks.md`. Ships with tests._
