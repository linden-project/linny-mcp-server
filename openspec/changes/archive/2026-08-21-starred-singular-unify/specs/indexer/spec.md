## ADDED Requirements

### Requirement: Config is loaded and keyed by filename

`lindenConfig` SHALL be loaded by scanning `L1-CONF-TAX-*.yml` and
`L2-CONF-TAX-*-TRM-*.yml`, keying each entry by the taxonomy name as it appears in the
filename — the same identifier the reference derives from `.Site.Data`.

#### Scenario: Singular-named config keyed by singular

- **WHEN** the config contains `L2-CONF-TAX-tag-TRM-note.yml`
- **THEN** it is keyed under taxonomy `tag`

### Requirement: Starred indexes are config-derived with no occurrence filter

`_index_taxonomies_starred.json` SHALL list every taxonomy whose L1 config has
`starred: true`, and `_index_terms_starred.json` SHALL list every `{taxonomy, term}`
whose L2 config has `starred: true` — both taken from the config filenames (matching
the singular L1 lookup convention), with no requirement that the term occurs in a
document.

#### Scenario: Starred term reported regardless of occurrence

- **WHEN** an L2 config has `starred: true`
- **THEN** its `{taxonomy, term}` appears in `_index_terms_starred.json` even if no
  document currently carries that term

#### Scenario: Starred taxonomy uses the config-filename name

- **WHEN** `L1-CONF-TAX-tag.yml` has `starred: true`
- **THEN** `_index_taxonomies_starred.json` contains `tag`
