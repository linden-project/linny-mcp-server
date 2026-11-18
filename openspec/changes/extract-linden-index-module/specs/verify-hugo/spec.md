# verify-hugo Specification

## REMOVED Requirements

### Requirement: verify --hugo builds and diffs the Hugo reference

**Reason**: The embedded Hugo reference site and the comparison move to
`linden-index`, which owns the index format.

**Migration**: See the `verify-hugo` capability in `linden-index`. The dev shell
there provides `hugo`.

### Requirement: Subset comparison ignores Hugo-only outputs

**Reason**: Comparison rules move with the verifier.

**Migration**: See `linden-index`.

### Requirement: L1 term-config lookup uses the singular taxonomy name

**Reason**: This is a format rule owned by the engine.

**Migration**: See `linden-index`.
