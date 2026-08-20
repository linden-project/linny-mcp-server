---
# linny-mcp-server-r7gl
title: Front-matter parser & taxonomy graph
status: todo
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:00:53Z
parent: linny-mcp-server-gupt
blocked_by:
    - linny-mcp-server-0gdv
    - linny-mcp-server-v1py
---

Standalone cmd/lindexer package. Parse YAML front matter directly, honour slug convention, build the taxonomy graph (terms, membership, co-occurrence, related). Scan for committed conflict markers and report loudly.

**OpenSpec change:** `indexer-frontmatter-taxonomy`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/indexer-frontmatter-taxonomy/tasks.md`. Ships with tests._
