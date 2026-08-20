## Why

The store, scopes, and redaction now exist but are unreachable — there is no MCP tool
surface. This change wires the read/navigate tools over the official MCP Go SDK so an
agent (Claude mobile / claude.ai connector) can actually query the notebook. It is the
DoD item "serves the read tools," and it composes the security work: every tool runs
behind bearer auth, filters through the caller's scopes **in SQL**, and passes results
through egress redaction before returning.

## What Changes

- Add the MCP Go SDK (`github.com/modelcontextprotocol/go-sdk` v1.7.0). Mount a
  streamable-HTTP MCP handler at `/mcp`, behind the existing bearer-auth middleware.
- Per request, resolve the caller's identity → parse its scopes → build an MCP server
  whose tool handlers are bound to that scope filter. Deny-by-default is inherited from
  the authz layer.
- Read/navigate tools (v1): `search`, `get_doc`, `list_taxonomies`, `terms`,
  `docs_by_term`. Each applies the scope SQL filter and pipes content fields through
  the redactor. Denied documents are indistinguishable from missing ones.
- Add scoped `TermsForTaxonomyScoped` to the store so term listings only expose terms
  with at least one readable document.
- Record the tool surface in `docs/tools.md` (tool names are an API; versioned there).

The remaining briefing tools (`co_occurring_terms`, `related`, `due_this_week`,
`open_items`) are a documented follow-up; this change lands the core reachable set.

## Capabilities

### New Capabilities
- `mcp-read-tools`: the MCP read/navigate tool surface, each tool enforcing scopes in
  SQL and redacting egress.

### Modified Capabilities
- `index-store`: adds `TermsForTaxonomyScoped`.
- `server-runtime`: `/mcp` now serves the MCP protocol (was a placeholder).

## Impact

- New dependency `github.com/modelcontextprotocol/go-sdk`; `vendorHash` updated.
- New: `internal/mcp/tools.go`, `docs/tools.md`. Modified: `internal/mcp/server.go`,
  `internal/index/query_scoped.go`.
