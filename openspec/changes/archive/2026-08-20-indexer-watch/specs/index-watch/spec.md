## ADDED Requirements

### Requirement: Debounced change coalescing

Rapid bursts of change signals SHALL be coalesced so the rebuild callback fires once
per quiet period, not once per event.

#### Scenario: Burst fires once

- **WHEN** several change signals arrive within the debounce window
- **THEN** the rebuild callback is invoked exactly once after the window elapses

### Requirement: Watch triggers a rebuild on corpus change

While watching, a change to a file under the content directory SHALL trigger a
rebuild callback (after the debounce window).

#### Scenario: Editing a record triggers a rebuild

- **WHEN** a record under the content dir is written while the watcher runs
- **THEN** the rebuild callback is invoked

### Requirement: watch CLI builds then watches

`lindexer watch --corpus <c> --state-dir <s>` SHALL build the index once, then watch
for changes and refresh the persisted store on each change batch, until interrupted.

#### Scenario: Requires a state dir

- **WHEN** `lindexer watch` is run without `--state-dir`
- **THEN** it exits non-zero with an error
