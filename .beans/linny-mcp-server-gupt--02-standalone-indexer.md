---
# linny-mcp-server-gupt
title: 02 Standalone Indexer
status: completed
type: milestone
priority: normal
created_at: 2026-08-20T16:59:42Z
updated_at: 2026-08-20T21:17:12Z
---

A standalone indexer package + CLI (cmd/lindexer) that parses YAML front matter, builds the taxonomy graph, persists to SQLite+FTS5, emits linny.vim-compatible JSON, and can 'verify' its output against Hugo. Intended to eventually replace the Hugo indexer (Carl).
