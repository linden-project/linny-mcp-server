## ADDED Requirements

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
