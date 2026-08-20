# indexer Specification

## Purpose
TBD - created by archiving change indexer-core. Update Purpose after archive.
## Requirements
### Requirement: Full rebuild from a corpus

The indexer SHALL build a complete index from a corpus directory in a single pass,
producing all index files defined by the `index-format` capability under the index
root. A rebuild SHALL be idempotent: building twice over an unchanged corpus yields
byte-identical JSON (modulo unordered-array equivalence).

#### Scenario: Build produces all files

- **WHEN** `lindexer build --corpus <dir> --index <out>` is run on a valid corpus
- **THEN** the eight home-level `_index_*.json` files and the nested L1/L2 files exist
  under `<out>` and parse as JSON

### Requirement: Front matter is parsed directly

The indexer SHALL parse each record's YAML front matter directly, without depending
on Hugo at runtime. Taxonomy membership SHALL be derived from front-matter keys that
name a taxonomy; scalar values contribute one term, list values contribute many.

#### Scenario: Multi-term membership

- **WHEN** a record has `tags: [work, health]`
- **THEN** it appears in both `tags/work/index.json` and `tags/health/index.json`

### Requirement: Title gating and task counting match the spec

`_index_docs_with_props.json` and `_index_docs_with_title.json` SHALL include only
records that have a `title`. `_index_docs_tasks_count.json` SHALL count `- [ ]`
(open) and `- [x]` (closed) task items and include only records with `total > 0`.

#### Scenario: Title-less record excluded

- **WHEN** a record has no `title`
- **THEN** its filename is absent from `_index_docs_with_props.json`

#### Scenario: Task counts

- **WHEN** a record body has two open and one closed task item
- **THEN** its `_index_docs_tasks_count.json` entry is `{open:2, closed:1, total:3}`

### Requirement: lindenConfig drives L1 and starred indexes

The L1 index value for a term SHALL be that term's `L2-CONF-TAX-<tax>-TRM-<term>.yml`
object, or `{}` when absent. `_index_taxonomies_starred.json` SHALL list taxonomies
whose L1 config has `starred: true`; `_index_terms_starred.json` SHALL list
`{taxonomy, term}` for terms whose L2 config has `starred: true`.

#### Scenario: Configured term surfaces its config

- **WHEN** term `eric` under `customer` has an L2 config with `title: Eric`
- **THEN** `customer/index.json["eric"].title == "Eric"`

### Requirement: Malformed front matter never crashes the build

A record with unparseable front matter SHALL be reported (path + reason) and skipped;
it SHALL NOT abort the build or corrupt other records' index entries.

#### Scenario: Malformed record skipped

- **WHEN** the corpus contains a record with invalid YAML front matter
- **THEN** the build completes, the record is reported, and other records are indexed

### Requirement: Committed conflict markers are reported

During the build the indexer SHALL scan record content for lines beginning with
`<<<<<<<`, `=======`, or `>>>>>>>` and report every offending file prominently.

#### Scenario: Conflict marker detected

- **WHEN** a record contains a committed conflict marker
- **THEN** the build reports that file as conflicted

