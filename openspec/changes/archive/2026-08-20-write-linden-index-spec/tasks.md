## 1. Draft the specification

- [x] 1.1 Write front matter/header: title, version v0.3.0, status, scope, audience
- [x] 1.2 Document the corpus model: flat content dir, slug convention, front-matter keys
- [x] 1.3 Document `lindenConfig` (L1-CONF-TAX-*, L2-CONF-TAX-*-TRM-*) as metadata source
- [x] 1.4 Document the index root + flat/nested layout and normalization rules

## 2. Document every index file

- [x] 2.1 Home-level files: taxonomies, docs_starred, docs_with_props, docs_with_title (deprecated), docs_tasks_count, indexer_info, taxonomies_starred, terms_starred
- [x] 2.2 Per-taxonomy L1 `<tax>/index.json` and per-term L2 `<tax>/<term>/index.json`
- [x] 2.3 For each: filename, JSON shape, field semantics, and the linny.vim consumer
- [x] 2.4 Consumer map table (index file -> linny.vim feature)

## 3. Reconciliation & corrections

- [x] 3.1 Record the lindex flat-filename legacy layout + migration note
- [x] 3.2 State that `$FORMAT`/`$INCLUDE` do not exist; WikiTags are an editor concern
- [x] 3.3 Mark deprecated/vestigial files (docs_with_title, per-page json)

## 4. Open questions

- [x] 4.1 List every ambiguity/accidental behaviour as explicit open questions
- [x] 4.2 Add the spec-version rationale and the divergences vs written spec 0.2.0

## 5. Verification

- [x] 5.1 Cross-check every claim against the cloned reference repos
- [x] 5.2 Acceptance self-check: could an implementer build a conforming indexer from this doc alone?
