## ADDED Requirements

### Requirement: Hardened systemd unit

The NixOS module's systemd unit SHALL apply a hardened sandbox: `ProtectSystem=strict`
with `ReadWritePaths` limited to the notebook corpus and state directories,
`PrivateTmp`, `NoNewPrivileges`, `RestrictAddressFamilies` limited to
`AF_INET AF_INET6 AF_UNIX`, `SystemCallFilter=@system-service`, an empty
`CapabilityBoundingSet`, `LockPersonality`, and `MemoryDenyWriteExecute`.

#### Scenario: Hardening directives present

- **WHEN** the module is evaluated into a NixOS system
- **THEN** the `linny-mcp` unit's serviceConfig includes the hardening directives
  above, with `ReadWritePaths` covering each notebook's corpus and state dir

### Requirement: ntfy option surfaced without secrets

The module SHALL expose an `ntfyTopicURL` option and pass it through the generated
config. Secrets SHALL NOT appear in any option: `tokensFile` remains a path.

#### Scenario: ntfy flows through config

- **WHEN** `ntfyTopicURL` is set and the module is evaluated
- **THEN** the generated config JSON carries that URL and no token value appears in
  any option
