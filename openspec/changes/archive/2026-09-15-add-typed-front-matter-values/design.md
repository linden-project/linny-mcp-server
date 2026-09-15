# Design: typed front-matter values

## The type ladder

JSON is the wire format, YAML is the storage format, and they do not agree about
numbers. The mapping has to be stated once and applied everywhere:

| Caller passes (JSON)     | YAML node written        | Note                       |
|--------------------------|--------------------------|----------------------------|
| `"acme"`                 | `!!str`                  | quoted only when needed    |
| `true` / `false`         | `!!bool`                 |                            |
| `3`                      | `!!int`                  | integral float64 → int     |
| `3.5`                    | `!!float`                |                            |
| `null`                   | `!!null`                 |                            |
| `["acme","globex"]`      | block sequence of `!!str`| the taxonomy case          |
| `{...}` / nested arrays  | *refused*                | not representable in Linny |

Two rules keep this honest:

**Types come from the caller, never from the text.** A JSON string `"2026-09-14"` is
written as a quoted string. Sniffing it into a YAML timestamp would mean an agent
setting a `title` of `"12:30"` silently produces a sexagesimal number — the classic
YAML footgun. If a caller wants a date type, that is a follow-up with an explicit
type field, not a heuristic.

**JSON's single number type is resolved by value.** A `float64` with no fractional
part and within int64 range is written as an int; otherwise as a float. This is the
only place a guess is made, and it is the guess every JSON-to-YAML bridge makes.

## Why refusing beats accepting

`index.termsOf` reads exactly two shapes:

```go
case string:  → one term
case []any:   → one term per string element
default:      → nil          ← no membership, no error, no trace
```

So `priority: 3` on a *declared taxonomy* key is not an error anywhere in the
pipeline — it simply produces nothing. The write succeeds, `finish()` reindexes, and
the returned membership is missing the term the agent thought it had just set. The
agent then has no way to distinguish "the write failed" from "the taxonomy is
genuinely empty".

Refusing at the tool boundary, with a message that names the taxonomy, converts a
silent semantic failure into a loud syntactic one. This only applies to keys in the
notebook's declared taxonomy set — `priority: 3` on an ordinary property is fine and
must stay fine.

## add_term / remove_term: the real verb

`set_front_matter` with a whole list is the wrong primitive for classification,
because it makes every addition a read-modify-write across a value the agent may hold
stale. `add_term` is the operation a notebook performs, and it is idempotent:

```
  absent key        ──add_term(customer, acme)──▶  customer:
                                                     - acme

  customer: acme    ──add_term(customer, globex)─▶  customer:
   (scalar)                                          - acme
                                                     - globex

  customer:         ──add_term(customer, acme)──▶  unchanged
   - acme                                           (already a member)

  customer:         ──remove_term(customer, acme)▶  key removed entirely
   - acme
```

Scalar promotion matters: real corpora mix `customer: acme` and `customer: [acme]`,
both of which Hugo and `termsOf` accept, so `add_term` must handle either without the
agent having to know which it is.

Removing the last term removes the key rather than leaving `customer: []`. Both are
equivalent to the indexer, but an empty sequence is noise that accumulates in a
hand-edited corpus, and its absence is what a human editing the file would write.

Terms are normalized on read (`normalizeTerm`: lowercase, spaces → dashes) but
**stored as given**. `add_term` therefore compares membership using the normalized
form — so adding `"Acme"` to a document already carrying `acme` is a no-op — while
writing the caller's spelling for a genuinely new term. Comparing raw strings would
let `Acme` and `acme` both land and collapse into one term at index time, which reads
as a duplicate in the file and a single term in the UI.

## Coined terms are reported, not refused

`loadLindenConfig` gives the declared terms (`L2-CONF-TAX-<tax>-TRM-<term>.yml`), but
a Linny notebook coins terms freely — requiring a config file before a term may be
used would make classification a two-step ceremony. So an undeclared term is written
and reported back in `new_terms`. That keeps the common case one call, while a term
invented by a typo (`custommer`) is visible in the response instead of silently
joining the taxonomy.

## What stays the same

Everything else in the safe-write pipeline is untouched: these are front-matter edits,
so they ride `editFM`'s surgical `yaml.Node` path and keep key order and comments;
they inherit the scope check, the degraded gate, the optimistic write, the reindex,
the returned membership, and the audit entry. The change is confined to how a *value*
becomes a node, plus two handlers that compose the existing machinery.

## Rejected alternatives

- **Accept anything and stringify** (today's behaviour). Produces junk terms that
  reach `linny.vim` through the emitted index. This is the bug.
- **Accept a number on a taxonomy key and coerce it to its string form.** Looks
  helpful; means `priority: 3` and `priority: "3"` are the same term, and quietly
  invents a term from data that was not one.
- **A `type` field on `set_front_matter`.** More explicit, more ceremony, and JSON
  already carries enough type information for every shape Linny supports. Revisit
  only if dates are needed.
- **`set_front_matter` with a full list as the only mechanism.** Leaves every
  classification a read-modify-write race against a corpus that syncs every 30 s.
