## 1. Single sources of truth

- [x] 1.1 extract the writes-enabled condition in `server.go` into one predicate used
      by both write-tool registration and the report
- [x] 1.2 thread the caller's identity and raw scope strings to the read-side
- [x] 1.5 add `ScopeSet.CanReadAny`, mirroring the write predicates
- [x] 1.3 make the notebook name and quarantine policy reachable from the handler
- [x] 1.4 reuse the existing guard-backed accessor for degraded/read-only state

## 2. Tool (internal/mcp/tools.go)

- [x] 2.1 result types for server, notebook, caller and status
- [x] 2.2 modify permission as `all` / `own-drafts` / `none` from `CanWriteAll` and
      `CanWriteInbox`; no delete field
- [x] 2.3 `can_write_now` as the conjunction of registration, guard and scope
- [x] 2.4 `reason` naming the first failing gate, in registration/guard/scope order
- [x] 2.5 register `session_info` with a description saying it answers what the caller
      may do

## 3. Tests

- [x] 3.1 server version, notebook name, identity and raw scopes reported
- [x] 3.2 modify permission across all three scope shapes
- [x] 3.3 no delete field; no counts, taxonomy names, slugs or paths in the response
- [x] 3.4 `can_write_now` false on a narrow scope, with a scope reason
- [x] 3.5 `can_write_now` false on a degraded tree, with a degraded reason
- [x] 3.6 `can_write_now` true when all three gates pass
- [x] 3.7 `sync_status` and `session_info` agree on a degraded tree
- [x] 3.8 a server with write tools unregistered reports writes as not enabled
- [x] 3.9 protocol-level e2e: `session_info` over MCP for a `write:inbox` token
      reports `own-drafts`, `can_create: true`, and `can_write_now: true` with a reason
      naming `write:*` as what modifying needs (creating is a write, so the verdict is
      true; the reason carries the limit)

## 4. Docs & gate

- [x] 4.1 `docs/tools.md`: add `session_info` to the operational table, document the
      three-state modify permission and the `can_write_now` verdict
- [ ] 4.2 `nix flake check` green (build + tests + lint + coverage floors)
