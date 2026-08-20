---
# linny-mcp-server-5lrx
title: 'Scopes: deny-by-default + SQL filtering'
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T19:44:00Z
parent: linny-mcp-server-fo8d
blocked_by:
    - linny-mcp-server-zin6
    - linny-mcp-server-gvc5
---

Scope vocabulary read:taxonomy:<name>, deny:taxonomy:<name>, write:inbox, write:*, delete:*. Deny by default. Intersection semantics: a doc tagged work+health is excluded when health denied (evaluate deny across ALL terms). Filter in SQL at query time, never post-filter; never leak existence. One token per client.

**OpenSpec change:** `authz-scopes`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/authz-scopes/tasks.md`. Ships with tests._

## Summary of Changes

Added internal/authz: deny-by-default scopes compiled to SQL. Parse() accepts read:*, read:taxonomy:<tax>, read:taxonomy:<tax>:<term>, deny:taxonomy:<tax>, deny:taxonomy:<tax>:<term>, write:inbox, write:*, delete:* and rejects unknown scopes. ReadableFilenamesSQL() emits a `SELECT filename FROM docs d WHERE (allow OR-set) AND NOT (deny OR-set)` subquery with bound args in placeholder order; with no read rule the allow expr is constant-false (deny by default). Deny uses correlated EXISTS over membership so it is evaluated across ALL of a document terms (intersection semantics). internal/index gained scoped queries (SearchScoped, DocsByTermScoped, ListTaxonomiesScoped, GetDocScoped) that inject the subquery as `filename IN (...)`, so filtering is in SQL, never post-filtered; GetDocScoped returns not-found for denied docs so a denied doc is indistinguishable from a missing one (no existence leak). The store does not import authz (opaque SQL + args).

Verified: authz parse/reject, deny-by-default SQL, arg-order-matches-placeholders; and end-to-end on a controlled corpus proving the work+health case (read:* + deny:taxonomy:tags:health excludes a work+health doc from DocsByTerm/Search/GetDoc while a work-only doc stays visible), denied==missing, and read:taxonomy scoping. nix flake check: all checks passed. Consumed by E0501 (read tools).
