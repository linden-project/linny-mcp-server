## 1. Differ

- [x] 1.1 VerifyDirs(ours, reference) -> []Discrepancy (missing files + content diffs)
- [x] 1.2 Arrays compared as sets; objects key-by-key; scalars canonical
- [x] 1.3 _indexer_info: ignore product_name/version/hugo_version + "TODO" values

## 2. CLI

- [x] 2.1 lindexer verify --corpus --reference: build ours, diff, exit non-zero on drift

## 3. Tests & gate

- [x] 3.1 identical trees -> no discrepancy; reordered array -> equal
- [x] 3.2 changed membership + missing file -> reported
- [x] 3.3 indexer_info identity/TODO ignored
- [x] 3.4 nix flake check green (coverage >= 70%)
