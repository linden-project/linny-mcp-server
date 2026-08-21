---
# linny-mcp-server-f2wc
title: 'Zero drift: L1 singular-key lookup'
status: completed
type: epic
priority: normal
created_at: 2026-08-21T07:42:37Z
updated_at: 2026-08-21T07:47:16Z
parent: linny-mcp-server-gupt
---

Reproduce Hugo's L1 term-config lookup exactly (keyed by the SINGULAR taxonomy name, .Data.Singular) so lindexer verify --hugo reports ZERO drift on a well-formed corpus. Root cause: Hugo builds L2-CONF-TAX-<singular>-TRM-<term>; ours used the plural. **OpenSpec change:** `l1-singular-lookup`

## Summary of Changes

Achieved ZERO drift vs the Hugo reference. Root cause (found in the vendored list.json.json): Hugo builds the L1 term-config key from the SINGULAR taxonomy name (.Data.Singular) — L2-CONF-TAX-<singular>-TRM-<term> — while our indexer used the plural. So Hugo emits {} for tag/project (its lookup misses the plural-named config files) while ours embedded the config; our richer output was the divergence. Fix: loadNotebook now returns the plural->singular map (from the Hugo taxonomies config; identity otherwise), stored on Graph.Singular; the L1 emit looks up L2-CONF by the singular name, exactly like Hugo. customer/type/subject (singular==plural) still resolve; tags/projects now emit {} matching the reference on any notebook.

Spec updated: section 9.1 documents the singular-keyed L1 lookup; section 13 Q7 marked resolved for the L1 lookup. The Hugo round-trip test now asserts ZERO discrepancies and runs inside nix flake check (coverage check has hugo). Verified: verify --hugo on a well-formed synthetic corpus reports "no discrepancies". Coverage 76.8 percent; all checks passed. Note: Hugo aborts on malformed front matter, so verify --hugo needs a Hugo-buildable corpus; our indexer still degrades gracefully.
