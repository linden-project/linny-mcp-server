## Context

Two independent implementations of the "same" format exist and disagree, plus a
written spec that predates both on the relevant points:

- **Hugo / "Carl"** (`linny-notebook-template`) — the current producer. Emits nested
  `<tax>/index.json` + `<tax>/<term>/index.json` and eight flat home-level
  `_index_*.json` files. This is what `linny.vim` reads today (verified: the two
  `linny.vim` clones are byte-identical at commit `3b722a4`).
- **lindex** (archived Crystal) — the predecessor. Emits **flat**
  `L1-INDEX-TAX-<tax>.json` / `L2-INDEX-TAX-<tax>-TRM-<term>.json`, adds a synthetic
  `front_matter: valid|invalid` taxonomy, and synthesizes titles for title-less docs.
- **Written Linden Specification 0.1.2 / 0.2.0** (website) — documents the *flat*
  lindex-era filenames and says (Reboot post, 2021-04-29): "The Hugo port needed us to
  change some of the designs of the index files, so we will implement these in the
  specification." That folding-back never happened.

The spec we write is the reconciliation the Reboot post promised.

## Goals / Non-Goals

**Goals:**
- Describe the format **as Hugo emits it today**, because that is what the live
  consumer reads. Where Hugo and the written spec disagree, Hugo wins for v0.3.0 and
  the divergence is recorded.
- Be implementable from the document alone; be the diff target for `verify_index`.
- Surface — never silently normalise — every ambiguity.

**Non-Goals:**
- Not defining embeddings, co-occurrence, or related-docs indexes (they do not exist
  in the current format; co-occurrence/related are MCP-query features, not index
  files).
- Not specifying WikiTags (`[[LIN ...]]`, `[[DIR ...]]`, …) — an editor concern.
- Not designing a new format; this documents the existing contract.

## Decisions

- **Version = v0.3.0.** History: spec 0.1.2 → 0.2.0 (published), Lindex code reached
  0.2.1, then the 2021 Hugo redesign was never written down. The redesign is a
  backward-incompatible index-layer change → a MINOR bump in the 0.x line. 0.2.x is
  taken; v1.0.0 is reserved for a first stability commitment once the standalone
  indexer ("replace Carl") is real.
- **Nested filenames are normative; flat lindex names are documented as legacy.** The
  live consumer (`linny.vim` `paths.lua`) builds nested paths, so a byte-compatible
  standalone indexer MUST emit nested. The flat `L1-INDEX-*/L2-INDEX-*` names are
  recorded as the 0.2 legacy layout with a migration note.
- **`$FORMAT`/`$INCLUDE` do not exist — corrected in the spec.** Grep across all
  clones (source, docs, spec repo) finds no such directives. The corpus uses `[[...]]`
  WikiTags, handled by the editor, not the indexer. The spec states this explicitly so
  the indexer epics don't chase a phantom feature.
- **Title-less / invalid-front-matter handling is called out as a divergence.** Hugo
  silently drops title-less docs from the props/title indexes; lindex synthesizes a
  title and tags `front_matter: valid|invalid`. The spec documents Hugo's behaviour as
  normative for v0.3.0 and lists lindex's behaviour as an open question for v0.4.
- **`_index_docs_with_title.json` documented as DEPRECATED** but still emitted (still
  read by `linny.vim:318`); titles otherwise come from `_index_docs_with_props.json`.

## Risks / Trade-offs

- [Choosing Hugo-as-normative freezes some accidental behaviour] → every such point is
  listed as an open question, so v0.4 can deliberately change it rather than
  rediscover it.
- [`_indexer_info.json` path fields are literal "TODO" in Hugo] → spec requires a
  standalone indexer to populate them; `verify` must tolerate the Hugo placeholder
  when diffing.
- [Ordering of array indexes is unspecified in Hugo output] → spec states arrays are
  unordered and consumers must not rely on order; `verify` compares as sets.

## Open Questions

Carried into the spec's own Open Questions section (surface, don't guess):
1. Nested vs. flat filenames — confirm nested is the permanent choice, or provide both.
2. Title-less / invalid-front-matter docs — adopt lindex's synthesize-and-tag, or keep
   Hugo's silent drop?
3. Should the vestigial per-page `<slug>/index.json` output be defined or dropped?
4. Multi-valued (list) taxonomy fields — confirm membership semantics
   (`has_many` vs `has_many_belong_to_many`) since samples only exercise scalars.
5. Task-count regex: must a task be at line start; is `[X]` (uppercase) a closed task?
