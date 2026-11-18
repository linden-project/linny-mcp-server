## MODIFIED Requirements

### Requirement: Coverage gate enforces a core floor

The `coverage` flake check SHALL fail unless total statement coverage is ≥70% AND each
core package still owned by this module (`internal/authz`, `internal/auth`,
`internal/redact`, `internal/defense`, `internal/gitsafe`, `internal/config`) has ≥80%
own coverage. Packages that moved to `linden-index` are gated in that repository.

#### Scenario: A core package below 80% fails the gate

- **WHEN** a core package's own coverage drops below 80%
- **THEN** the coverage check fails

#### Scenario: Overall and core both satisfied passes

- **WHEN** total coverage is ≥70% and every core package is ≥80%
- **THEN** the coverage check passes
