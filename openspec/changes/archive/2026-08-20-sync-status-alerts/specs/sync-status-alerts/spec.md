## ADDED Requirements

### Requirement: sync_status tool reports the live tree state

The server SHALL expose an authenticated `sync_status` MCP tool that returns the live
git-safety state: degraded, conflicted (with paths), in-progress operation,
detached-HEAD, forced-read-only, and a human-readable reason.

#### Scenario: Conflicted tree is reported degraded

- **WHEN** the working tree contains committed conflict markers and `sync_status` is
  called
- **THEN** it returns `degraded: true`, `conflicted: true`, and lists the paths

#### Scenario: Clean tree is reported healthy

- **WHEN** the working tree is clean and `sync_status` is called
- **THEN** it returns `degraded: false`

### Requirement: Alert on transition into degraded mode

The degraded monitor SHALL alert the configured `Alerter` when the tree transitions
from clean to degraded, and SHALL NOT re-alert on every check while it stays
degraded. It SHALL send a recovery notice when the tree becomes clean again.

#### Scenario: Single alert per degradation

- **WHEN** the tree becomes degraded and the monitor is checked several times
- **THEN** exactly one degraded alert is sent

#### Scenario: Recovery notice

- **WHEN** the tree returns to clean after being degraded
- **THEN** a recovery alert is sent

### Requirement: Alerts never touch the corpus

Alerts SHALL be delivered out-of-band (ntfy HTTP POST) and SHALL NOT be written into
the corpus.

#### Scenario: ntfy is an external POST

- **WHEN** the ntfy alerter fires
- **THEN** it issues an HTTP POST to the configured topic URL and writes nothing to
  the corpus

### Requirement: Alerting is optional and off by default

When no ntfy topic URL is configured, alerting SHALL be a no-op and the server SHALL
run normally.

#### Scenario: No ntfy configured

- **WHEN** `ntfyTopicURL` is empty
- **THEN** the monitor uses a no-op alerter and the server still serves `sync_status`
