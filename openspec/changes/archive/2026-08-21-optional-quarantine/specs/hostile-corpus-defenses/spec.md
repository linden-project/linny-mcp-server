## ADDED Requirements

### Requirement: Quarantine-on-create can be disabled

The server SHALL support disabling quarantine-on-create via a `--no-quarantine` flag
(and a `disableQuarantine` config value / NixOS `quarantine = false`). Quarantine is
**on by default**; disabling it SHALL be logged with a prominent warning at startup.

#### Scenario: Disabled skips quarantine

- **WHEN** the server runs with quarantine disabled and `create_doc` is called
- **THEN** the new document is written without the quarantine term and the result
  reports `quarantined: false`

#### Scenario: Default keeps quarantine

- **WHEN** the server runs without the flag
- **THEN** `create_doc` still lands the document in the quarantine term

#### Scenario: Startup warning

- **WHEN** the server starts with quarantine disabled
- **THEN** it emits a warning that agent writes are not quarantined
