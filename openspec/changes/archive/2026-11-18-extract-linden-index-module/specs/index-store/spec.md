# index-store Specification

## REMOVED Requirements

### Requirement: SQLite+FTS5 store in stateDir

**Reason**: The store is part of the extracted engine and its CLI.

**Migration**: See the `index-store` capability in `linden-index`. Existing state
directories remain readable.

### Requirement: Rebuildable disposable cache

**Reason**: Store lifecycle moves with the engine.

**Migration**: See `linden-index`. Deleting the state directory still forces a
rebuild.

### Requirement: Ranked full-text search

**Reason**: Search ranking is engine behaviour.

**Migration**: Use `lindexer search` or the `linden-index` package. The MCP server
gains no search tool through this change.

### Requirement: Taxonomy, term, and document queries

**Reason**: Query surface moves with the store.

**Migration**: See `linden-index`.

### Requirement: lindexer persists and searches

**Reason**: The `lindexer` CLI moves to the `linden-index` repository.

**Migration**: Install the `lindexer` binary released by `linden-index`.
