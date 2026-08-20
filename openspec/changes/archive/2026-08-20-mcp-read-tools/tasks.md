## 1. SDK + transport

- [x] 1.1 Add github.com/modelcontextprotocol/go-sdk
- [x] 1.2 Mount streamable-HTTP MCP handler at /mcp behind auth.Middleware
- [x] 1.3 Per-request getServer: identity -> authz.Parse -> scope SQL -> bound tool handlers

## 2. Tools

- [x] 2.1 search(query, limit?) -> ranked hits (scoped + redacted snippet/title)
- [x] 2.2 get_doc(slug) -> title/props/body (scoped; denied==not found; redacted)
- [x] 2.3 list_taxonomies() -> scoped taxonomy names
- [x] 2.4 terms(taxonomy) -> scoped terms (TermsForTaxonomyScoped)
- [x] 2.5 docs_by_term(taxonomy, term) -> scoped filenames

## 3. Store + docs

- [x] 3.1 Add TermsForTaxonomyScoped to the store
- [x] 3.2 Write docs/tools.md (tool surface, args, versioning note)

## 4. Tests & gate

- [x] 4.1 Tool handlers: scoped results (work+health denial), denied==not-found
- [x] 4.2 Redaction applied to get_doc body + search snippet
- [x] 4.3 terms hides fully-denied terms
- [x] 4.4 /mcp requires auth (401 unauth); tools registered
- [x] 4.5 Update vendorHash; nix flake check green
