## 1. Implementation

- [x] 1.1 defense.Policy.Disabled; ApplyQuarantine no-ops when disabled
- [x] 1.2 config.disableQuarantine + FromFlags param
- [x] 1.3 serve --no-quarantine flag + startup warning
- [x] 1.4 NixOS module `quarantine` option (default true) -> disableQuarantine

## 2. Tests & gate

- [x] 2.1 defense: disabled ApplyQuarantine no-op
- [x] 2.2 mcp: create_doc with disabled policy -> not quarantined
- [x] 2.3 nix flake check green (coverage >= 70%)
