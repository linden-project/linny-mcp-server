## MODIFIED Requirements

### Requirement: verify --hugo builds and diffs the Hugo reference

`lindexer verify --corpus <c> --hugo` SHALL assemble a Hugo site from the vendored
reference layouts/config plus the corpus content and lindenConfig, run the `hugo`
binary to produce a reference index, build our index, diff them, and exit non-zero if
discrepancies remain. On a well-formed corpus the diff SHALL be empty: our index
reproduces the Hugo reference exactly, including its singular-keyed L1 term-config
lookup.

#### Scenario: Zero drift on a well-formed corpus

- **WHEN** `verify --hugo` runs on a Hugo-buildable synthetic corpus
- **THEN** no discrepancies are reported

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

## ADDED Requirements

### Requirement: L1 term-config lookup uses the singular taxonomy name

The L1 `<tax>/index.json` term-config index SHALL be looked up by the taxonomy's
singular name (`L2-CONF-TAX-<singular>-TRM-<term>`), matching the Hugo reference. When
no config resolves under the singular key, the term's value SHALL be an empty object.

#### Scenario: Singular==plural resolves the config

- **WHEN** taxonomy `customer` (singular == plural) has an L2 config for term `eric`
- **THEN** `customer/index.json["eric"]` is that config object

#### Scenario: Singular≠plural yields empty, matching Hugo

- **WHEN** taxonomy `tags` (singular `tag`) has L2 config files named with the plural
- **THEN** `tags/index.json` maps each term to `{}`, exactly as Hugo emits
