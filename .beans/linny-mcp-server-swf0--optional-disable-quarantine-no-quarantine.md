---
# linny-mcp-server-swf0
title: 'Optional: disable quarantine (--no-quarantine)'
status: completed
type: feature
priority: normal
created_at: 2026-08-21T11:14:34Z
updated_at: 2026-08-21T11:18:17Z
parent: linny-mcp-server-fo8d
---

A serve flag / config field / NixOS option to turn off quarantine-on-create for trusted-client setups. **OpenSpec change:** `optional-quarantine`

## Summary of Changes

Added an explicit opt-out for quarantine-on-create. defense.Policy.Disabled makes ApplyQuarantine a no-op; config.disableQuarantine (default false) + serve --no-quarantine flag (logs a prominent startup warning); NixOS module quarantine option (default true) -> generated config disableQuarantine. When disabled, create_doc writes the doc as-is (quarantined:false); all other defenses unchanged. Default behaviour unchanged (quarantine on). Tests: defense disabled no-op; mcp create_doc-not-quarantined. nix flake check all green.
