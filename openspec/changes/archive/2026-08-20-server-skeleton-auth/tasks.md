## 1. Auth package

- [ ] 1.1 `Authenticator` interface + `Identity{Name, Scopes}`
- [ ] 1.2 Token file format (JSON records: name, hash, scopes); loader
- [ ] 1.3 `StaticTokenAuthenticator` with `subtle.ConstantTimeCompare` over SHA-256 hashes
- [ ] 1.4 Bearer header parsing (missing/malformed both fail closed)

## 2. gen-token

- [ ] 2.1 `linny-mcp gen-token --name --scopes`: CSPRNG >=32B base64url, print token once + record line

## 3. HTTP skeleton

- [ ] 3.1 Bearer middleware: 401 empty body, no timing signal, on any failure
- [ ] 3.2 `/healthz` (unauth) with status/degraded/sync placeholders
- [ ] 3.3 `/mcp` placeholder behind auth (real MCP SDK wiring deferred)

## 4. Bind safety

- [ ] 4.1 `config` helper: allow loopback + Nebula mesh; refuse public without override
- [ ] 4.2 `linny-mcp serve` flags: --corpus --state-dir --listen --port --tokens-file --log-level --read-only --i-know-what-im-doing

## 5. Tests

- [ ] 5.1 Valid token authenticates; identity + scopes resolved
- [ ] 5.2 Missing vs wrong token: both 401, empty body, identical response
- [ ] 5.3 Token file stores hashes, never raw tokens; gen-token round-trips
- [ ] 5.4 Bind safety: public refused, loopback + mesh allowed
- [ ] 5.5 /healthz reachable unauthenticated
- [ ] 5.6 nix flake check green
