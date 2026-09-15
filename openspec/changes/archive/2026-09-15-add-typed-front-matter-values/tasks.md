## 1. Value typing

- [x] 1.1 replace `scalarNode` with `valueNode(any) (*yaml.Node, error)`
- [x] 1.2 map string / bool / int / float / null per the design's type ladder
- [x] 1.3 integral float64 within int64 range → `!!int`, otherwise `!!float`
- [x] 1.4 sequences of scalars → block sequence node
- [x] 1.5 refuse mappings and non-scalar list elements
- [x] 1.6 update the `value` jsonschema description (it currently says scalars only)

## 2. Taxonomy awareness in the write path

- [x] 2.1 expose the notebook's declared taxonomy set and declared terms from
      `internal/index` to the writer
- [x] 2.2 refuse a non-string, non-string-sequence value on a declared taxonomy key,
      naming the taxonomy in the message
- [x] 2.3 `writeOut` gains `new_terms`; populate it for undeclared terms

## 3. add_term / remove_term

- [x] 3.1 `add_term`: absent → sequence; scalar → promote then append; sequence →
      append unless already a member
- [x] 3.2 membership comparison via the indexer's `normalizeTerm`, writing the
      caller's spelling
- [x] 3.3 `remove_term`: drop the term; removing the last one removes the key
- [x] 3.4 both idempotent; both ride the existing `editFM` pipeline

## 3b. Degraded gate (defect found while implementing)

`editFM` and `append_to_doc` never called `guard.EnsureWritable()`, so four of the
five shipped write tools wrote happily while the tree was conflicted or mid-rebase —
against the existing `mcp-write-tools` requirement "Writes are refused while
degraded". `add_term` / `remove_term` would have inherited it.

- [x] 3b.1 gate `editFM` on `guard.EnsureWritable()`
- [x] 3b.2 gate `append_to_doc` on `guard.EnsureWritable()`
- [x] 3b.3 test every write tool refuses while degraded

## 4. Tests

- [x] 4.1 type round-trips: int, float, bool, null, string; `"2026-09-14"` stays a
      string
- [x] 4.2 sequence write → membership includes every element; no `'[…]'` in the file
- [x] 4.3 nested mapping refused; list-of-lists refused
- [x] 4.4 number on a declared taxonomy key refused; number on an ordinary key accepted
- [x] 4.5 `add_term` across all four starting shapes (absent / scalar / sequence /
      already present)
- [x] 4.6 case- and space-insensitive duplicate detection (`"Acme"` vs `acme`)
- [x] 4.7 `remove_term` removes the key on the last term; absent term is a no-op
- [x] 4.8 `new_terms` populated for an undeclared term, empty for a declared one
- [x] 4.9 key order + comments preserved; body untouched
- [x] 4.10 protocol-level e2e: `add_term` then `get_doc` shows the term in `props`

## 5. Docs & gate

- [x] 5.1 `docs/tools.md`: typed values, the type ladder, `add_term` / `remove_term`
- [x] 5.2 `docs/future.md`: drop the "fred-style editing for all value types" entry
- [x] 5.3 `nix flake check` green (gotest + lint + coverage >= 70%)
