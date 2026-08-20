---
# linny-mcp-server-r7gl
title: Front-matter parser & taxonomy graph
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:29:01Z
parent: linny-mcp-server-gupt
blocked_by:
    - linny-mcp-server-0gdv
    - linny-mcp-server-v1py
---

Standalone cmd/lindexer package. Parse YAML front matter directly, honour slug convention, build the taxonomy graph (terms, membership, co-occurrence, related). Scan for committed conflict markers and report loudly.

**OpenSpec change:** `indexer-frontmatter-taxonomy`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/indexer-frontmatter-taxonomy/tasks.md`. Ships with tests._

## Summary of Changes

Delivered the standalone indexer core in internal/index (wired to `lindexer build --corpus --index`). Parses YAML front matter directly (yaml.v3), builds the taxonomy graph (scalar + list membership), loads lindenConfig L1/L2, and emits every index-format file: the 8 home-level _index_*.json plus nested L1 <tax>/index.json (term -> L2 config) and L2 <tax>/<term>/index.json (member filenames). Title-gating, task counts, and starred sets match the spec. Malformed front matter is reported and skipped (never crashes); committed conflict markers are detected and reported (feeds git-safety degraded mode). Idempotent rebuilds. Tests cover front-matter parsing, task counting, conflict detection, all home files, multi-term membership (work_and_health in both tags/work and tags/health), L1 config embedding, malformed-skip, and rebuild determinism.

nix vendorHash updated for the new yaml.v3 dep; nix flake check (gotest + lint) green offline.
