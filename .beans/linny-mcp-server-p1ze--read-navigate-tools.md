---
# linny-mcp-server-p1ze
title: Read / navigate tools
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T19:53:10Z
parent: linny-mcp-server-s9mf
blocked_by:
    - linny-mcp-server-z4gy
    - linny-mcp-server-5lrx
    - linny-mcp-server-ljtf
---

search, list_taxonomies, terms, docs_by_term, co_occurring_terms, related, get_doc, due_this_week, open_items. Ranked snippets with front-matter headers, not grep dumps. Confirm names; version in docs/tools.md.

**OpenSpec change:** `mcp-read-tools`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/mcp-read-tools/tasks.md`. Ships with tests._

## Summary of Changes

Wired the MCP read/navigate tools over the official MCP Go SDK (v1.7.0) streamable-HTTP transport, mounted at /mcp behind bearer auth. Per request, getServer reads the authenticated identity, parses its scopes, and builds an MCP server whose tool handlers are bound to that scope filter. Tools (v1): search, get_doc, list_taxonomies, terms, docs_by_term. Every tool filters via the authz SQL subquery (deny-by-default, cross-term deny; denied==not-found) and pipes content fields (titles, snippets, body, props) through the egress redactor. Added TermsForTaxonomyScoped so term listings only expose terms with a readable doc. docs/tools.md records the surface (v1) and reserves the planned tool names.

Verified: handler unit tests (redaction on get_doc, work+health denial across get_doc/docs_by_term/terms, deny-by-default empties, scoped search) AND a full end-to-end test using the SDK client over httptest with a bearer-injecting RoundTripper — lists all five tools, get_doc on a health-denied doc returns not-found, and the fake-secrets note comes back redacted; unauthenticated /mcp returns 401. nix vendorHash updated for the SDK tree; nix flake check: all checks passed.
