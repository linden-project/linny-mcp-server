# hostile-corpus-defenses Specification

## Purpose
TBD - created by archiving change hostile-corpus-defenses. Update Purpose after archive.
## Requirements
### Requirement: Append-only audit log outside the corpus

Writes SHALL be recorded to an append-only audit log stored outside the corpus (in
`stateDir`). Each entry SHALL capture the time, identity, tool, target slug, a diff,
and the outcome. Existing entries SHALL never be modified or removed by the logger.

#### Scenario: Entries accumulate append-only

- **WHEN** two audit entries are appended
- **THEN** both are present, in order, and the first entry is unchanged

#### Scenario: Log lives outside the corpus

- **WHEN** the audit log path is resolved
- **THEN** it is under the state directory, not inside the corpus working tree

### Requirement: Agent writes are quarantined by default

The quarantine policy SHALL, by default, place agent-created documents into a
quarantine taxonomy/term (default taxonomy `status`, term `agent-draft`). Promotion
out of quarantine is a separate action, not a side effect of creation.

#### Scenario: New document is quarantined

- **WHEN** the policy is applied to a new document's front matter
- **THEN** that front matter is a member of the quarantine term

#### Scenario: Quarantine is detectable

- **WHEN** a document carries the quarantine term
- **THEN** `IsQuarantined` reports true

### Requirement: Destructive tools require out-of-band confirmation

The policy SHALL mark destructive operations (delete, bulk-retag) as requiring
out-of-band confirmation, so they cannot be performed by an in-band tool call alone.

#### Scenario: Delete needs confirmation

- **WHEN** the policy is asked whether `delete` requires confirmation
- **THEN** it returns true

### Requirement: Returned bodies are wrapped in data delimiters

Returned note bodies SHALL be wrapped in explicit data-delimiter markers that signal
the content is data and not instructions. Any occurrence of the marker text inside
the body SHALL be stripped before wrapping so the framing cannot be forged.

#### Scenario: Body is delimited

- **WHEN** a document body is returned by `get_doc`
- **THEN** it is enclosed between the begin and end data-delimiter markers

#### Scenario: Forged delimiter is neutralized

- **WHEN** a body itself contains the end-delimiter marker text
- **THEN** that occurrence is removed before wrapping, leaving exactly one framing pair

### Requirement: No tool may widen its own scope

A request's authorization scope SHALL be fixed at construction from the caller's
token; there SHALL be no API by which a tool changes the scope it runs under.

#### Scenario: Scope is immutable per request

- **WHEN** a tool executes
- **THEN** it uses the scope compiled from the token and cannot broaden it

