## Why

After the L1 term-config lookup was moved onto the singular taxonomy name (matching
Hugo), the `_index_taxonomies_starred.json` / `_index_terms_starred.json` indexes were
still derived from the *plural* taxonomy list — an internal inconsistency (spec §13-Q7,
remaining half). This unifies everything on the singular convention and, as a bonus,
makes the L1 config resolve for the singular≠plural taxonomies (no more `{}`), while
keeping zero drift versus Hugo.

## What Changes

- **Canonical config filenames = singular taxonomy name.** The synthetic generator now
  writes `L1-CONF-TAX-<singular>.yml` and `L2-CONF-TAX-<singular>-TRM-<term>.yml`
  (`tag`/`project` instead of `tags`/`projects`), the key Hugo's `.Data.Singular`
  lookup resolves. The Hugo config's `singular: plural` map is derived from the same
  taxonomy table so all three (Hugo config, filenames, indexer) agree.
- **Config loading is filename-keyed.** `loadLindenConfig` scans `L1-CONF-TAX-*.yml`
  and `L2-CONF-TAX-*-TRM-*.yml`, keying by the taxonomy name in the filename — exactly
  Hugo's `.Site.Data` scan.
- **Starred indexes are config-derived.** `taxonomies_starred` and `terms_starred` are
  now built from the config maps (filename-keyed → singular), not the plural taxonomy
  list, and the non-standard "term must occur" filter is dropped (Hugo has none) — so
  we match Hugo on any notebook.
- **Result:** the L1 config now resolves (rich `title`/`views`/… for every taxonomy,
  not `{}`), the starred indexes use the singular convention consistently, and
  `verify --hugo` stays zero-drift.

## Capabilities

### Modified Capabilities
- `indexer`: config is filename-keyed; starred indexes are config-derived (singular),
  with no occurrence filter.
- `synthetic-corpus`: config files use the singular taxonomy name.

## Impact

- Modified: `internal/corpus/{corpus,config}.go`, `internal/index/{config,build}.go`,
  one corpus test, and the spec. No new dependency. Membership, search, and the
  document indexes are unchanged; zero Hugo drift preserved.
