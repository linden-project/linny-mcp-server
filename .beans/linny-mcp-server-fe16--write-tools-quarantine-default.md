---
# linny-mcp-server-fe16
title: Write tools (quarantine default)
status: todo
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:00:53Z
parent: linny-mcp-server-s9mf
blocked_by:
    - linny-mcp-server-ef2b
    - linny-mcp-server-ixxf
---

create_doc (enforce slug convention), set_front_matter (mirror fred semantics; validate against lindenConfig), unset_front_matter, append_to_doc, archive. Validate -> write atomically -> reindex -> return resulting term membership.

**OpenSpec change:** `mcp-write-tools`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/mcp-write-tools/tasks.md`. Ships with tests._
