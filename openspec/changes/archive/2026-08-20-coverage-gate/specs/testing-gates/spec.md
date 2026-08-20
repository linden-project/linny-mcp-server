## ADDED Requirements

### Requirement: Coverage gate in nix flake check

`nix flake check` SHALL include a coverage check that measures total statement
coverage across the module (using `-coverpkg=./...` so end-to-end tests count toward
the packages they exercise) and FAILS when total coverage is below 70%.

#### Scenario: Below the floor fails

- **WHEN** total statement coverage is below 70%
- **THEN** the coverage check fails and `nix flake check` fails

#### Scenario: At or above the floor passes

- **WHEN** total statement coverage is at least 70%
- **THEN** the coverage check passes

### Requirement: Git-backed tests run under the coverage check

The coverage check SHALL make `git` available so the git-backed history and
working-tree tests execute (rather than skipping), so their coverage is counted.

#### Scenario: History tests are not skipped

- **WHEN** the coverage check runs
- **THEN** the `git`-dependent tests execute and contribute to coverage

### Requirement: Testing agreement is documented

`openspec/project.md` SHALL state the testing agreement: what unit and e2e testing
mean for this project, the required test classes, and the 70% coverage floor enforced
by the gate.

#### Scenario: Agreement present

- **WHEN** `openspec/project.md` is read
- **THEN** it documents the unit/e2e expectations and the 70% coverage gate
