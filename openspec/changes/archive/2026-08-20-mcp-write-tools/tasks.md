## 1. Write scope predicates

- [x] 1.1 authz CanWriteAll / CanWriteInbox

## 2. Write pipeline (internal/mcp/write.go)

- [x] 2.1 writer bound to store/scope/guard/audit/policy/corpus
- [x] 2.2 scope check; readable==loadable (denied==not-found)
- [x] 2.3 guard.EnsureWritable degraded gate
- [x] 2.4 atomic + optimistic write (create must-not-exist; edits by hash)
- [x] 2.5 quarantine-by-default on create
- [x] 2.6 reindex + return resulting membership (store.TermsOfDoc)
- [x] 2.7 audit every attempt (ok/denied/error)

## 3. Tools

- [x] 3.1 create_doc (slug convention, quarantine, render)
- [x] 3.2 append_to_doc
- [x] 3.3 set_front_matter / unset_front_matter (surgical yaml.Node, order-preserving)
- [x] 3.4 archive (sets archived: true)
- [x] 3.5 register only when writes enabled; serve opens audit log

## 4. Tests & gate

- [x] 4.1 create quarantined + membership + file + audit ok
- [x] 4.2 forbidden without write scope; modify forbidden with read-only
- [x] 4.3 refused when degraded
- [x] 4.4 set_front_matter updates membership + file
- [x] 4.5 archive sets flag
- [x] 4.6 protocol-level e2e create_doc then read-back
- [x] 4.7 docs/tools.md updated; nix flake check green (coverage >= 70%)
