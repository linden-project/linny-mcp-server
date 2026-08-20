## 1. Scope parsing

- [x] 1.1 `Parse(scopes)` → ScopeSet; actions read/write/delete + deny effect
- [x] 1.2 Selectors: `*`, `taxonomy:<tax>`, `taxonomy:<tax>:<term>`, `inbox`
- [x] 1.3 Reject unknown scopes with an error

## 2. SQL compilation

- [x] 2.1 `ReadableFilenamesSQL()` → subquery + args (allow OR-set) AND NOT (deny OR-set)
- [x] 2.2 Deny-by-default: no read rule => selects nothing
- [x] 2.3 Deny EXISTS across all memberships (cross-term)

## 3. Scoped store queries

- [x] 3.1 `SearchScoped`, `DocsByTermScoped`, `ListTaxonomiesScoped`
- [x] 3.2 `GetDocScoped` returns not-found for denied docs (no existence leak)

## 4. Tests

- [x] 4.1 Parse: valid vocabulary + reject unknown
- [x] 4.2 Deny-by-default reads nothing
- [x] 4.3 work+health excluded when health denied; work-only visible (SQL)
- [x] 4.4 GetDocScoped denied == missing
- [x] 4.5 read:taxonomy:<tax> limits to that taxonomy's docs
- [x] 4.6 nix flake check green
