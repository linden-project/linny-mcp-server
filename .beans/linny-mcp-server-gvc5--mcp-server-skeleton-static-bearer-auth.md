---
# linny-mcp-server-gvc5
title: MCP server skeleton + static bearer auth
status: todo
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:00:53Z
parent: linny-mcp-server-vptn
blocked_by:
    - linny-mcp-server-0gdv
---

MCP server (official Go SDK) with Authenticator interface; StaticTokenAuthenticator the only impl. Bearer via ConstantTimeCompare; 401 empty body no timing signal; >=32B CSPRNG base64url tokens; gen-token helper; refuse non-loopback/non-mesh bind without override; hashed token file with per-token scopes metadata.

**OpenSpec change:** `server-skeleton-auth`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/server-skeleton-auth/tasks.md`. Ships with tests._
