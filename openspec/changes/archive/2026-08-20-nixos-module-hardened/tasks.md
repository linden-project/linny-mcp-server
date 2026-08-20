## 1. Hardening

- [x] 1.1 Add the full systemd hardening directive set to the unit
- [x] 1.2 ReadWritePaths = each notebook corpus + state dir (already computed)

## 2. ntfy option

- [x] 2.1 Add ntfyTopicURL option; include in generated config JSON

## 3. Verify

- [x] 3.1 NixOS eval asserts the hardening keys + ntfy URL on the unit
- [x] 3.2 nix flake check green (module still evaluates)
