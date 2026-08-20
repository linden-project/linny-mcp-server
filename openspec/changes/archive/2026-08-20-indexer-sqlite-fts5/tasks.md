## 1. Store & schema

- [x] 1.1 Add `modernc.org/sqlite`; open with MaxOpenConns(1) + disposable-cache pragmas
- [x] 1.2 Schema: docs, taxonomies, terms, membership, docs_fts (FTS5) + indexes
- [x] 1.3 `Open(path)` creates schema if absent; `Close()`

## 2. Populate

- [x] 2.1 `Populate(graph)` transactional recreate + bulk insert (idempotent)
- [x] 2.2 docs (props_json, tasks, starred); taxonomies (occurring); terms (config_json, starred); membership; docs_fts (title+body)

## 3. Queries

- [x] 3.1 `Search(query, limit)` FTS5 MATCH + bm25 ranking + snippet
- [x] 3.2 `ListTaxonomies`, `TermsForTaxonomy`, `DocsByTerm`
- [x] 3.3 `GetDoc(filename)` → title, props, body

## 4. CLI wiring

- [x] 4.1 `lindexer build --state-dir` persists the store (JSON emit unchanged)
- [x] 4.2 `lindexer search --state-dir <dir> <query>` prints ranked hits

## 5. Tests & gate

- [x] 5.1 Open/create schema; populate; idempotent re-populate
- [x] 5.2 Delete-and-rebuild recovers identical query results
- [x] 5.3 Search ranked + snippets; empty result on no match
- [x] 5.4 ListTaxonomies / DocsByTerm / GetDoc against the synthetic corpus
- [x] 5.5 Update nix vendorHash; nix flake check green
