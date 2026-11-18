# index-format Specification

## REMOVED Requirements

### Requirement: Index output location and layout

**Reason**: The canonical format document moves to `linden-index`, which owns the
emitter.

**Migration**: See the `index-format` capability in `linden-index`. Layout is
unchanged.

### Requirement: Taxonomy and term path normalization

**Reason**: Path normalization is implemented by the extracted emitter.

**Migration**: See `linden-index`. Paths are byte-identical to what this
repository produced.

### Requirement: `_index_taxonomies.json` is an array of taxonomy keys

**Reason**: Emitter ownership moves to `linden-index`.

**Migration**: See `linden-index`. The file and its shape are unchanged.

### Requirement: `_index_docs_with_props.json` maps filename to full front matter

**Reason**: Emitter ownership moves to `linden-index`.

**Migration**: See `linden-index`. The mapping is unchanged.

### Requirement: `_index_docs_tasks_count.json` maps filename to task counts

**Reason**: Emitter ownership moves to `linden-index`.

**Migration**: See `linden-index`. The mapping is unchanged.

### Requirement: `_index_docs_starred.json` is an array of starred filenames

**Reason**: Emitter ownership moves to `linden-index`.

**Migration**: See `linden-index`. The array is unchanged.

### Requirement: Per-taxonomy L1 index maps term to term config

**Reason**: Emitter ownership moves to `linden-index`.

**Migration**: See `linden-index`. The mapping is unchanged.

### Requirement: Per-term L2 index is an array of member filenames

**Reason**: Emitter ownership moves to `linden-index`.

**Migration**: See `linden-index`. The array is unchanged.

### Requirement: `_index_terms_starred.json` is an array of {taxonomy, term}

**Reason**: Emitter ownership moves to `linden-index`.

**Migration**: See `linden-index`. The array is unchanged.

### Requirement: `_indexer_info.json` reports indexer identity

**Reason**: The identity file is written by the engine that produces it.

**Migration**: See `linden-index`. `product_name` stays `lindexer` and
`product_version` now follows the `linden-index` release.

### Requirement: Committed git conflict markers are reported, not indexed silently

**Reason**: Conflict detection is part of the build.

**Migration**: See `linden-index`.

### Requirement: Implementable from the document alone

**Reason**: Each project documents what it owns; the index format is no longer
implemented here.

**Migration**: The `linden-index` repository carries the format document. This
repository keeps only its MCP-facing documentation.
