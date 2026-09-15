# mcp-read-tools Specification

## Purpose
TBD - created by archiving change mcp-read-tools. Update Purpose after archive.

## Requirements

### Requirement: MCP read tools are served behind auth

The server SHALL expose an MCP endpoint at `/mcp` using the official MCP Go SDK
streamable-HTTP transport, behind bearer authentication. Read/navigate tools SHALL
include at least `search`, `get_doc`, `list_taxonomies`, `terms`, and `docs_by_term`.

#### Scenario: Tools require authentication

- **WHEN** an unauthenticated request hits `/mcp`
- **THEN** it receives `401` before any tool runs

#### Scenario: Core tools are registered

- **WHEN** an authenticated MCP session lists tools
- **THEN** `search`, `get_doc`, `list_taxonomies`, `terms`, and `docs_by_term` are
  present

### Requirement: Every tool enforces the caller's scopes in SQL

Each read tool SHALL restrict results to the caller's scope using the authz SQL
filter (deny-by-default, cross-term deny), never a post-filter. A denied document
SHALL be indistinguishable from a missing one.

#### Scenario: Scoped search hides denied docs

- **WHEN** a caller whose scope denies `tags:health` runs `search`
- **THEN** documents tagged `health` do not appear and no "hidden results" hint is given

#### Scenario: get_doc on a denied slug reports not found

- **WHEN** `get_doc` is called for a document the caller's scope denies
- **THEN** the tool reports the document was not found

### Requirement: Every tool response is redacted

Read tools SHALL pass all returned content — search snippets and titles, document
bodies and front matter — through the egress redactor before returning it.

#### Scenario: Secret in a readable note is redacted

- **WHEN** `get_doc` returns a note whose body contains a credential
- **THEN** the returned body has that credential replaced by a redaction placeholder

### Requirement: Term listings do not leak denied terms

`terms` SHALL return only terms that have at least one document readable by the
caller, so terms consisting solely of denied documents are not exposed.

#### Scenario: Fully-denied term is absent

- **WHEN** every document under a term is denied to the caller
- **THEN** that term is absent from the `terms` result

### Requirement: Tool surface is documented and versioned

The tool names, arguments, and semantics SHALL be recorded in `docs/tools.md`, which
notes that tool names are an API and are versioned there.

#### Scenario: docs/tools.md lists the tools

- **WHEN** `docs/tools.md` is read
- **THEN** it documents each shipped tool and its arguments

### Requirement: get_doc reports a content hash and redaction status

`get_doc` SHALL return a `content_hash` for the document, computed from the file on
disk rather than from the index, and a `redacted` flag reporting whether the returned
body differs from the stored body because of egress redaction. Both fields are
additive; existing fields SHALL be unchanged.

#### Scenario: Hash is usable as a write base version

- **WHEN** `get_doc` returns a `content_hash` and `update_doc` is then called with it
  as `base_hash` while the file is untouched
- **THEN** the base version is accepted

#### Scenario: Hash tracks the file, not the index

- **WHEN** a document on disk has changed but the index has not yet been rebuilt
- **THEN** `get_doc` reports the hash of the current file

#### Scenario: Redaction is disclosed

- **WHEN** `get_doc` returns a body the redactor altered
- **THEN** `redacted` is true, signalling that whole-body replacement is unavailable
  for that document

### Requirement: starred_docs lists the documents flagged as starred

The server SHALL expose an MCP read tool `starred_docs` that takes no arguments and
returns the documents whose front matter carries `starred: true`, each with its
filename and title. Documents without the flag SHALL NOT appear.

#### Scenario: Starred documents are returned with titles

- **WHEN** `starred_docs` is called on a corpus containing starred documents
- **THEN** every starred document the caller may read is returned with its filename
  and its title

#### Scenario: Unstarred documents are excluded

- **WHEN** a document carries no `starred` key, or carries `starred: false`
- **THEN** it does not appear in the result

#### Scenario: No starred documents

- **WHEN** no document in the corpus is starred
- **THEN** the result is an empty list rather than an error

### Requirement: Scope is enforced in SQL and never disclosed

`starred_docs` SHALL restrict its results with the caller's readable-filenames
predicate inside the query, never by filtering returned rows. The result SHALL NOT
include a count, or any other measure, of the documents scope removed. The result
SHALL carry a constant `scope_filtered` flag reporting that results reflect the
caller's scopes.

#### Scenario: A denied starred document is absent

- **WHEN** a caller denied a taxonomy calls `starred_docs` and a starred document is a
  member of that taxonomy
- **THEN** that document is absent from the result

#### Scenario: Nothing reveals how much was hidden

- **WHEN** a caller whose scopes exclude some starred documents calls `starred_docs`
- **THEN** the result carries no count or other indication of how many documents were
  excluded, and is shaped identically to a result where nothing was excluded

#### Scenario: The scope flag is constant

- **WHEN** a caller holding `read:*` calls `starred_docs`
- **THEN** `scope_filtered` is still reported as true

### Requirement: Titles are redacted and ordering is deterministic

`starred_docs` SHALL pass every returned title through the egress redactor, and SHALL
return documents ordered by title so repeated calls agree.

#### Scenario: A title carrying a credential is redacted

- **WHEN** a starred document's title contains a detectable credential
- **THEN** the returned title is redacted

#### Scenario: Ordering is stable

- **WHEN** `starred_docs` is called twice with no intervening write
- **THEN** both results list the documents in the same title order
