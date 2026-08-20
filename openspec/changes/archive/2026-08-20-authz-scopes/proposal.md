## Why

The corpus mixes personal and business material, and an agent must only ever see
what its token is scoped to. Per the briefing (§7.1): deny by default, **filter in SQL
at query time — never post-filter**, and **never leak existence** (no "N results
hidden"). The intersection rule is called out as a known bug generator: a note tagged
both `work` and `health` must be excluded when `health` is denied — the deny is
evaluated across **all** of a document's terms, not any single one. The synthetic
corpus already carries the `work_and_health.md` fixture for exactly this.

## What Changes

- Add `internal/authz`: parse a token's scope strings into a `ScopeSet`. Vocabulary:
  `read:*`, `read:taxonomy:<tax>`, `read:taxonomy:<tax>:<term>` (granular),
  `deny:taxonomy:<tax>`, `deny:taxonomy:<tax>:<term>`, plus `write:inbox`, `write:*`,
  `delete:*` (parsed and retained for the write-tools epic).
- `ScopeSet.ReadableFilenamesSQL()` compiles the read + deny rules into a SQL
  subquery (`SELECT filename FROM docs …`) with bound args: a document is readable iff
  some read-allow rule matches **and** no deny rule matches any of its memberships.
  Deny-by-default: with no read rule, the subquery selects nothing.
- Add scoped store queries (`SearchScoped`, `DocsByTermScoped`, `GetDocScoped`,
  `ListTaxonomiesScoped`) that inject the subquery as `filename IN (…)` so filtering
  happens **in SQL**. `GetDocScoped` returns not-found for a denied document, so a
  denied doc is indistinguishable from a missing one (no existence leak).

## Capabilities

### New Capabilities
- `authorization`: deny-by-default scope evaluation compiled to SQL, with correct
  cross-term intersection semantics and no existence leakage.

### Modified Capabilities
- `index-store`: adds scope-filtered query variants.

## Impact

- New: `internal/authz/**` (replaces the doc.go stub).
- Modified: `internal/index` gains scoped query methods (opaque SQL subquery + args;
  the store does not import authz).
- Consumed by E0501 (read tools): every read passes the token's scope filter.
- Standard library only.
