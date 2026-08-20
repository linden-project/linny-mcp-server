## 1. Config model

- [x] 1.1 `Notebook{Name, CorpusPath, StateDir}` and `Config{ListenAddress, Port, TokensFile, LogLevel, ReadOnly, PublicHostname, Notebooks}`
- [x] 1.2 `Validate()`: >=1 notebook, unique non-empty names, required paths
- [x] 1.3 `Notebook(name)` lookup; `DefaultNotebook()`

## 2. Loading

- [x] 2.1 `Load(path)` parses JSON and validates
- [x] 2.2 `FromFlags(...)` builds a single-notebook `default` config

## 3. serve wiring

- [x] 3.1 `--config` and `--notebook` flags; resolve config (file or flag sugar)
- [x] 3.2 Build a git-safety Guard for the selected notebook

## 4. NixOS module

- [x] 4.1 Add `publicHostname` option (no default host)
- [x] 4.2 Add `notebooks` list option (name/corpusPath/stateDir); desugar `corpusPath`

## 5. Tests

- [x] 5.1 Validation: unique/non-empty names, >=1 notebook, missing paths
- [x] 5.2 Load round-trip from JSON; flag-sugar single notebook
- [x] 5.3 Notebook selection by name + default-first
- [x] 5.4 No hardcoded hostname (unset validates)
- [x] 5.5 nix flake check green
