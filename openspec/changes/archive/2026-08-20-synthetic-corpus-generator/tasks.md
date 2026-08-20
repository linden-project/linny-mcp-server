## 1. Generator package

- [x] 1.1 `internal/corpus`: `Options` (seed, count, enableEdgeCases, taxonomies)
- [x] 1.2 Deterministic PRNG seeded from Options; fixed word lists
- [x] 1.3 `Generate(dir, Options)` writes content records + lindenConfig + Hugo config.yaml
- [x] 1.4 Slug convention for filenames; front matter with title/crdate/starred/taxonomies/tasks

## 2. Edge cases

- [x] 2.1 unicode, long front matter, empty body, malformed YAML
- [x] 2.2 committed conflict markers
- [x] 2.3 fake secrets (AWS-key-shaped, private-key header)

## 3. CLI

- [x] 3.1 `cmd/gen-corpus` materializes into a target dir (default testdata-gen/)

## 4. Tests

- [x] 4.1 Determinism test (same seed → identical bytes)
- [x] 4.2 Structure test (flat, parseable, config consistent)
- [x] 4.3 Edge-case presence tests (conflict marker + fake secret present)
