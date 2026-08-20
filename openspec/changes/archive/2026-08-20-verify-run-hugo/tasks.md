## 1. Vendored reference + builder

- [x] 1.1 Vendor layouts + config under internal/hugoref/site (go:embed)
- [x] 1.2 hugoref.BuildReference(corpus): assemble temp site (embedded + corpus copy), run hugo, return index dir
- [x] 1.3 Clear error when hugo binary is absent

## 2. Differ options

- [x] 2.1 VerifyDirsWithOpts(ours, ref, {IgnoreReferenceOnly})
- [x] 2.2 Skip reference-only + unparseable files; normalize draft/iscjklanguage in props

## 3. CLI

- [x] 3.1 lindexer verify --hugo (build reference via hugo instead of --reference)

## 4. Tests & gate

- [x] 4.1 (skip if no hugo) verify --hugo: load-bearing files (taxonomies + L2 memberships) match
- [x] 4.2 subset ignores per-page files; props built-ins normalized
- [x] 4.3 nix flake check green (coverage >= 70%)
