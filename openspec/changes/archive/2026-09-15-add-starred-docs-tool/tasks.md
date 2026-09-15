## 1. Scoped query

- [x] 1.1 `Store.StarredDocsScoped(readableSubquery string, subArgs []any)` selecting
      `starred = 1` intersected with the readable-filenames subquery, ordered by title
- [x] 1.2 a small result type carrying filename + title
- [x] 1.3 no second, unscoped query anywhere in the path

## 2. Tool (internal/mcp/tools.go)

- [x] 2.1 `starredDocsOut` with `docs` and a constant `scope_filtered`
- [x] 2.2 handler redacting every title
- [x] 2.3 register `starred_docs` with a description stating results reflect scopes

## 3. Tests

- [x] 3.1 starred documents returned with titles; unstarred excluded
- [x] 3.2 empty corpus returns an empty list, not an error
- [x] 3.3 a denied starred document is absent, and the result shape is identical to
      the unfiltered case (no count, no marker)
- [x] 3.4 `scope_filtered` is true for a `read:*` caller
- [x] 3.5 a credential in a title is redacted
- [x] 3.6 ordering is stable across calls
- [x] 3.7 protocol-level e2e: `starred_docs` over MCP returns the starred set

## 4. Docs & gate

- [x] 4.1 `docs/tools.md`: add `starred_docs` to the read table and note that starred
      taxonomies and terms remain unexposed
- [ ] 4.2 `nix flake check` green (build + tests + lint + coverage floors)
