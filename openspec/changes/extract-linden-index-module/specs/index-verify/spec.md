# index-verify Specification

## REMOVED Requirements

### Requirement: Index trees are diffed, arrays as sets

**Reason**: Verification is part of the extracted engine and CLI.

**Migration**: See the `index-verify` capability in `linden-index`. Comparison
semantics are unchanged.

### Requirement: indexer_info identity fields are ignored

**Reason**: Verification rules move with the verifier.

**Migration**: See `linden-index`. Identity fields remain excluded from comparison.

### Requirement: verify CLI reports drift and exits non-zero

**Reason**: The `verify` subcommand moves to the `lindexer` binary.

**Migration**: Run `lindexer verify` from the `linden-index` release.
