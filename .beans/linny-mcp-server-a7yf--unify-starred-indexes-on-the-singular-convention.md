---
# linny-mcp-server-a7yf
title: Unify starred indexes on the singular convention
status: completed
type: epic
priority: normal
created_at: 2026-08-21T08:05:25Z
updated_at: 2026-08-21T08:08:48Z
parent: linny-mcp-server-gupt
---

Canonicalize lindenConfig filenames on the singular taxonomy name and derive the starred indexes (taxonomies_starred, terms_starred) from the config files (filename-keyed, like Hugo) rather than the plural taxonomy list. Unifies with the L1 singular lookup, makes the L1 config resolve (no more {}), keeps zero Hugo drift. **OpenSpec change:** `starred-singular-unify`

## Summary of Changes

Unified all config-derived identifiers on the singular taxonomy convention, resolving the remaining half of spec 13-Q7. (1) The synthetic generator now names lindenConfig files by the SINGULAR taxonomy name (L1-CONF-TAX-tag.yml, L2-CONF-TAX-tag-TRM-<term>.yml) and derives the Hugo config singular:plural map from the same taxonomy table. (2) loadLindenConfig scans and keys L1/L2 config by the filename taxonomy name (like Hugo .Site.Data). (3) finalize derives _index_taxonomies_starred and _index_terms_starred from the config maps (filename-keyed -> singular), dropping the non-standard occurrence filter (Hugo has none) so we match Hugo on any notebook.

Effects: the L1 term-config index now RESOLVES for every taxonomy (rich title/views/... instead of {} for tag/project), the starred indexes use the singular convention consistently with the L1 lookup, and verify --hugo stays ZERO-drift (both sides filename-derived on singular). Updated the corpus config-consistency test and spec 5.2/9.1/13-Q7 (Q7 fully resolved; canonical convention = singular config filenames). 15/15 packages pass; coverage 76.7 percent; nix flake check (incl. the hugo round-trip) all green.
