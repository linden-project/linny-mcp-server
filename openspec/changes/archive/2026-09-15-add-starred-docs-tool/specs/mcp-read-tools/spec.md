## ADDED Requirements

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
