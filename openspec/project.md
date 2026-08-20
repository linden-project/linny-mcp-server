# Project: linny-mcp

## Purpose

A single Go binary exposing a private markdown corpus (a Hugo/front-matter based
"Linny notebook" — thousands of flat `.md` files) to AI agents over the Model
Context Protocol (MCP). This repository is the PoC / alpha base; it is consumed
downstream as a Nix flake input in `github.com/mipmip/mipnix`.

Core properties:
- Its **own** disposable index (SQLite + FTS5), never committed to git.
- A **standalone indexer** package/CLI intended to eventually replace the current
  Hugo-based indexer ("Carl").
- A written **Linden Index Specification** produced *before* the indexer code.
- **Static bearer token** authentication (explicitly *not* OAuth).
- **Read-only degradation** whenever the git working tree is conflicted.
- A Nix flake exporting a package, a NixOS module with a hardened systemd unit,
  and `checks`.

## Tech Stack

- **Language:** Go (target toolchain: 1.26). Official MCP Go SDK.
- **SQLite driver:** `modernc.org/sqlite` (pure Go, no cgo). NOT `mattn/go-sqlite3`.
  Keeps cross-compilation and `buildGoModule` painless.
- **Index storage:** SQLite + FTS5 in `stateDir`. Disposable cache, never in git.
- **File watching:** `fsnotify` for incremental index updates.
- **Packaging:** Nix flake with `packages`, `nixosModules`, `overlays`,
  `devShells`, `checks`. Plain nix for supported systems — **do NOT use
  flake-utils**. Enumerate `x86_64-linux` and `aarch64-linux` explicitly.
- **Version control:** jj (Jujutsu), colocated with git. Remote:
  `git@github.com:linden-project/linny-mcp-server.git`.

## Project Conventions

### Code Style
- Standard Go formatting (`gofmt`/`goimports`), vetted by `golangci-lint`.
- The indexer is its own package with its own CLI entrypoint (`cmd/lindexer`),
  importable by the server but runnable alone.
- Authentication is behind an `Authenticator` interface; `StaticTokenAuthenticator`
  is the only implementation for v1. An `OIDCAuthenticator` may be added later
  without rebuilding anything — do NOT build it now.

### Architecture
- `cmd/linny-mcp` — the MCP server binary (+ `gen-token` helper).
- `cmd/lindexer` — the standalone indexer CLI (`build`, `watch`, `verify`).
- `internal/index` — front-matter parsing, taxonomy graph, SQLite/FTS5, JSON emit.
- `internal/auth` — Authenticator interface + static bearer tokens.
- `internal/authz` — scopes, deny-by-default, SQL-level filtering.
- `internal/gitsafe` — working-tree inspection, degraded read-only mode, atomic writes.
- `internal/redact` — gitleaks-style egress secret redaction on all tool responses.
- `internal/mcp` — MCP tool surface wiring.
- `nix/` — flake modules (package, nixosModule, overlay, devShell, checks).
- `docs/` — `linden-index-spec.md`, `tools.md`, `future.md`.
- `testdata/synthetic-corpus/` — generated synthetic notebook for tests.

### Testing Strategy
- Thorough unit + e2e tests are non-negotiable. Every OpenSpec change ships tests.
- **Generate a synthetic corpus** (a few thousand realistic notes). NEVER point any
  build, test, or tool at the real private `secondbrain` repo.
- Required test classes: JSON-index round-trip vs Hugo; auth timing-safety +
  no-information-leak 401; scope intersection (the `work`+`health` deny case);
  degraded-mode write refusal on a conflicted tree; concurrency (two writers, one
  stale hash); a tested backup/restore path.
- `nix flake check` runs `go test` + lint and must pass — it gates this build.

### Git Workflow
- jj (colocated with git), committing as **Pim Snel** — no self-promoting trailers,
  no Co-authored-by, no "Generated with" attribution.
- **Commit after every archival of an OpenSpec change.** One change → one shippable,
  tested increment → archive → commit.
- Bean file(s) are included in the same commit as the code they track.

## Security & Secret Hygiene (load-bearing)
- **No token value may ever appear in a Nix option** — Nix options land
  world-readable in `/nix/store`. The NixOS module takes a `tokensFile` *path*,
  sourced from `age.secrets` with `owner` set to the service user.
- Compare bearer tokens with `crypto/subtle.ConstantTimeCompare`, never `==`.
- On auth failure: `401`, empty body, no detail, no timing signal.
- The corpus is **hostile input** (prompt injection). Agent writes land in a
  quarantine taxonomy (`inbox`/`agent-draft`) by default. Deny-by-default scopes.
  Egress secret redaction on all responses. Append-only audit log kept OUTSIDE the
  corpus.
- The server must refuse to start bound to a non-loopback / non-mesh address
  without an explicit override flag. TLS is terminated upstream (Hetzner proxy).

## Decisions already made — do not relitigate
Go; `modernc.org/sqlite`; SQLite+FTS5 disposable index; server builds its own index
(Hugo JSON output is the *reference* implementation, diffed via `verify_index`);
static bearer tokens (no OAuth/OIDC/Zitadel); existing external git-sync stays and
the server never takes ownership of git; embeddings (if ever) local-only and out of
scope for v1; plain-nix flake (no flake-utils).

## Open questions (surface, don't guess)
1. Exact public hostname (`secondbrain.pimsnel.com`? — treat as config, not constant).
2. Multi-notebook support (one corpus or several?). Recommend designing config for
   N notebooks even if v1 serves one.
3. Spec version number (`linden-index-spec v0.3`?).
4. Final MCP tool names (tool names are an API — version them in `docs/tools.md`).
5. Which git-sync script (affects whether we read its state or inspect the tree).
