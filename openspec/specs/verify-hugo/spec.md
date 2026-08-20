# verify-hugo Specification

## Purpose
TBD - created by archiving change verify-run-hugo. Update Purpose after archive.
## Requirements
### Requirement: verify --hugo builds and diffs the Hugo reference

`lindexer verify --corpus <c> --hugo` SHALL assemble a Hugo site from the vendored
reference layouts/config plus the corpus content and lindenConfig, run the `hugo`
binary to produce a reference index, build our index, diff them, and exit non-zero if
discrepancies remain.

#### Scenario: Load-bearing files match Hugo

- **WHEN** `verify --hugo` runs on the synthetic corpus
- **THEN** the taxonomy list and every term-membership (L2) file produced by our
  indexer match Hugo's

#### Scenario: hugo binary absent

- **WHEN** `verify --hugo` runs and the `hugo` binary is not on PATH
- **THEN** it fails with a clear error naming hugo

### Requirement: Subset comparison ignores Hugo-only outputs

The Hugo diff SHALL compare only the files our indexer emits, ignoring Hugo's
vestigial per-page `<slug>/index.json` outputs, and SHALL normalize away Hugo's
injected built-in params (`draft`, `iscjklanguage`) before comparing `docs_with_props`.

#### Scenario: Per-page files are not reported

- **WHEN** the Hugo reference contains per-page `<slug>/index.json` files our indexer
  does not emit
- **THEN** those files are not reported as discrepancies

#### Scenario: Built-in params ignored in props

- **WHEN** Hugo's `_index_docs_with_props.json` carries `draft`/`iscjklanguage` keys
  that our front matter lacks
- **THEN** those keys do not, by themselves, cause a props discrepancy

