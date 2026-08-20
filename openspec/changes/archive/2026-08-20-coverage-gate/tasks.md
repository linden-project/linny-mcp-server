## 1. Coverage headroom

- [x] 1.1 Add CLI e2e tests (cmd/lindexer build+search; cmd/linny-mcp serve validation)

## 2. Gate

- [x] 2.1 Add checks.<system>.coverage: full suite with -coverpkg=./..., fail < 70%
- [x] 2.2 Make git available in the coverage check so git-backed tests run
- [x] 2.3 Keep the coverage profile as a build output

## 3. Docs

- [x] 3.1 Document the testing agreement + 70% gate in openspec/project.md

## 4. Verify

- [x] 4.1 nix flake check runs gotest + lint + coverage, all pass
- [x] 4.2 Record measured total coverage in the bean
