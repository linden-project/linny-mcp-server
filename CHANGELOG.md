# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/), and this project aims to follow
semantic versioning once it leaves alpha.

## [Unreleased]

### Added
- **Standalone indexer** (`lindexer`): `build` (linny.vim-compatible JSON + a
  SQLite/FTS5 store), `search` (FTS5, bm25-ranked with snippets), `verify` (diff our
  JSON against a reference or a live Hugo build), and `watch` (debounced fsnotify
  rebuilds).
- **MCP server** (`linny-mcp serve`) over the official MCP Go SDK (streamable-HTTP at
  `/mcp`) behind static bearer auth, with an unauthenticated `/healthz`.
- **MCP tools** — read/navigate (`search`, `get_doc`, `list_taxonomies`, `terms`,
  `docs_by_term`), history (`history`, `diff`, `changed_since`), write (`create_doc`,
  `append_to_doc`, `set_front_matter`, `unset_front_matter`, `archive`), and operational
  (`sync_status`, `verify_index`).
- **Security model** — deny-by-default authorization scopes enforced in SQL (correct
  cross-term intersection), gitleaks-style egress redaction on every response,
  quarantine-on-create (with an `--no-quarantine` opt-out) and an append-only audit log
  kept outside the corpus, and data-delimited document bodies.
- **Git-safety** — degraded read-only mode on a conflicted/mid-operation tree, atomic
  writes, optimistic concurrency, and ntfy alerts on degraded transitions.
- **Operational commands** — `linny-mcp gen-token`, and `backup`/`restore` with a
  verified round-trip.
- **Packaging** — a plain-nix flake (no flake-utils; `x86_64`/`aarch64-linux`) with a
  hardened NixOS module, N-notebook config, a configurable public hostname, and a gated
  `nix flake check` (build + lint + coverage floors + a zero-drift Hugo round-trip).
- **Developer workflow** — a deterministic synthetic corpus generator (`gen-corpus`) for
  tests; `CHANGELOG.md`; `scripts/ship-change.sh` (gated per-change ship) and
  `scripts/release.sh` (VERSION-driven `gum`/tag/`gh` releases, where the first release
  cuts the current untagged version without a bump); `gum`/`gh` in the dev shell.
- **Docs** — the Linden Index Specification v0.3.0, `docs/tools.md`, `docs/future.md`,
  `docs/deploy-mipnix.md`, `docs/local-testing.md`, and `RELEASING.md`.
- **Branding** — a Linnaeus-mascot hero banner in the README (with the "Connect.
  Classify. Empower." tagline and a MIT license badge), and a `docs/brand/` reference
  recording the palette and taglines.

### Changed
- The version is sourced from a `VERSION` file (single source of truth), read by the
  flake and reported by the `version` subcommand, the `serve` banner, `_indexer_info`,
  and the MCP handshake.
- The L1 term-config lookup and the starred indexes are keyed by the singular taxonomy
  name (with lindenConfig files named by the singular taxonomy), matching the Hugo
  reference exactly — `verify --hugo` is zero-drift.

[Unreleased]: https://github.com/linden-project/linny-mcp-server/commits/main
