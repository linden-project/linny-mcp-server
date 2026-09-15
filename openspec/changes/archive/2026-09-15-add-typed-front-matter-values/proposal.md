## Why

`set_front_matter` cannot write a list. `scalarNode()` handles `bool` and `string`
and sends everything else through `fmt.Sprintf("%v", …)` tagged `!!str`, so setting a
taxonomy key to two terms does this:

```
set_front_matter("acme_note.md", "customer", ["acme", "globex"])
        │
        ▼   scalarNode default branch
customer: '[acme globex]'                      ← a single YAML string
        │
        ▼   index.termsOf sees a string, normalizeTerm lowercases + dashes spaces
membership: customer → "[acme-globex]"         ← a junk term, now real
```

This is not a no-op that an agent can notice and route around. The junk term is
written to the corpus, indexed into the SQLite membership table, and emitted into the
`_index_*.json` files that `linny.vim` consumes — so a mis-typed write pollutes the
taxonomy for every reader of the notebook, not just the agent that made it.

Numbers are affected the same way: a JSON number arrives as `float64` and is written
as a quoted string, so `priority: 3` becomes `priority: '3'`.

The practical consequence is that **an agent cannot classify a document at all.**
Membership in a Linny notebook *is* the document's meaning; in a corpus where a note
typically carries several terms across several taxonomies, the single-scalar
limitation rules out the most common write there is. The tool's own schema
description ("scalar value (string, bool, or number)") currently advertises the
limitation, so a well-behaved agent will not even attempt it.

`docs/future.md` already records this as "surgical, `fred`-style front-matter editing
for all value types"; the write-tools epic shipped with it as a known follow-up.

## What Changes

- **Type-preserving writes.** `set_front_matter` SHALL map the caller's JSON type to
  the corresponding YAML node type — string, bool, integer, float, null, and
  **sequences of scalars** — and SHALL never render a value through `fmt.Sprintf`.
  Types are taken from the caller, never guessed from a string's shape, so a string
  that looks like a date stays a quoted string.
- **Sequences are first-class.** Setting a taxonomy key to a list produces a YAML
  block sequence, and the returned membership reflects every element.
- **Add `add_term` and `remove_term`** — the operations a notebook actually performs.
  Both are idempotent, both take `(slug, taxonomy, term)`:
  - `add_term` on an absent key creates a sequence; on an existing scalar it promotes
    the scalar to a sequence and appends; on an existing sequence it appends unless
    the term is already present.
  - `remove_term` drops the term; removing the last remaining term removes the key
    rather than leaving an empty sequence behind.
- **Refuse writes that would be silently inert.** `index.termsOf` only reads strings
  and sequences of strings. Setting a *declared taxonomy key* to a number, a bool, or
  a nested structure produces no membership at all, so such a write SHALL be refused
  with a message naming the taxonomy — rather than accepted and quietly ignored by
  the indexer.
- **Report coined terms.** A term with no `L2-CONF-TAX-<tax>-TRM-<term>.yml` is still
  allowed — terms in a Linny notebook are open-ended — but the result SHALL list it
  under `new_terms`, so coining a term is visible rather than accidental.
- Update the `value` schema description, which currently tells agents that only
  scalars are accepted.

## Capabilities

### Modified Capabilities
- `mcp-write-tools`: typed front-matter values including sequences, `add_term` /
  `remove_term`, refusal of inert taxonomy writes, and `new_terms` reporting.

## Impact

- Modified: `internal/mcp/write.go` — `scalarNode` becomes a general
  `valueNode(any) (*yaml.Node, error)`; two new handlers sharing the existing
  `editFM` pipeline; `writeOut` gains `new_terms`. `docs/tools.md` and
  `docs/future.md` updated. No new dependencies.
- The notebook's declared taxonomy set is already loaded by
  `index.loadNotebook`, and declared terms by `index.loadLindenConfig`; both need
  exposing to the write path. This is the `lindenConfig` validation the write-tools
  change deferred.
- Order-preserving surgical editing is unchanged: values become richer `yaml.Node`
  trees, and untouched keys keep their nodes, so comments and key order survive.
- `writeOut` gains a field only — additive, existing clients unaffected.

## Non-Goals

- Nested mapping values in front matter. Linny front matter is flat: scalars and
  sequences of scalars. Nested maps stay unsupported and are refused, not stringified.
- Renaming or merging a term across the whole corpus (a bulk operation, and a very
  different safety problem).
- Validating that a *taxonomy* key is declared. An undeclared key is an ordinary
  front-matter property, which is legitimate.
