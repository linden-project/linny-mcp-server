# Linden Index Specification

- **Version:** 0.3.0
- **Status:** Draft (alpha). Normative for the `linny-mcp` PoC indexer.
- **Date:** 2026-08-20
- **Editor of record:** Pim Snel (Linden Project)
- **Supersedes (in spirit):** Linden Specification 0.2.0 — see §11 and §12.

> This document describes the **JSON index format** of a Linny notebook: the set of
> files a Linden *indexer* produces from a markdown corpus, and that Linden *clients*
> (`linny.vim`, `linnydroid`) consume. It is written so a competent implementer can
> build a conforming indexer **without reading the Hugo templates**. Where the current
> Hugo indexer's behaviour is ambiguous or apparently accidental, it is flagged in
> §13 (Open Questions) rather than silently normalised.

---

## 1. Purpose and scope

A Linny notebook is a private, flat directory of markdown *records*. All structure —
projects, customers, subjects, tags, dates, flags — lives in each record's YAML
**front matter**, not in the filesystem. The taxonomy graph is *derived output*: an
**indexer** reads the records plus a small configuration set and emits a tree of JSON
**index files**. Clients read those index files to build navigation, dashboards, and
lookups; they never re-parse the whole corpus.

This specification defines:

- the corpus and configuration inputs an indexer reads (§4–§5);
- the location, layout, and naming of the index it writes (§6–§7);
- the exact shape and semantics of every index file (§8–§9);
- which client feature consumes which file (§10);
- the legacy flat layout and how it maps to the current one (§11);
- conformance and the `verify` diff against the reference indexer (§12);
- known ambiguities as open questions (§13).

**Out of scope:** WikiTags (`[[...]]`), embeddings, and any co-occurrence / related /
"due this week" queries. Those are client- or server-side features computed *from* the
index or the corpus, not index files. See §11.3.

**Reference implementation.** The current producer is a Hugo site ("Carl", in
`linden-project/linny-notebook-template`). Hugo's output is the *reference*: a
conforming indexer's output must be close enough that `linny.vim` works against it
unmodified, and a `verify` tool (§12) diffs the two. Where this document and Hugo
disagree, that is a defect in one of them and must be recorded in §13.

---

## 2. Terminology

| Term | Meaning |
|------------------|--------------------------------------------------------------------|
| Record / document| One markdown file in the content directory. Its **filename** (basename incl. `.md`) is its identity key across all indexes. |
| Front matter     | The YAML block at the top of a record declaring its properties and taxonomy memberships. |
| Taxonomy         | A named grouping dimension (e.g. `tags`, `projects`, `customer`, `subject`, `type`). Declared by the notebook, referenced by front-matter keys. |
| Term             | One value within a taxonomy (taxonomy `customer` → term `eric`). |
| Membership       | The set of documents that carry a given term. |
| Index root       | The directory the indexer writes to (Hugo `publishDir`, default `lindenIndex`). |
| lindenConfig     | The directory of per-taxonomy / per-term YAML metadata files (§5). |
| Client           | Any Linden-compliant reader (`linny.vim`, `linnydroid`). |
| Indexer          | The component that reads records + lindenConfig and emits the index. |
| Starred          | A boolean favourite flag, applicable to a document, a taxonomy, or a term. |
| L1 / L2          | Level-1 = per-taxonomy index/config; Level-2 = per-term index/config. |

---

## 3. Conventions

- Key words **SHALL**, **MUST**, **SHOULD**, **MAY** are used in the RFC 2119 sense.
- JSON is UTF-8. Pretty-printing/whitespace is not significant; clients parse with a
  standard JSON decoder.
- Arrays in this format are **unordered**: consumers MUST NOT depend on element order,
  and `verify` (§12) compares arrays as sets.
- All filename keys are the record basename **including** the `.md` extension
  (e.g. `my_note.md`).

---

## 4. The corpus model

### 4.1 Content directory

- A single **flat** directory of `.md` records. No subdirectories carry meaning.
- Created from `linden-project/linny-notebook-template`.

### 4.2 Slug / filename convention

New documents are named by slugifying their title (as `linny.vim`'s
`word_to_filename` does, `doc/linny.txt`):

1. trim surrounding whitespace;
2. replace spaces with the space-replacement character (default `_`);
3. replace `/` and `:` with the same character;
4. lowercase;
5. append `.md`.

