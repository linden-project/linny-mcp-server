## Why

The briefing (§9) says to **document, not implement**, the mipnix-side deployment
wiring, and (§6) to record the out-of-scope semantic YAML merge driver. Several
changes have also referenced `docs/future.md` as the home for deferred work; it must
exist. This change writes both documents.

## What Changes

- Add `docs/future.md`: the consolidated out-of-scope list (verify_index, incremental
  watch, multi-content-dir, surgical fred-style editor, full lindenConfig validation,
  delete/bulk-retag with confirmation, semantic YAML merge driver, local embeddings,
  remaining navigate tools, OIDCAuthenticator).
- Add `docs/deploy-mipnix.md`: the deployment topology and mipnix wiring — flake
  input, host config sketch, agenix secret with `owner`, Nebula/Hetzner/ntfy, and the
  git-sync relationship — reaffirming the secret-hygiene rule.

## Capabilities

### New Capabilities
- `project-docs`: the future-work and deployment-wiring documents.

### Modified Capabilities

## Impact

- New: `docs/future.md`, `docs/deploy-mipnix.md`. No code change (docs are excluded
  from the build closure; the flake is unaffected).
