---
# linny-mcp-server-u33m
title: NixOS module + hardened systemd unit
status: todo
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:00:53Z
parent: linny-mcp-server-xwyf
blocked_by:
    - linny-mcp-server-gvc5
---

nixosModules.linny-mcp options: enable, corpusPath, stateDir, listenAddress, port, tokensFile, user, group, logLevel, readOnly. Hardened unit: dynamic user, ProtectSystem=strict, ReadWritePaths=corpus+stateDir, PrivateTmp, NoNewPrivileges, RestrictAddressFamilies, SystemCallFilter=@system-service, CapabilityBoundingSet=, LockPersonality, MemoryDenyWriteExecute. tokensFile is a PATH from age.secrets; no token in any Nix option.

**OpenSpec change:** `nixos-module-hardened`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/nixos-module-hardened/tasks.md`. Ships with tests._
