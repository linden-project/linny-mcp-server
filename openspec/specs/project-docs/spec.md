# project-docs Specification

## Purpose
TBD - created by archiving change docs-future-mipnix. Update Purpose after archive.
## Requirements
### Requirement: Future-work document

`docs/future.md` SHALL exist and enumerate the deliberately out-of-scope work,
including the semantic YAML front-matter git merge driver and local-only embeddings.

#### Scenario: Future doc lists deferred work

- **WHEN** `docs/future.md` is read
- **THEN** it lists the semantic YAML merge driver and states embeddings would be
  local-only

### Requirement: Deployment wiring document

`docs/deploy-mipnix.md` SHALL document (not implement) the mipnix-side wiring: the
flake input, a host-config sketch, the agenix secret sourcing `tokensFile` as a path
with an owner, and the Nebula/Hetzner/ntfy topology.

#### Scenario: Deploy doc reaffirms secret hygiene

- **WHEN** `docs/deploy-mipnix.md` is read
- **THEN** it shows `tokensFile` sourced from an age secret path and states that no
  token value appears in a Nix option

