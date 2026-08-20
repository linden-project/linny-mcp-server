## 1. Detectors

- [x] 1.1 PEM private-key block (multiline, whole-block redaction)
- [x] 1.2 AWS access-key ID (AKIA/ASIA); AWS secret-access-key assignment
- [x] 1.3 GitHub token (ghp_/gho_/...); Slack token; JWT
- [x] 1.4 Generic `api_key|secret|token|password|access_key` assignment (redact value, keep key)
- [x] 1.5 IBAN-shaped account number

## 2. Redactor API

- [x] 2.1 `Redact(text) (string, int)` — sequential detectors, typed placeholders, count
- [x] 2.2 `RedactValue(any) any` — deep-walk strings in maps/slices/strings
- [x] 2.3 Group-aware replacement (redact only the value in assignments)

## 3. Tests

- [x] 3.1 Each detector: secret removed, placeholder present
- [x] 3.2 Assignment keeps key, redacts value
- [x] 3.3 Deep-walk of nested map/slice
- [x] 3.4 End-to-end: index the synthetic fake-secrets note, GetDoc, redact, assert no secret remains
- [x] 3.5 No false-positive on ordinary prose; count reported
- [x] 3.6 nix flake check green
