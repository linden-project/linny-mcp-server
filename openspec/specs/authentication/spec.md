# authentication Specification

## Purpose
TBD - created by archiving change server-skeleton-auth. Update Purpose after archive.
## Requirements
### Requirement: Authenticator interface with static tokens

Authentication SHALL sit behind an `Authenticator` interface so an OIDC
implementation can be added later without changing callers.
`StaticTokenAuthenticator` SHALL be the only implementation for v1.

#### Scenario: Valid token authenticates

- **WHEN** a request carries `Authorization: Bearer <valid-token>`
- **THEN** the authenticator resolves it to that token's identity and scopes

### Requirement: Constant-time comparison

Token verification SHALL use `crypto/subtle.ConstantTimeCompare` (never `==`), so a
wrong token cannot be discovered via timing.

#### Scenario: Comparison is constant-time

- **WHEN** an invalid token of the same length as a valid one is presented
- **THEN** verification fails without an early-exit byte comparison

### Requirement: Information-free failure

The HTTP layer SHALL respond to every authentication failure identically: `401`
with an **empty body** and no detail. Missing header, malformed header, unknown
token, and wrong token MUST be indistinguishable to the caller.

#### Scenario: Missing and wrong tokens are indistinguishable

- **WHEN** one request sends no `Authorization` header and another sends a wrong
  bearer token
- **THEN** both receive `401` with an empty body and identical headers

### Requirement: Tokens stored as hashes with metadata

The token file SHALL store one record per token as JSON with `name`, `hash`
(SHA-256 of the token, hex), and `scopes`. Raw token values SHALL NOT be stored.

#### Scenario: File never contains raw tokens

- **WHEN** a token file produced by `gen-token` is read
- **THEN** it contains a `hash` field and no field equal to the raw token

### Requirement: gen-token produces a strong token

`linny-mcp gen-token` SHALL generate at least 32 bytes of CSPRNG entropy encoded as
base64url, print the raw token exactly once, and print the record line (name, hash,
scopes) to add to the token file.

#### Scenario: Token strength

- **WHEN** `gen-token` runs
- **THEN** it emits a base64url token decoding to ≥ 32 bytes and a matching hash record

### Requirement: Refuse unsafe bind

The server SHALL refuse to start bound to a non-loopback, non-mesh address unless an
explicit override flag is provided. Loopback and Nebula-mesh ranges are allowed by
default.

#### Scenario: Public bind refused

- **WHEN** `serve` is asked to bind a public address without the override
- **THEN** it exits non-zero with an error and does not listen

#### Scenario: Loopback allowed

- **WHEN** `serve` binds `127.0.0.1`
- **THEN** it starts without the override

### Requirement: Health endpoint is unauthenticated

`/healthz` SHALL be reachable without authentication and report liveness plus
placeholders for degraded-mode and sync status (filled by the git-safety change).

#### Scenario: Healthz open

- **WHEN** `/healthz` is requested with no `Authorization` header
- **THEN** it returns `200` with a JSON body containing a `status` field

