---
# linny-mcp-server-u33m
title: NixOS module + hardened systemd unit
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T20:52:19Z
parent: linny-mcp-server-xwyf
blocked_by:
    - linny-mcp-server-gvc5
---

nixosModules.linny-mcp options: enable, corpusPath, stateDir, listenAddress, port, tokensFile, user, group, logLevel, readOnly. Hardened unit: dynamic user, ProtectSystem=strict, ReadWritePaths=corpus+stateDir, PrivateTmp, NoNewPrivileges, RestrictAddressFamilies, SystemCallFilter=@system-service, CapabilityBoundingSet=, LockPersonality, MemoryDenyWriteExecute. tokensFile is a PATH from age.secrets; no token in any Nix option.

**OpenSpec change:** `nixos-module-hardened`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/nixos-module-hardened/tasks.md`. Ships with tests._

## Summary of Changes

Hardened the NixOS module systemd unit per briefing section 9: ProtectSystem=strict with ReadWritePaths limited to each notebook corpus + state dir, PrivateTmp, PrivateDevices, NoNewPrivileges, RestrictAddressFamilies=AF_INET/AF_INET6/AF_UNIX, SystemCallFilter=@system-service ~@privileged, empty CapabilityBoundingSet + AmbientCapabilities, LockPersonality, MemoryDenyWriteExecute, plus ProtectHome/Kernel*/ControlGroups/Clock/Proc, RestrictNamespaces/SUIDSGID/Realtime, UMask=0077. Added the ntfyTopicURL option (plain POST URL, not a secret) and threaded it into the generated config JSON. tokensFile stays a path from age.secrets — no token value in any option.

Verified by evaluating the module into a NixOS system and asserting the hardening keys and ReadWritePaths (corpus + state) on the unit; nix flake check still evaluates the module and all checks passed. No Go change.
