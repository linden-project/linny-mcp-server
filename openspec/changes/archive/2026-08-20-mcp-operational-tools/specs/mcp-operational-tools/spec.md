## ADDED Requirements

### Requirement: verify_index reports served-index-vs-corpus consistency

The server SHALL expose an authenticated `verify_index` MCP tool that rebuilds the
taxonomy graph from the corpus and compares its document set to the served store,
reporting the corpus and store document counts, documents missing from the store,
stale documents in the store, conflicted paths, and an `in_sync` flag.

#### Scenario: In-sync index reports in_sync

- **WHEN** the served store was populated from the current corpus and `verify_index`
  is called
- **THEN** it reports `in_sync: true` with no missing or stale documents

#### Scenario: Corpus change without reindex is reported

- **WHEN** a new record is added to the corpus but the store has not been rebuilt
- **THEN** `verify_index` reports that record under `missing_from_store` and
  `in_sync: false`

#### Scenario: Conflicted corpus is reported

- **WHEN** the corpus contains a committed conflict marker
- **THEN** `verify_index` lists the conflicted path
