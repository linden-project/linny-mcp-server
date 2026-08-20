---
# linny-mcp-server-0xuw
title: 'Config model: N notebooks + configurable public hostname'
status: completed
type: epic
priority: normal
created_at: 2026-08-20T18:50:08Z
updated_at: 2026-08-20T18:56:02Z
parent: linny-mcp-server-vptn
---

Introduce a config model supporting N notebooks (each with its own corpus path, state dir, and git-safety guard) and a configurable publicHostname (never a hardcoded constant). Update the NixOS module options accordingly. Resolves the two open questions the user answered on 2026-08-20.

**OpenSpec change:** `config-notebooks-hostname`

## Summary of Changes

Added a real Config model in internal/config: server settings (listenAddress, port, tokensFile, logLevel, readOnly) + a configurable publicHostname (no hostname is compiled in) + a list of Notebooks (name/corpusPath/stateDir). Validate() enforces >=1 notebook and unique non-empty names; Load() parses+validates JSON; FromFlags() builds a single-notebook "default" config; Resolve(name) selects by name or defaults to the first. serve gained --config and --notebook and now resolves its notebook and builds the git-safety guard from it. The NixOS module gained publicHostname and a notebooks list (with corpusPath desugaring to a single default notebook), generates the server config JSON in the store (paths only, never a token value), and sets ReadWritePaths per notebook.

Verified: unit tests (validation, JSON round-trip, flag sugar, selection, unset-hostname) + a live NixOS-module eval producing the correct multi-notebook ExecStart and config JSON, which the Go loader then parsed and selected the business notebook from. nix flake check: all checks passed. Resolves the user-answered open questions (configurable hostname; multi-notebook designed in now).
