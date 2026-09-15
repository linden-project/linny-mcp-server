---
# linny-mcp-server-pijf
title: method to show all starred documents
status: completed
type: task
priority: normal
created_at: 2026-08-24T10:11:07Z
updated_at: 2026-09-15T20:50:49Z
---


**OpenSpec change:** `add-starred-docs-tool`

## Summary of Changes

Added the `starred_docs` MCP read tool. The `starred: true` front-matter flag was
already collected by the indexer into `Graph.StarredDocs`, written to the
`docs.starred` column on every build, and emitted as `_index_docs_starred.json` for
linny.vim, but no query ever selected it back: the state was write-only and an agent
could not enumerate starred notes, only see the flag on a document it already held.

`Store.StarredDocsScoped` selects `starred = 1` intersected with the caller's
readable-filenames subquery in SQL, ordered by title, returning filename and title so
the result reads as a list without a get_doc per row. Titles pass through the egress
redactor. The result carries a constant `scope_filtered: true`, telling the caller not
to report the list as exhaustive, and deliberately no count of what scope removed:
that would disclose the existence of denied documents, which the authz model forbids,
and producing it would need a second unscoped query.

Scoped to documents. Starred taxonomies and terms are indexed too but are declared in
lindenConfig rather than by documents, so either can exist with no readable members;
exposing them needs its own answer to the existence-leak question, recorded as a
non-goal.

Verified: 7 handler and protocol tests (starred returned with titles, unstarred and
`starred: false` excluded, corpus-generated stars found, denied documents absent with
an identically shaped response, titles redacted, ordering stable and by title,
deny-by-default empty, e2e over MCP). Mutation-tested the scope predicate: replacing
the AND with an OR leaked the whole corpus and failed both the leak and
deny-by-default tests. Coverage 83.1% total, internal/mcp 84.8%. nix flake check green.
