# index-watch Specification

## REMOVED Requirements

### Requirement: Debounced change coalescing

**Reason**: Watch mode moves to the extracted CLI.

**Migration**: See the `index-watch` capability in `linden-index`. The debounce
window is unchanged.

### Requirement: Watch triggers a rebuild on corpus change

**Reason**: Watch rebuilds are driven by the extracted engine.

**Migration**: See `linden-index`.

### Requirement: watch CLI builds then watches

**Reason**: The `watch` subcommand moves to the `lindexer` binary.

**Migration**: Run `lindexer watch --corpus <c> --state-dir <s> --index <dir>` from
the `linden-index` release.