So `LinnyNewDoc My new Document` → `my_new_document.md`. The filename is the record's
stable identity; **titles may change, filenames should not.**

### 4.3 Front matter

YAML at the top of each record. Observed keys in the reference notebook:

| Key       | Type    | Role |
|-----------|---------|------|
| `title`   | string  | Display title. **Gates** inclusion in the props/title indexes (§8.3, §8.4). |
| `crdate`  | string  | Creation date, quoted `"YYYY-MM-DD"`. A plain string, **not** a Hugo `date`. |
| `starred` | boolean | Marks the document as a favourite (§8.2). |
| taxonomy keys | scalar or list | Any key that names a taxonomy (e.g. `tags`, `customer`, `project`, `type`, `subject`) declares membership. A scalar is one term; a list is many. Term text is lowercased when used as an index key. |

New documents created through a client are pre-populated with `title`,
`crdate: <today>`, the current `taxonomy: term`, and any `frontmatter_template`
values from the term's L2 config (§5.2).

An indexer **SHALL** parse YAML front matter directly (it does not depend on Hugo at
runtime) and **SHALL** scan record content for committed git conflict markers
(`<<<<<<<`, `=======`, `>>>>>>>`) and report them loudly (§12.3).

---

## 5. Configuration: `lindenConfig`

The `lindenConfig/` directory (Hugo `dataDir`; resolved by `linny.vim` as
`<notebook>/lindenConfig`) holds per-taxonomy and per-term **presentation and
metadata**. The **list of taxonomies itself** is a notebook-level setting (in the Hugo
template it lives in Hugo's own `config.yaml` `taxonomies:` map, e.g.
`tag → tags`, `project → projects`, `customer → customer`, `type → type`,
`subject → subject`). A standalone indexer SHOULD read the taxonomy list from an
explicit notebook config; see §13-Q6.

### 5.1 L1 — taxonomy config: `L1-CONF-TAX-<taxonomy>.yml`

Keys: `title`, `infotext`, `starred` (bool), `plural`, `description`, and `views`
(a map of named views → `{ group_by, sort, sort_key, only, except }`). A taxonomy
with `starred: true` appears in `_index_taxonomies_starred.json` (§8.7).

### 5.2 L2 — term config: `L2-CONF-TAX-<taxonomy>-TRM-<term>.yml`

Keys: `title`, `infotext`, `archive` (bool), `starred` (bool), `views`, and
(currently commented in the template) `mounts`, `locations`, `frontmatter_template`.
**This entire object is what is embedded, verbatim, as the value for the term in the
L1 index** (§9.1). A term with `starred: true` appears in
`_index_terms_starred.json` (§8.8).

> **Filename-derived identifiers (caution).** The reference indexer parses the
> taxonomy/term *out of the config filename* for the starred indexes, replacing spaces
> with dashes. Keep config filenames aligned with front-matter term slugs, or the
> emitted identifiers will mismatch (§13-Q7).

### 5.3 `lindenConfig/views/*.yml`

`all.yml`, `custom.yml`, `root.yml` configure the client's dashboard widgets
(`starred_documents`, `starred_terms`, `starred_taxonomies`, `all_taxonomies`,
`recently_modified_documents`, `menu`). They are client configuration, **not** index
files, and are out of scope for the indexer.

---

## 6. Index location and overall layout

The indexer writes to a single **index root** (default `lindenIndex`, resolved by
clients as `<notebook>/lindenIndex`). It contains two kinds of file:

```
lindenIndex/
├── _index_taxonomies.json            # home-level (flat)
├── _index_taxonomies_starred.json
├── _index_terms_starred.json
├── _index_docs_starred.json
├── _index_docs_with_props.json
├── _index_docs_with_title.json       # DEPRECATED (§8.4)
├── _index_docs_tasks_count.json
├── _indexer_info.json
├── <taxonomy>/index.json             # L1: per-taxonomy (nested)
└── <taxonomy>/<term>/index.json      # L2: per-term (nested)
```

The index root is a **disposable cache**. It is git-ignored and never committed;
deleting it and rebuilding is always a valid recovery step and MUST be documented as
such.

---

## 7. Naming and normalization

- **Taxonomy directory segment:** lowercased plural taxonomy key.
- **Term directory segment:** lowercased, with spaces replaced by dashes
  (`Learn Linny` → `learn-linny`).
