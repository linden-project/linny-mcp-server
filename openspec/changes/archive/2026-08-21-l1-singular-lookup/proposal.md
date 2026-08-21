## Why

`verify --hugo` left one class of drift: the L1 term-config index (`<tax>/index.json`)
differed for the singular≠plural taxonomies (`tags`, `projects`). Root cause found in
the reference layout `list.json.json`:

```
{{- $taxname := $.Data.Singular -}}
{{- $conf := printf "L2-CONF-TAX-%s-TRM-%s" $taxname $term -}}
```

Hugo builds the L2-config key from the **singular** taxonomy name, while our indexer
used the taxonomy name as it appears in membership (plural). So Hugo emits `{}` for
`tags`/`projects` (its lookup misses the plural-named config files) while ours embedded
the config. Our richer output was actually the divergence from what Hugo — and thus
`linny.vim` — sees in production. To be a byte-faithful drop-in, we must reproduce
Hugo's singular-keyed lookup.

## What Changes

- `loadNotebook` now returns the plural→singular taxonomy map (from the Hugo
  `taxonomies:` config; identity when a taxonomy has no declared singular), stored on
  the graph as `Singular`.
- The L1 `<tax>/index.json` emit looks up `L2-CONF-TAX-<singular>-TRM-<term>` — exactly
  like Hugo. For `customer`/`type`/`subject` (singular==plural) it still finds the
  config; for `tags`/`projects` it yields `{}`, matching the reference on any notebook.
- Spec updated: §9.1 documents the singular-keyed L1 lookup; §13-Q7 is resolved.
- The Hugo round-trip test now asserts **zero** discrepancies.

## Capabilities

### Modified Capabilities
- `indexer`: L1 term-config index is keyed by the singular taxonomy name.
- `verify-hugo`: the round-trip is now zero-drift on a well-formed corpus.

## Impact

- Modified: `internal/index/{config,build,model,emit}.go`, the round-trip test, and
  `docs/linden-index-spec.md`. No behavioural change for membership, search, or the
  other index files. No new dependency.
- Note: Hugo aborts on a corpus containing malformed front matter, so `verify --hugo`
  requires a Hugo-buildable corpus; our own indexer still degrades gracefully.
