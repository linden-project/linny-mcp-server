# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/), and this project aims to follow
semantic versioning once it leaves alpha.

## [Unreleased]

### Added
- Standalone indexer (`lindexer`): `build`, `search` (SQLite + FTS5), `verify`
  (diff our JSON against a reference or a live Hugo build), and `watch` (debounced
  fsnotify rebuilds).
- MCP server (`linny-mcp serve`) over the official MCP Go SDK (streamable-HTTP at
  `/mcp`) behind static bearer auth, with `/healthz`.
- MCP tools: read/navigate (`search`, `get_doc`, `list_taxonomies`, `terms`,
  `docs_by_term`), history (`history`, `diff`, `changed_since`), write
  (`create_doc`, `append_to_doc`, `set_front_matter`, `unset_front_matter`,
  `archive`), and operational (`sync_status`, `verify_index`).
- Deny-by-default authorization scopes enforced in SQL, gitleaks-style egress
  redaction, quarantine-on-create with an append-only audit log, and data-delimited
  document bodies.
- Git-safety: degraded read-only mode on a conflicted/mid-operation tree, atomic
  writes, and optimistic concurrency; `sync_status` + ntfy degraded alerts.
- `linny-mcp gen-token`, `backup`/`restore`, and an `--no-quarantine` opt-out.
- Nix flake (plain nix, no flake-utils) with a hardened NixOS module, N-notebook
  config, a configurable public hostname, and a gated `nix flake check`
  (tests + lint + coverage, including a zero-drift Hugo round-trip).
- Docs: Linden Index Specification v0.3.0, `docs/tools.md`, `docs/future.md`,
  `docs/deploy-mipnix.md`, `docs/local-testing.md`.

### Changed
- Index L1 term-config lookup and the starred indexes are keyed by the singular
  taxonomy name, matching the Hugo reference (zero drift).

[Unreleased]: https://github.com/linden-project/linny-mcp-server/commits/main
