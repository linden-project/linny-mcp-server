## Why

The DoD requires that the server "starts from the NixOS module, authenticates a
bearer token, serves the read tools." Authentication is the gate everything else sits
behind, and the briefing (§5) is prescriptive about how it must work — constant-time
comparison, information-free 401s, CSPRNG tokens, hashed storage, and refusal to bind
to a public address. This change delivers that gate and the server skeleton it lives
in. The official MCP SDK tool-wiring is a deliberate follow-up (noted below); this
change stops at an authenticated HTTP skeleton with health reporting.

## What Changes

- Add `internal/auth`: an `Authenticator` interface with `StaticTokenAuthenticator`
  as the only implementation. Tokens are compared with
  `crypto/subtle.ConstantTimeCompare`; the token file stores SHA-256 **hashes**, not
  raw tokens, with per-token `name` and `scopes` metadata.
- Add `linny-mcp gen-token`: generate a ≥32-byte CSPRNG token (base64url) and print
  both the secret (once) and the record line (hash + name + scopes) to add to the
  token file.
- Add an HTTP bearer middleware: on any auth failure return `401` with an **empty
  body**, no detail, and no timing signal (constant-time path).
- Add `internal/config` bind-safety: the server refuses to start bound to a
  non-loopback / non-mesh address unless an explicit override flag is set.
- Wire `linny-mcp serve`: start an HTTP server exposing an unauthenticated
  `/healthz` (liveness + degraded/sync fields placeholder) and an authenticated
  `/mcp` placeholder. Actual MCP protocol + tools land in the next change.

## Capabilities

### New Capabilities
- `authentication`: static bearer-token authentication behind an `Authenticator`
  interface, with hashed storage, constant-time comparison, and information-free
  failures.
- `server-runtime`: the `serve` skeleton — bind safety, `/healthz`, and the
  authenticated request path.

### Modified Capabilities

## Impact

- New: `internal/auth/**` (replaces the doc.go stub), `internal/config` bind safety,
  `cmd/linny-mcp` `serve` + `gen-token`.
- No new external dependency (standard library only).
- Follow-up (documented, not in this change): official MCP Go SDK wiring and the read
  tools registered on `/mcp`.
