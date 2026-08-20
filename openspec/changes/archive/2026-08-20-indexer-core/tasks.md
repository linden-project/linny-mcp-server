## 1. Front-matter parsing

- [x] 1.1 Split `---` front matter from body; parse YAML into a map (yaml.v3)
- [x] 1.2 Record model: filename, title, props (lowercased keys), body, task counts
- [x] 1.3 Report + skip malformed front matter (never crash)

## 2. Taxonomy graph

- [x] 2.1 Derive membership from taxonomy front-matter keys (scalar + list)
- [x] 2.2 Normalize term keys (lower, spaces->dashes) for on-disk names
- [x] 2.3 Load lindenConfig L1/L2 YAML; collect starred taxonomies/terms

## 3. JSON emission (per index-format spec)

- [x] 3.1 Home-level: taxonomies, docs_starred, docs_with_props, docs_with_title, docs_tasks_count, indexer_info, taxonomies_starred, terms_starred
- [x] 3.2 Nested L1 `<tax>/index.json` and L2 `<tax>/<term>/index.json`
- [x] 3.3 Atomic-ish write of the index tree under the index root

## 4. Safety

- [x] 4.1 Conflict-marker scan reported via a build report struct
- [x] 4.2 `lindexer build` wired with --corpus / --index flags

## 5. Tests

- [x] 5.1 Build against the synthetic corpus; assert every file exists + parses
- [x] 5.2 Shape tests: props excludes title-less; tasks counts; multi-term membership
- [x] 5.3 lindenConfig-driven L1 values + starred indexes
- [x] 5.4 Malformed record skipped; conflict marker reported
- [x] 5.5 Idempotent rebuild (set-equal JSON)
- [x] 5.6 Update nix vendorHash; nix flake check green
