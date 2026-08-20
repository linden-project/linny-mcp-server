---
# linny-mcp-server-0gdv
title: Project scaffolding, Nix flake & devShell
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T17:09:49Z
parent: linny-mcp-server-tq36
---

Repo layout, Go module, plain-nix flake (packages/overlays/devShells for x86_64+aarch64-linux, NO flake-utils), devShell (Go, golangci-lint, hugo), README with the secret-hygiene rule stated prominently.

**OpenSpec change:** `scaffold-repo-and-flake`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/scaffold-repo-and-flake/tasks.md`. Ships with tests._

## Summary of Changes

Scaffolded the repo: Go module (github.com/linden-project/linny-mcp-server, go 1.24 for nixpkgs compat (target 1.26)), cmd/linny-mcp + cmd/lindexer with a `version` subcommand and smoke tests, internal package layout (index/auth/authz/gitsafe/redact/mcp/config) as doc.go stubs, and a plain-nix flake (NO flake-utils) with explicit systems x86_64-linux + aarch64-linux, buildGoModule package (vendorHash=null), overlay, devShell (go/golangci-lint/hugo), and checks.{gotest,lint}. README states the secret-hygiene rule prominently; .gitignore excludes the disposable index/state.

**Verification:** `go build/vet/test ./...` green; `nix flake check` => "all checks passed!" (aarch64-linux evaluated but omitted from build on this x86_64 host). OpenSpec change scaffold-repo-and-flake archived; specs project-scaffolding + nix-packaging created.
