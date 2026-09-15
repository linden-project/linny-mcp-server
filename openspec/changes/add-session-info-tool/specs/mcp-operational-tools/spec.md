## ADDED Requirements

### Requirement: session_info describes the current connection

The server SHALL expose an MCP tool `session_info` that takes no arguments and returns
the server name and version, the configured notebook name, whether quarantine is
active, whether writes are enabled, and the caller's identity and raw scope strings.

#### Scenario: Server and caller are reported

- **WHEN** an authenticated caller invokes `session_info`
- **THEN** the result reports the server version, the notebook name, and the caller's
  identity and scope strings

#### Scenario: No corpus content is disclosed

- **WHEN** `session_info` returns
- **THEN** the result contains no document count, no taxonomy or term name, no
  document slug, and no filesystem path

### Requirement: Derived permissions come from the enforcing predicates

`session_info` SHALL report the caller's modify permission as one of `all`,
`own-drafts`, or `none`, derived from the same scope predicates the write path uses to
enforce it. It SHALL NOT report a delete permission, which no tool implements.

#### Scenario: write:* reports full modify permission

- **WHEN** the caller holds `write:*`
- **THEN** the reported modify permission is `all`

#### Scenario: write:inbox reports draft-only modify permission

- **WHEN** the caller holds `write:inbox` and not `write:*`
- **THEN** the reported modify permission is `own-drafts`, and creating is reported as
  permitted

#### Scenario: A read-only token reports no modify permission

- **WHEN** the caller holds only read scopes
- **THEN** the reported modify permission is `none` and creating is reported as not
  permitted

### Requirement: session_info answers whether writing is possible now

`session_info` SHALL report a single `can_write_now` verdict that is true only when
write tools are enabled, the working tree is not degraded, and the caller's scopes
permit some write. When it is false the result SHALL carry a reason naming the
condition that failed.

#### Scenario: Scope is the blocking condition

- **WHEN** writes are enabled, the tree is clean, and the caller holds no write scope
- **THEN** `can_write_now` is false and the reason refers to the caller's scopes

#### Scenario: A degraded tree blocks writing

- **WHEN** the working tree is conflicted and the caller holds `write:*`
- **THEN** `can_write_now` is false and the reason refers to the degraded tree

#### Scenario: Everything permits writing

- **WHEN** writes are enabled, the tree is clean, and the caller holds `write:*`
- **THEN** `can_write_now` is true

### Requirement: Operational state is not recomputed

The degraded and read-only state reported by `session_info` SHALL come from the same
source `sync_status` reports from, and the writes-enabled condition SHALL be the same
predicate that decides whether the write tools are registered.

#### Scenario: The two tools agree

- **WHEN** `sync_status` and `session_info` are called against a degraded tree
- **THEN** both report the tree as degraded

#### Scenario: Registration and report agree

- **WHEN** the server is configured such that the write tools are not registered
- **THEN** `session_info` reports writes as not enabled
