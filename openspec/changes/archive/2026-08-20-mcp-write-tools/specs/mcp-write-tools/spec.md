## ADDED Requirements

### Requirement: Write tools exist and require write scope

The server SHALL expose MCP write tools `create_doc`, `append_to_doc`,
`set_front_matter`, `unset_front_matter`, and `archive`. `create_doc` SHALL require
`write:inbox` (or `write:*`); modifying an existing document SHALL require `write:*`,
or `write:inbox` when the document is a quarantined draft. A write to a document the
caller cannot read SHALL be reported as not-found.

#### Scenario: create_doc forbidden without write scope

- **WHEN** a caller with only `read:*` calls `create_doc`
- **THEN** the tool refuses and no file is written

#### Scenario: modifying an existing doc requires write:*

- **WHEN** a caller with only `read:*` calls `set_front_matter` on a readable document
- **THEN** the tool refuses

### Requirement: Agent-created documents are quarantined

`create_doc` SHALL place the new document in the quarantine term by default, and the
result SHALL report the document as quarantined.

#### Scenario: New doc is quarantined

- **WHEN** `create_doc` succeeds
- **THEN** the created document is a member of the quarantine term and the result's
  membership reflects it

### Requirement: Writes are refused while degraded

Write tools SHALL refuse to write while the git working tree is degraded
(conflicted / mid-operation), returning a retryable error, and SHALL leave the corpus
unchanged.

#### Scenario: Create refused during conflict

- **WHEN** the working tree contains committed conflict markers and `create_doc` is
  called
- **THEN** the write is refused and no file is created

### Requirement: Writes are atomic and optimistically concurrent

Writes SHALL be atomic (temp → fsync → rename). `create_doc` SHALL fail if the slug
already exists; edits SHALL fail with a retryable stale-write error if the file
changed since it was read.

#### Scenario: Create does not clobber an existing slug

- **WHEN** `create_doc` targets a slug that already exists
- **THEN** it fails and the existing file is unchanged

### Requirement: Front-matter edits preserve key order

`set_front_matter`, `unset_front_matter`, and `archive` SHALL edit the front matter
surgically (as a YAML node), preserving the order of untouched keys and comments.

#### Scenario: Setting a key keeps existing keys

- **WHEN** `set_front_matter` adds a key to a document
- **THEN** the previously-present keys remain in their original order

### Requirement: Writes reindex and return resulting membership

After a successful write the server SHALL reindex and return the document's resulting
term membership, so the agent observes the effect of its write.

#### Scenario: Setting a taxonomy key updates membership

- **WHEN** `set_front_matter` sets a taxonomy key to a term
- **THEN** the returned membership includes that (taxonomy, term)

### Requirement: Every write is audited

Every write attempt (success, denial, or error) SHALL be recorded to the append-only
audit log outside the corpus, including the tool, slug, and outcome.

#### Scenario: Successful write is audited

- **WHEN** `create_doc` succeeds
- **THEN** an audit entry with tool `create_doc` and outcome `ok` is appended
