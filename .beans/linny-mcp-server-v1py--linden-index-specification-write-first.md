---
# linny-mcp-server-v1py
title: Linden Index Specification (write FIRST)
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:15:52Z
parent: linny-mcp-server-tq36
---

docs/linden-index-spec.md v0.3. Document the Hugo JSON index format as emitted today; rewrite from the ground up using the stranded spec for vocabulary only. Acceptance: a competent implementer could write a conforming indexer from this doc alone. List ambiguities as open questions.

**OpenSpec change:** `write-linden-index-spec`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/write-linden-index-spec/tasks.md`. Ships with tests._

## Summary of Changes

Wrote docs/linden-index-spec.md, Linden Index Specification v0.3.0, documenting the Hugo index format as emitted today: index root (lindenIndex), naming/normalization, all eight home-level `_index_*.json` files and the nested L1 `<tax>/index.json` + L2 `<tax>/<term>/index.json` files, with exact JSON shapes, field semantics, and a consumer map to linny.vim. Reconciled vs the legacy flat Lindex/Spec-0.2.0 layout, corrected the $FORMAT/$INCLUDE misconception (they do not exist; WikiTags are an editor concern), flagged 8 open questions, and recorded the version rationale (0.2.0 -> Hugo redesign -> 0.3.0).

All claims cross-checked against the cloned reference repos (publishDir, the 8 baseNames, paths.lua normalization, absence of $FORMAT/$INCLUDE). OpenSpec capability index-format created and becomes the contract for milestone 02 and verify_index.
