## Why

Quarantine-on-create (agent writes land in `status: agent-draft`) is the right default
for an untrusted-agent setup, but it gets in the way when driving the server from a
trusted client during local testing or single-user use. There was no way to turn it
off. This adds an explicit, loudly-warned opt-out.

## What Changes

- `defense.Policy` gains a `Disabled` flag; `ApplyQuarantine` is a no-op when set.
- `config.Config` gains `disableQuarantine` (default false); `serve` gains
  `--no-quarantine`, which sets it and logs a prominent warning at startup.
- The NixOS module gains a `quarantine` option (default **true**); when false the
  generated config sets `disableQuarantine`.
- When disabled, `create_doc` writes the document as-is (no quarantine term) and the
  result reports `quarantined: false`. All other defenses (auth, scopes, redaction,
  git-safety, audit log) are unaffected.

## Capabilities

### Modified Capabilities
- `hostile-corpus-defenses`: quarantine-on-create is now optional (default on).
- `configuration`: adds `disableQuarantine`.

## Impact

- Modified: `internal/defense`, `internal/config`, `cmd/linny-mcp` serve,
  `nix/module.nix`. Default behaviour is unchanged (quarantine stays on). Disabling it
  removes a hostile-corpus defense and is warned about; keep it on in production.
