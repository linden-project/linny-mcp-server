---
# linny-mcp-server-d0so
title: mipnix wiring docs & docs/future.md
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:29Z
updated_at: 2026-08-20T20:54:07Z
parent: linny-mcp-server-xwyf
blocked_by:
    - linny-mcp-server-0gdv
---

Document (do NOT implement) mipnix-side wiring: linden repo as flake input, MacBook host config, nebula, ntfy. docs/future.md also records the out-of-scope semantic YAML git merge driver.

**OpenSpec change:** `docs-future-mipnix`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/docs-future-mipnix/tasks.md`. Ships with tests._

## Summary of Changes

Added docs/future.md (consolidated out-of-scope work: verify_index, incremental watch, multi-content-dir, surgical fred-style front-matter editor, full lindenConfig validation, delete/bulk-retag with confirmation, semantic YAML front-matter merge driver, local-only embeddings, remaining navigate tools, OIDCAuthenticator) and docs/deploy-mipnix.md (documentation-only deployment wiring: topology diagram, flake input + overlay, host-config sketch with N notebooks, agenix secret sourcing tokensFile as a path with owner, and the Nebula/Hetzner/ntfy/git-sync relationships). Updated docs/README status table. No code change; docs are excluded from the build closure.

nix flake check: all checks passed.
