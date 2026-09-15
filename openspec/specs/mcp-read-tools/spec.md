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
