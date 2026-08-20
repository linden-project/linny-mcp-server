# mcp-history-tools Specification

## Purpose
TBD - created by archiving change mcp-history-tools. Update Purpose after archive.
## Requirements
### Requirement: History tools over git

The server SHALL expose MCP tools `history(slug, limit?)`, `diff(slug, ref)`, and
`changed_since(since)` that read the corpus git history without modifying it.

#### Scenario: History lists commits for a document

- **WHEN** `history` is called for a document with prior commits
- **THEN** it returns those commits (hash, author, date, subject), newest first

#### Scenario: Diff returns a document's change against a ref

- **WHEN** `diff` is called with a document slug and a git ref
- **THEN** it returns the textual diff of that document between the ref and the
  working tree

### Requirement: History tools enforce scope

Each history tool SHALL enforce the caller's scope: the history or diff of a document
the caller may not read SHALL be reported as not-found, and `changed_since` SHALL
return only the changed documents the caller may read.

#### Scenario: History of a denied document is not-found

- **WHEN** `history` or `diff` is called for a document the caller's scope denies
- **THEN** the tool reports the document was not found (indistinguishable from a
  document that does not exist)

#### Scenario: changed_since is scope-filtered

- **WHEN** `changed_since` would include a document the caller may not read
- **THEN** that document is omitted from the result

### Requirement: History output is redacted

Free-text output of the history tools (commit subjects and diff hunks) SHALL pass
through the egress redactor, because a diff can reintroduce a secret that the
current-state read would have stripped.

#### Scenario: Secret in a diff is redacted

- **WHEN** `diff` output contains a credential
- **THEN** that credential is replaced by a redaction placeholder

### Requirement: Ref and date inputs cannot be interpreted as flags

`ref` and `since` inputs SHALL be rejected if they begin with `-`, so a caller cannot
smuggle a git option through them.

#### Scenario: Flag-like ref rejected

- **WHEN** `diff` is called with a `ref` that begins with `-`
- **THEN** the call fails with an error and no git command runs

