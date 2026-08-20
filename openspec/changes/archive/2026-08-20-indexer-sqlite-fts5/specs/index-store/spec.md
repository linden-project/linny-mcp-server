## ADDED Requirements

### Requirement: SQLite+FTS5 store in stateDir

The indexer SHALL persist the taxonomy graph and full text to a SQLite database in
`stateDir`, using `modernc.org/sqlite` (pure Go, no cgo) with an FTS5 virtual table
for full-text search. The database SHALL be treated as a disposable cache.

#### Scenario: Store opens and creates schema

- **WHEN** a store is opened at a path that does not yet exist
- **THEN** the database and its schema (docs, taxonomies, terms, membership, and the
  `docs_fts` FTS5 table) are created

### Requirement: Rebuildable disposable cache

Populating the store from a graph SHALL be a clean, transactional rebuild: running it
again over the same store SHALL yield the same result, and deleting the database file
and rebuilding from the corpus SHALL always be a valid recovery step.

#### Scenario: Re-populate is idempotent

- **WHEN** the store is populated twice from the same graph
- **THEN** the second run leaves the same rows as the first (no duplicates)

#### Scenario: Delete-and-rebuild recovers

- **WHEN** the database file is deleted and the store is populated again
- **THEN** all queries return the same results as before deletion

### Requirement: Ranked full-text search

The store SHALL provide a full-text `Search` over document titles and bodies using
FTS5 `MATCH`, returning results ranked by `bm25` with a snippet and the matching
document filename and title, limited by a caller-supplied limit.

#### Scenario: Search returns ranked hits with snippets

- **WHEN** `Search("backup", 10)` is called and a document's body contains "backup"
- **THEN** that document is returned with a non-empty snippet, best matches first

#### Scenario: No matches

- **WHEN** a query matches nothing
- **THEN** Search returns an empty result set and no error

### Requirement: Taxonomy, term, and document queries

The store SHALL provide `ListTaxonomies`, `TermsForTaxonomy(taxonomy)`,
`DocsByTerm(taxonomy, term)`, and `GetDoc(filename)` (title, front matter, body), so
the read-tool epic can build on SQL queries and later filter in SQL.

#### Scenario: Membership query

- **WHEN** `DocsByTerm("customer", "eric")` is called
- **THEN** it returns exactly the filenames tagged `customer: eric`

#### Scenario: Get a document

- **WHEN** `GetDoc(filename)` is called for an indexed document
- **THEN** it returns that document's title, front matter, and body

### Requirement: lindexer persists and searches

`lindexer build` SHALL persist the store when `--state-dir` is given (independently of
JSON emission), and `lindexer search --state-dir <dir> <query>` SHALL run a full-text
search against the persisted store.

#### Scenario: Build then search from the CLI

- **WHEN** `lindexer build --corpus <c> --state-dir <s>` then
  `lindexer search --state-dir <s> "backup"` are run
- **THEN** the search prints ranked matches from the persisted store
