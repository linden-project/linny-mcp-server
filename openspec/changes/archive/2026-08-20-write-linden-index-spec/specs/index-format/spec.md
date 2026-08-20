## ADDED Requirements

### Requirement: Index output location and layout

A conforming indexer SHALL write all index files under a single index root
directory (the Hugo `publishDir`, named `lindenIndex` by default, resolved by
`linny.vim` as `<notebook>/lindenIndex`). Home-level index files SHALL sit flat at
the index root; per-taxonomy and per-term index files SHALL sit in a nested layout
`<taxonomy>/index.json` and `<taxonomy>/<term>/index.json`.

#### Scenario: Home-level files at root

- **WHEN** the index is built
- **THEN** `_index_taxonomies.json`, `_index_docs_starred.json`,
  `_index_docs_with_props.json`, `_index_docs_with_title.json`,
  `_index_docs_tasks_count.json`, `_indexer_info.json`,
  `_index_taxonomies_starred.json`, and `_index_terms_starred.json` all exist at the
  index root

#### Scenario: Nested taxonomy/term files

- **WHEN** taxonomy `projects` has a term `acme` with member documents
- **THEN** `projects/index.json` and `projects/acme/index.json` exist

### Requirement: Taxonomy and term path normalization

Taxonomy directory segments SHALL be lowercased. Term directory segments SHALL be
lowercased and have spaces replaced with dashes. The document filename
(`basename` including the `.md` extension) is the identity key used across all
document indexes.

#### Scenario: Term with spaces normalized

- **WHEN** taxonomy `subject` has term `Learn Linny`
- **THEN** its member list is written to `subject/learn-linny/index.json`

### Requirement: `_index_taxonomies.json` is an array of taxonomy keys

`_index_taxonomies.json` SHALL be a JSON array of strings, each the plural taxonomy
key (e.g. `tags`, `projects`, `customer`).

#### Scenario: Shape

- **WHEN** the file is parsed
- **THEN** it is a JSON array whose elements are strings

### Requirement: `_index_docs_with_props.json` maps filename to full front matter

`_index_docs_with_props.json` SHALL be a JSON object keyed by document filename,
whose value is that document's complete front matter with keys normalized to
lowercase. Only documents that have a `title` key SHALL be included.

#### Scenario: Included doc

- **WHEN** `my_note.md` has front matter with a `title`
- **THEN** `_index_docs_with_props.json["my_note.md"]` is an object containing all of
  that note's front-matter keys (lowercased)

#### Scenario: Title-less doc excluded

- **WHEN** a document has no `title` key
- **THEN** it does not appear in `_index_docs_with_props.json`

### Requirement: `_index_docs_tasks_count.json` maps filename to task counts

`_index_docs_tasks_count.json` SHALL be a JSON object keyed by document filename,
whose value is `{ "open": <int>, "closed": <int>, "total": <int> }` counting
Markdown task list items (`- [ ]` open, `- [x]` closed). Only documents with
`total > 0` SHALL be included.

#### Scenario: Counts

- **WHEN** a document body has two `- [ ]` and one `- [x]`
- **THEN** its entry is `{ "open": 2, "closed": 1, "total": 3 }`

### Requirement: `_index_docs_starred.json` is an array of starred filenames

`_index_docs_starred.json` SHALL be a JSON array of document filenames for which
front matter `starred` is `true`.

#### Scenario: Starred membership

- **WHEN** `pinned.md` has `starred: true`
- **THEN** `"pinned.md"` appears in `_index_docs_starred.json`

### Requirement: Per-taxonomy L1 index maps term to term config

`<taxonomy>/index.json` SHALL be a JSON object keyed by the terms that actually
occur in that taxonomy, whose value is that term's configuration object taken from
`lindenConfig/L2-CONF-TAX-<taxonomy>-TRM-<term>.yml` (e.g. `title`, `infotext`,
`starred`, `views`). A term with no L2 config SHALL map to an empty object.

#### Scenario: Configured term

- **WHEN** taxonomy `customer` has term `eric` with an L2 config file
- **THEN** `customer/index.json["eric"]` is that config object

#### Scenario: Unconfigured term

- **WHEN** a term occurs but has no L2 config file
- **THEN** its value in the L1 index is `{}`

### Requirement: Per-term L2 index is an array of member filenames

`<taxonomy>/<term>/index.json` SHALL be a JSON array of the document filenames that
belong to that (taxonomy, term).

#### Scenario: Membership list

- **WHEN** `first_note.md` and `address_eric.md` are tagged `customer: eric`
- **THEN** `customer/eric/index.json` contains both filenames

### Requirement: `_index_terms_starred.json` is an array of {taxonomy, term}

`_index_terms_starred.json` SHALL be a JSON array of objects
`{ "taxonomy": <string>, "term": <string> }` for every term whose L2 config has
`starred: true`.

#### Scenario: Shape

- **WHEN** term `learn-linny` under `subject` is starred
- **THEN** the array contains `{ "taxonomy": "subject", "term": "learn-linny" }`

### Requirement: `_indexer_info.json` reports indexer identity

`_indexer_info.json` SHALL be a JSON object identifying the indexer, at minimum
`product_name` and `product_version`, plus resolved `index_dir`, `content_dir`, and
`config_dir` paths. (The Hugo reference emits literal `"TODO"` for the three paths;
a conforming standalone indexer SHALL populate them.)

#### Scenario: Identity present

- **WHEN** the file is parsed
- **THEN** it contains `product_name` and `product_version` string fields

### Requirement: Committed git conflict markers are reported, not indexed silently

While indexing, the indexer SHALL scan tracked content for committed git conflict
markers (`<<<<<<<`, `=======`, `>>>>>>>`) and report them loudly rather than folding
corrupted content into the index silently.

#### Scenario: Conflict marker found

- **WHEN** a document contains a line beginning with `<<<<<<<`
- **THEN** the indexer reports the offending file and marker prominently

### Requirement: Implementable from the document alone

`docs/linden-index-spec.md` SHALL be complete enough that a competent implementer
could write a conforming indexer from it without reading the Hugo templates, and
SHALL list every known ambiguity/deprecation as an explicit open question.

#### Scenario: Deprecations flagged

- **WHEN** the spec describes `_index_docs_with_title.json`
- **THEN** it marks the file DEPRECATED and points to `_index_docs_with_props.json`
  as the replacement source of titles
