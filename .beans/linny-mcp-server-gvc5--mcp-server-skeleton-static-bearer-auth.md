---
# linny-mcp-server-gvc5
title: MCP server skeleton + static bearer auth
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:36:55Z
parent: linny-mcp-server-vptn
blocked_by:
    - linny-mcp-server-0gdv
---

MCP server (official Go SDK) with Authenticator interface; StaticTokenAuthenticator the only impl. Bearer via ConstantTimeCompare; 401 empty body no timing signal; >=32B CSPRNG base64url tokens; gen-token helper; refuse non-loopback/non-mesh bind without override; hashed token file with per-token scopes metadata.

**OpenSpec change:** `server-skeleton-auth`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/server-skeleton-auth/tasks.md`. Ships with tests._

## Summary of Changes

Delivered static bearer authentication and the server skeleton. internal/auth: Authenticator interface with StaticTokenAuthenticator (SHA-256-hashed token file, per-token name+scopes), crypto/subtle.ConstantTimeCompare over all records (no early-exit), fail-closed bearer parsing, and an HTTP middleware returning a bare 401 (empty body, no timing/detail signal) on every failure mode. linny-mcp gen-token emits a >=32B CSPRNG base64url token once plus its hashed record line. internal/config.CheckBindAddress refuses public/unspecified binds unless overridden (loopback, RFC1918, CGNAT-mesh, link-local allowed). internal/mcp.Server exposes unauthenticated /healthz and an authenticated /mcp placeholder. linny-mcp serve wires it all with the NixOS-module flags.

Verified: unit tests (valid auth, indistinguishable 401s, token hashing/no-raw-token, bind matrix, healthz-open, gen-token round-trip) + a live smoke test (loopback bind, healthz 200 unauth, valid token -> 200 identity, public bind refused). nix flake check: all checks passed. Follow-up: official MCP Go SDK + read tools on /mcp.
