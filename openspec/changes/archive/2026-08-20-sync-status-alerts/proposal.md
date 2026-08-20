## Why

When the working tree breaks while the user is away from his phone, two things must
happen (briefing §6): an agent asking `sync_status()` must be told the server is
degraded (a DoD item — "a conflicted tree puts the server read-only, and
`sync_status()` says so"), and the user must be **alerted out-of-band** via
self-hosted ntfy — never into the corpus, which is the wrong channel when the corpus
is what's broken.

## What Changes

- Add a `sync_status` MCP tool that returns the live git-safety state: degraded flag,
  conflicted flag + paths, in-progress operation, detached-HEAD, forced-read-only, and
  a human reason. It is operational (no taxonomy scope needed) and available to any
  authenticated caller.
- Add `internal/alert`: an `Alerter` interface with an `NtfyAlerter` (POSTs a title +
  body to a self-hosted ntfy topic URL) and a `NopAlerter`, plus a `DegradedMonitor`
  that fires an alert on the **transition** into degraded mode (and a recovery notice
  on the way out) — not repeatedly while degraded.
- Wire `serve`: a background poller ticks the monitor; the alerter is the ntfy one
  when `ntfyTopicURL` is configured, else a no-op. Add `ntfyTopicURL` to the config
  model (path/URL only — no secrets in Nix options; the module change in E0601 exposes
  the option).

## Capabilities

### New Capabilities
- `sync-status-alerts`: the `sync_status` tool and ntfy degraded-transition alerting.

### Modified Capabilities
- `configuration`: adds `ntfyTopicURL`.
- `mcp-read-tools`: registers the operational `sync_status` tool.

## Impact

- New: `internal/alert/**`. Modified: `internal/mcp` (sync_status tool + status
  wiring), `internal/config` (ntfyTopicURL), `cmd/linny-mcp` serve (monitor poller).
  Standard library only (ntfy is a plain HTTP POST).
- The `verify_index` operational tool and the full ntfy NixOS wiring are follow-ups
  (E0504 / E0601).
