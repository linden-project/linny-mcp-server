---
# linny-mcp-server-ljtf
title: Egress secret redaction
status: completed
type: epic
priority: normal
created_at: 2026-08-20T17:00:28Z
updated_at: 2026-08-20T19:37:01Z
parent: linny-mcp-server-fo8d
blocked_by:
    - linny-mcp-server-gvc5
    - linny-mcp-server-sac5
---

gitleaks-style redaction filter on ALL tool responses so no response can return a credential regardless of what an agent asks for. High priority. Exercised by fake secrets in the synthetic corpus.

**OpenSpec change:** `egress-redaction`

_This epic is administered via its OpenSpec change; tasks live in `openspec/changes/egress-redaction/tasks.md`. Ships with tests._

## Summary of Changes

Added internal/redact: a gitleaks-style, rule-based Redactor. Detectors (in order): PEM private-key blocks, AWS access-key IDs (AKIA/ASIA), AWS secret-access-key assignments, GitHub tokens, Slack tokens, JWTs, IBAN-shaped account numbers, and a generic api_key|secret|token|password|access_key assignment rule (redacts the value, keeps the key for context). Each match becomes a typed [REDACTED:<kind>] placeholder. Redact(text) returns the scrubbed text + a redaction count (never the secret); RedactValue(any) deep-walks strings in maps/slices so every field of a structured response is scrubbed. The generic rule is guarded so it never re-redacts a placeholder a more specific detector produced.

Verified: per-detector tests (secret removed, placeholder present), assignment key-kept, deep-walk of nested map/slice, no false positive on ordinary prose, count reporting, and an end-to-end test that indexes the synthetic fake_secrets.md, reads it back via the SQLite store, and asserts none of the planted AWS key / secret / PEM / token / IBAN survive. nix flake check: all checks passed. Wiring point: E0501 read tools pipe every response through this redactor before serialization.
