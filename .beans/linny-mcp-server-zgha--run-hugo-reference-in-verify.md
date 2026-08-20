---
# linny-mcp-server-zgha
title: Run Hugo reference in verify
status: completed
type: epic
priority: normal
created_at: 2026-08-20T21:23:48Z
updated_at: 2026-08-20T21:30:55Z
parent: linny-mcp-server-gupt
---

Make lindexer verify actually run the Hugo reference indexer (vendored layouts + reference config over the corpus) and diff our JSON against it — the drift safety net running end-to-end. Follows up the indexer-verify epic which diffs against a pre-built reference dir.

**OpenSpec change:** `verify-run-hugo`

## Summary of Changes

Made lindexer verify actually run the Hugo reference indexer. Vendored the reference Hugo site (layouts + config) from linny-notebook-template under internal/hugoref/site (go:embed). hugoref.BuildReference(corpus) assembles a temp Hugo site from the embedded layouts/config + a COPY of the corpus content/lindenConfig (the real corpus is never mutated), runs hugo, and returns the produced index dir; a clear error when hugo is absent. Extended the differ with VerifyDirsWithOpts + VerifyOpts{IgnoreReferenceOnly}: it compares only the files our indexer emits (skipping Hugo vestigial per-page <slug>/index.json, which are also invalid JSON) and normalizes away Hugo built-in params (draft, iscjklanguage) in docs_with_props. Wired lindexer verify --hugo.

Result on the synthetic corpus: every load-bearing file (taxonomies, all L2 memberships, task counts, props, starred sets, docs_with_title, terms/taxonomies_starred, indexer_info) MATCHES Hugo; the only reported drift is the L1 term-config index for the singular!=plural taxonomies (tag/project) where Hugo emits {} vs ours embeds the L2 config (spec 13 Q7) — exactly the drift the tool exists to surface. Verified by a real Hugo round-trip integration test (skips if hugo absent) that asserts load-bearing files match and only L1 files may differ; plus hugo-free unit tests for the subset/normalization options. Added pkgs.hugo to the coverage check so the round-trip runs inside nix flake check. Coverage 76.7 percent; all checks passed.
