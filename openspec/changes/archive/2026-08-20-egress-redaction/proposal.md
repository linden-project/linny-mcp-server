## Why

A second brain spanning years almost certainly contains passwords, API keys, and
account numbers. Per the briefing (§7.3, "high priority"), **no tool response may ever
return a credential**, regardless of what an agent asks for — the difference between a
leaked note and a leaked bank account. This is defense-in-depth applied at the egress
boundary: even a correctly-scoped, non-conflicted read of a legitimate note must have
its secrets stripped before the bytes leave the server.

## What Changes

- Add `internal/redact`: a gitleaks-style, rule-based `Redactor` that finds and
  replaces common secret shapes (PEM private-key blocks, AWS access-key IDs, AWS
  secret-access-key assignments, GitHub/Slack tokens, JWTs, generic
  `key|secret|token|password` assignments, and IBAN-shaped account numbers) with a
  typed `[REDACTED:<kind>]` placeholder.
- `Redact(text)` returns the redacted text plus a count of redactions; `RedactValue`
  deep-walks arbitrary response values (strings, maps, slices) so every string in a
  structured tool response is scrubbed — the read-tool epic pipes **every** response
  through it.
- Detectors redact only the secret value (keeping surrounding context like the key
  name) where the shape is an assignment, and the whole match otherwise.

The wiring point is defined here and consumed by the MCP read tools (E0501): all tool
output flows through `redact.Redactor` before serialization.

## Capabilities

### New Capabilities
- `egress-redaction`: rule-based redaction of credentials from all outgoing tool
  responses.

### Modified Capabilities

## Impact

- New: `internal/redact/**` (replaces the doc.go stub). Standard library only.
- Verified end-to-end against the synthetic corpus's fake-secret fixtures via the
  indexer store (index → GetDoc → redact → assert no secret remains).
- Consumed by E0501 (read tools). Best-effort by design: it reduces blast radius; it
  is not a substitute for keeping secrets out of the corpus.