- **Home-level files:** the eight fixed `_index_*.json` basenames listed in §6/§8.
- **Document keys:** record basename including `.md`.

(These normalizations match `linny.vim`'s `paths.lua`: taxonomy `string.lower`, term
`lower` + spaces→dashes.)

---

## 8. Home-level index files

### 8.1 `_index_taxonomies.json`

- **Shape:** JSON array of strings.
- **Semantics:** every taxonomy key present in the notebook (plural form).
- **Example:** `["customer", "projects", "subject", "tags", "type"]`
- **Consumer:** taxonomy autocomplete in front matter; "set/remove taxonomy" popups;
  the All-Taxonomies dashboard widget.

### 8.2 `_index_docs_starred.json`

- **Shape:** JSON array of document filenames.
- **Semantics:** every record whose front matter `starred` is `true`.
- **Example:** `["pinned.md", "todo_today.md"]`
- **Consumer:** Starred-Documents dashboard widget.

### 8.3 `_index_docs_with_props.json`

- **Shape:** JSON object, `filename → { …full front matter… }` (keys lowercased).
- **Semantics:** the primary document catalogue. **Only records that have a `title`
  key are included** (§13-Q2).
- **Example:**
  ```json
  {
    "first_note.md": { "title": "First Note", "tags": "note", "customer": "eric",
                        "starred": true, "crdate": "2021-03-04" }
  }
  ```
- **Consumer:** term (L2) menu document listing, view filtering/grouping, export/zip.

### 8.4 `_index_docs_with_title.json` — DEPRECATED

- **Shape:** JSON object, `filename → title` (string).
- **Status:** **Deprecated.** Titles are available from `_index_docs_with_props.json`.
  Still emitted and still read by current `linny.vim` for title resolution; new
  clients SHOULD read titles from the props index and treat this file as legacy.
- **Example:** `{ "first_note.md": "First Note" }`

### 8.5 `_index_docs_tasks_count.json`

- **Shape:** JSON object, `filename → { "open": int, "closed": int, "total": int }`.
- **Semantics:** Markdown task-list counts per record: `- [ ]` → open, `- [x]` →
  closed, `total = open + closed`. **Only records with `total > 0` are included.**
- **Example:** `{ "sprint.md": { "open": 2, "closed": 1, "total": 3 } }`
- **Consumer:** task-count badges on documents in the term menu.
- **Ambiguity:** line-start anchoring and uppercase `[X]` are unspecified (§13-Q5).

### 8.6 `_indexer_info.json`

- **Shape:** JSON object of strings.
- **Semantics:** identifies the indexer and the resolved directories.
- **Fields:** `product_name`, `product_version`, plus a version field for the engine
  (`hugo_version` in the reference), and `index_dir`, `content_dir`, `config_dir`.
- **Note:** the Hugo reference emits literal `"TODO"` for the three directory fields
  and `product_name: "hugo-lindex"`. A conforming standalone indexer **SHALL**
  populate the directory fields with real paths and set its own `product_name` /
  `product_version` (e.g. `lindexer`). `verify` (§12) tolerates the Hugo placeholder.
- **Consumer:** none currently reads it (informational).

### 8.7 `_index_taxonomies_starred.json`

- **Shape:** JSON array of taxonomy name strings.
- **Semantics:** taxonomies whose `L1-CONF-TAX-*` config has `starred: true`.
- **Consumer:** Starred-Taxonomies dashboard widget.

### 8.8 `_index_terms_starred.json`

- **Shape:** JSON array of objects `{ "taxonomy": string, "term": string }`.
- **Semantics:** terms whose `L2-CONF-TAX-*-TRM-*` config has `starred: true`.
- **Example:** `[ { "taxonomy": "subject", "term": "learn-linny" } ]`
- **Consumer:** Starred-Terms dashboard widget.

---

## 9. Per-taxonomy and per-term index files

### 9.1 L1 — `<taxonomy>/index.json`

- **Shape:** JSON object, `term → term-config-object`.
- **Semantics:** the catalogue of terms that actually occur in the taxonomy. Each
  value is the term's L2 config object (§5.2) — `title`, `infotext`, `starred`,
  `views`, … Terms are keyed by their occurring value. A term with **no** resolvable
  L2 config maps to `{}`.
