# authorization Specification

## Purpose
TBD - created by archiving change authz-scopes. Update Purpose after archive.
## Requirements
### Requirement: Deny by default

Access SHALL be denied unless explicitly granted. A token with no `read` scope SHALL
be able to read no documents.

#### Scenario: No read scope reads nothing

- **WHEN** a scope set has no `read:*`/`read:taxonomy:…` rule
- **THEN** the readable-documents query returns an empty set

### Requirement: Scope vocabulary

The parser SHALL accept `read:*`, `read:taxonomy:<tax>`,
`read:taxonomy:<tax>:<term>`, `deny:taxonomy:<tax>`, `deny:taxonomy:<tax>:<term>`,
`write:inbox`, `write:*`, and `delete:*`. Unknown scopes SHALL be rejected with an
error.

#### Scenario: Valid scopes parse

- **WHEN** parsing `["read:*", "deny:taxonomy:tags:health", "write:inbox"]`
- **THEN** parsing succeeds and yields one read-allow, one deny, and one write rule

#### Scenario: Unknown scope rejected

- **WHEN** parsing `["frobnicate:everything"]`
- **THEN** parsing returns an error

### Requirement: Filtering happens in SQL

Read authorization SHALL be applied as a SQL predicate at query time — a subquery of
readable filenames injected into the store's queries — never as a post-filter over
already-returned rows.

#### Scenario: Filter compiles to a subquery

- **WHEN** `ReadableFilenamesSQL()` is called on a non-empty scope set
- **THEN** it returns a `SELECT filename FROM docs …` subquery and its bound args

### Requirement: Deny is evaluated across all of a document's terms

A document SHALL be excluded if **any** of its (taxonomy, term) memberships matches a
deny rule, even when a read-allow rule (including `read:*`) would otherwise grant it.

#### Scenario: work+health excluded when health denied

- **WHEN** the scope set is `read:*` plus `deny:taxonomy:tags:health` and a document
  is tagged both `tags:work` and `tags:health`
- **THEN** that document is not returned by any scoped query

#### Scenario: work-only still visible

- **WHEN** the same scope set is used and a document is tagged only `tags:work`
- **THEN** that document is returned

### Requirement: Existence is never leaked

A scoped read of a denied or non-existent document SHALL be indistinguishable:
`GetDocScoped` SHALL report not-found for a denied document, and scoped searches SHALL
never indicate that results were hidden.

#### Scenario: Denied document reads as missing

- **WHEN** `GetDocScoped` is called for a document the scope denies
- **THEN** it returns not-found, exactly as for a document that does not exist

