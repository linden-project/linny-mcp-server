## 1. Audit log

- [x] 1.1 internal/audit: Entry{time, identity, tool, slug, diff, outcome}
- [x] 1.2 Open(path) append-only (O_APPEND|O_CREATE); Append; Close; injectable clock
- [x] 1.3 Not redacted (operator-facing, outside corpus)

## 2. Quarantine policy

- [x] 2.1 internal/defense Policy (taxonomy=status, term=agent-draft)
- [x] 2.2 ApplyQuarantine / IsQuarantined (scalar + list merge)
- [x] 2.3 RequiresConfirmation(delete, bulk_retag)

## 3. Data delimiters

- [x] 3.1 Delimit(body): strip forged markers, wrap in begin/end
- [x] 3.2 Apply to get_doc body (after redaction)

## 4. Scope invariant

- [x] 4.1 Document/assert scope is fixed per request (no widening API)

## 5. Tests & gate

- [x] 5.1 Audit append-only (order + immutability); path under stateDir
- [x] 5.2 Quarantine apply/detect (scalar + list); confirmation flags
- [x] 5.3 Delimit wraps + neutralizes forged marker
- [x] 5.4 get_doc returns a delimited body
- [x] 5.5 nix flake check green (coverage >= 70%)
