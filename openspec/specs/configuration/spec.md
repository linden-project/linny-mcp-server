# configuration Specification

## Purpose
TBD - created by archiving change config-notebooks-hostname. Update Purpose after archive.
## Requirements
### Requirement: N-notebook configuration model

The configuration SHALL support a list of one or more notebooks, each with a unique
non-empty `name`, a `corpusPath`, and a `stateDir`. The model SHALL accept several
notebooks even when only one is served, so multi-notebook support needs no retrofit.

#### Scenario: Multiple notebooks accepted

- **WHEN** a config declares notebooks `personal` and `business`
- **THEN** it validates and both notebooks are addressable by name

#### Scenario: Notebook names must be unique and non-empty

- **WHEN** two notebooks share a name, or a name is empty
- **THEN** config validation fails with a clear error

#### Scenario: At least one notebook required

- **WHEN** a config declares zero notebooks
- **THEN** validation fails

### Requirement: Public hostname is configuration

The public hostname SHALL be a configurable value (`publicHostname`) with no
hardcoded default host. Nothing in the code SHALL embed a specific hostname as a
constant.

#### Scenario: Hostname comes from config

- **WHEN** `publicHostname` is set in config
- **THEN** the server uses that value and never a compiled-in host

#### Scenario: Unset hostname is allowed

- **WHEN** `publicHostname` is omitted
- **THEN** config still validates (the value is simply empty)

### Requirement: Config file loading and flag sugar

The server SHALL load configuration from a JSON file when `--config` is given, and
otherwise construct a single-notebook configuration from the individual `serve`
flags. Explicitly provided values SHALL be validated before the server starts.

#### Scenario: Load from file

- **WHEN** `serve --config config.json` is run with a valid file
- **THEN** the server starts with the notebooks and settings from that file

#### Scenario: Single-notebook flag sugar

- **WHEN** `serve --corpus <dir> --tokens-file <f>` is run with no `--config`
- **THEN** a one-notebook config named `default` is constructed and validated

### Requirement: Notebook selection

When more than one notebook is configured, the server SHALL serve the notebook chosen
by `--notebook <name>`, defaulting to the first when unset. Each served notebook SHALL
get its own git-safety guard.

#### Scenario: Select by name

- **WHEN** two notebooks are configured and `--notebook business` is passed
- **THEN** the server serves the `business` notebook's corpus