- **Config key uses the SINGULAR taxonomy name.** The reference builds the lookup
  as `L2-CONF-TAX-<singular>-TRM-<term>` (Hugo's `.Data.Singular`). So for a taxonomy
  whose singular differs from its plural (e.g. `tag`→`tags`, `project`→`projects`),
  L2-CONF files named with the *plural* do **not** resolve and every term maps to
  `{}` — this is the reference's actual behaviour, not a bug in the consumer. For
  taxonomies whose singular equals their plural (`customer`, `type`, `subject`) the
  config resolves normally. A conforming indexer MUST key this lookup by the singular
  name to stay byte-compatible with the reference.
- **Example** (`customer/index.json`):
  ```json
  {
    "eric": { "title": "Eric", "infotext": "About eric", "starred": false,
              "views": { "type": { "group_by": "type" } } },
    "note": {}
  }
  ```
- **Consumer:** term autocomplete; Level-1 (taxonomy) side menu; term
  listing/grouping via `views.*.group_by`.

### 9.2 L2 — `<taxonomy>/<term>/index.json`

- **Shape:** JSON array of document filenames.
- **Semantics:** the membership list — every record carrying that (taxonomy, term).
- **Example** (`customer/eric/index.json`): `["first_note.md", "address_eric.md"]`
- **Consumer:** Level-2 (term) side menu — the list of member documents.

---

## 10. Consumer map (index file → client feature)

| Index file | Client feature |
|----------------------------------|-------------------------------------------------|
| `_index_taxonomies.json`         | Taxonomy autocomplete; set/remove-taxonomy popups; All-Taxonomies widget |
| `_index_taxonomies_starred.json` | Starred-Taxonomies widget |
| `_index_terms_starred.json`      | Starred-Terms widget |
| `_index_docs_starred.json`       | Starred-Documents widget |
| `_index_docs_with_props.json`    | Term menu doc listing; view filtering/grouping; export |
| `_index_docs_with_title.json`    | Title resolution (deprecated, still read) |
| `_index_docs_tasks_count.json`   | Task-count badges |
| `_indexer_info.json`             | (none — informational) |
| `<tax>/index.json` (L1)          | Term autocomplete; Level-1 side menu; term grouping |
| `<tax>/<term>/index.json` (L2)   | Level-2 side menu — member documents |

---

## 11. Relationship to prior art

### 11.1 Legacy flat layout (Lindex / Linden Spec 0.2.0)

The predecessor indexer **lindex** and the written Linden Specification 0.2.0 used a
**flat** per-taxonomy/per-term layout:

| Current (nested, normative) | Legacy (flat, 0.2.0) |
|------------------------------------|-------------------------------------------|
| `<tax>/index.json`                 | `L1-INDEX-TAX-<tax>.json`                 |
| `<tax>/<term>/index.json`          | `L2-INDEX-TAX-<tax>-TRM-<term>.json`      |

The home-level `_index_*.json` files are largely shared. `verify` (§12) MAY offer a
`--legacy-flat` mode to compare against a flat producer.

### 11.2 Behavioural differences in lindex (not adopted in v0.3.0)

- lindex **synthesizes** a title from the filename for records lacking `title`, and
  tags every record into a synthetic `front_matter: valid|invalid` taxonomy. The Hugo
  reference does neither. v0.3.0 follows Hugo (title-less records are dropped from the
  props/title indexes); revisiting this is §13-Q2.
- lindex additionally wrote a human-facing `index.md` (A–Z `[[wikilink]]` list) into
  the wiki directory. That is not an index file and is out of scope.

### 11.3 Correction: `$FORMAT` / `$INCLUDE` do not exist

Earlier briefing material referred to `$FORMAT` / `$INCLUDE` front-matter directives.
**No such directives exist** anywhere in the notebook template, `linny.vim`, or the
spec repo. What exists are **WikiTags** — `[[LIN …]]`, `[[DIR …]]`, `[[VIM …]]`,
`[[SHELL …]]`, `[[GH …]]`, `[[FILE …]]` — which are *editor navigation actions*
handled by the client, unrelated to the JSON index. An indexer MUST NOT attempt to
expand them.

---

## 12. Conformance and verification

### 12.1 Conformance

A conforming indexer:

1. writes every home-level file in §8 and the nested L1/L2 files in §9 to the index
   root (§6), using the naming/normalization of §7;
2. produces JSON of the exact shapes in §8–§9, such that `linny.vim` operates against
   it unmodified;
3. treats the index root as a disposable cache (rebuild-from-empty always valid);
4. reports committed conflict markers (§12.3).

### 12.2 `verify` against the reference

The indexer ships a `verify` subcommand that runs the Hugo reference indexer, runs the
standalone indexer over the same corpus, and diffs the JSON outputs, reporting any
discrepancy. Rules:

- object indexes are compared key-by-key; array indexes are compared as **sets**
  (order-insensitive, §3);
- `_indexer_info.json` is compared **excluding** the fields the reference leaves as
  `"TODO"` and the `product_*` identity fields;
- any non-empty diff is a reportable drift, not a silent pass.

### 12.3 Conflict markers

During indexing, tracked content is scanned for lines beginning `<<<<<<<`,
`=======`, or `>>>>>>>`. On any hit the indexer reports the file and marker
prominently and (per the server's git-safety rules) the corpus is treated as
degraded.

---

## 13. Open questions

These are surfaced deliberately; they are **not** decided in v0.3.0.

- **Q1 — Nested vs. flat filenames.** Nested (`<tax>/index.json`) is normative because
  the live client reads it. Confirm this is permanent, or require emitting both the
  nested and legacy-flat names during a migration window.
- **Q2 — Title-less / invalid-front-matter records.** Hugo silently drops them from
  the props/title indexes; lindex synthesizes a title and tags
  `front_matter: valid|invalid`. Which behaviour is canonical for v0.4?
- **Q3 — Vestigial per-page JSON.** The Hugo template also emits a per-record
  `<slug>/index.json` (`{ data: { title, date, type, permalink, summary } }`) that no
  client reads. Define it or drop it.
- **Q4 — Multi-valued taxonomy fields.** Samples only exercise scalar term values.
  Confirm membership semantics for list-valued fields (lindex distinguished
  `has_many` vs `has_many_belong_to_many`).
- **Q5 — Task-count regex.** Must a task marker be at line start? Is uppercase `[X]`
  a closed task? The Hugo and lindex regexes differ subtly.
- **Q6 — Source of the taxonomy list.** Hugo takes it from its own site config; a
  standalone indexer needs an explicit notebook-level taxonomy declaration (the old
  `L0-CONF-ROOT.yml` was removed in spec 0.2.0). Define where the taxonomy list lives.
- **Q7 — Filename-derived term identifiers. (RESOLVED for the L1 lookup.)** The L1
  term-config index is keyed by the **singular** taxonomy name
  (`L2-CONF-TAX-<singular>-TRM-<term>`), matching the reference (§9.1); `linny-mcp`
  reproduces this exactly, so plural-named L2-CONF files under a singular≠plural
  taxonomy resolve to `{}` just as they do under Hugo. The remaining open part: the
  *starred* indexes still derive taxonomy/term names by splitting the config filename
  (plural) and replacing spaces with dashes — a separate identifier source from the
  L1 lookup. Whether to unify these (and to canonicalize L2-CONF filenames on the
  singular name so configs always resolve) is deferred.
- **Q8 — `_indexer_info.json` field set.** Standardize the field set (the reference is
  unstable: literal `"TODO"` paths, engine-specific version keys).

---

## 14. Version history

| Version | Notes |
|---------|-------|
| 0.1.1   | Initial draft (website). |
| 0.1.2   | All index/config file names and options enumerated (website). |
| 0.2.0   | Removed L0 configuration; added `starred` to L1 config (website). Flat index layout. |
| (0.2.1) | Lindex code only: added `_index_docs_tasks_count`. Never written into the spec. |
| —       | 2021 Hugo/"Carl" port redesigned index files (nested layout); never folded back into the spec (per "Project Reboot", 2021-04-29). |
| **0.3.0** | **This document.** Folds the Hugo redesign back into the spec: nested L1/L2 layout normative, home-level files documented as emitted today, deprecations and ambiguities recorded, `$FORMAT`/`$INCLUDE` correction. |

---

## 15. References

- `linden-project/linny-notebook-template` — Hugo layouts (reference producer).
- `linden-project/linny.vim` (== `mipmip/linny.vim`) — reference consumer
  (`autoload/linny.vim`, `lua/linny/**`, `doc/linny.txt`).
- `linden-project/lindex` — archived Crystal predecessor indexer.
- `linden-project/linden-project.github.io` — Linden Specification 0.1.2 / 0.2.0 and
  the "Project Reboot" post.
