## Why

The deployment target is an old MacBook running Linux (NixOS/systemd) reachable only
via loopback and the Nebula mesh, fronted by a Hetzner TLS proxy. The briefing (§9)
specifies a fully hardened systemd unit. The scaffold module has the complete option
interface but only a minimal unit; this change lands the hardening so the DoD item
"server starts from the NixOS module" ships something safe to expose.

## What Changes

- Harden `systemd.services.linny-mcp` with the briefing's directives: dedicated user,
  `ProtectSystem=strict`, `ReadWritePaths` limited to each notebook's corpus + state
  dir, `PrivateTmp`, `NoNewPrivileges`, `RestrictAddressFamilies=AF_INET AF_INET6
  AF_UNIX`, `SystemCallFilter=@system-service` (+ `~@privileged`),
  `CapabilityBoundingSet=` (empty), `LockPersonality`, `MemoryDenyWriteExecute`, plus
  the usual `ProtectHome`/`ProtectKernelTunables`/`ProtectControlGroups`/
  `RestrictNamespaces`/`RestrictSUIDSGID`/`ProtectProc` set.
- Add the `ntfyTopicURL` module option and include it in the generated config JSON
  (a plain POST URL — not a secret, safe in the store).
- Keep the secret-hygiene rule: `tokensFile` remains a path from `age.secrets`; no
  token value is ever a Nix option.

## Capabilities

### New Capabilities

### Modified Capabilities
- `nix-packaging`: the NixOS module ships a hardened unit and the `ntfyTopicURL`
  option.

## Impact

- Modified: `nix/module.nix` only. No Go change (coverage/tests unaffected).
- Verified by evaluating the module into a NixOS system and asserting the hardening
  keys are present on the unit.
