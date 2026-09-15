# Design: update_doc

## The constraint that shapes everything

`get_doc` does not return what is on disk:

```
  disk                    store              get_doc response
  ───────────────────────────────────────────────────────────────────
  body with an            indexed            Redact()  ──▶ [REDACTED]
  API key in a       ──▶  copy of      ──▶   Delimit() ──▶ <data> … </data>
  code fence              the body
```

Both transforms are one-way and intentional (egress redaction is a hard project
rule; delimiting marks the corpus as untrusted data). The consequence for writing is
that **the agent never holds the authoritative body**. Any tool that accepts a body
and writes it back verbatim will, sooner or later, overwrite a real secret with the
literal text `[REDACTED]` or paste a `<data>` fence into a note.

That is not a hypothetical: the redactor fires on exactly the kind of content a
developer's notebook contains (tokens pasted into a scratch note, `.env` snippets,
connection strings).

So the design question is not "replace or append" — it is **how does the server know
the agent is replacing text it actually saw?**

## Two modes, two different proofs

| Mode        | Proof that the agent saw the truth            | Failure mode          |
|-------------|-----------------------------------------------|-----------------------|
| Anchored    | `old` must match on-disk bytes exactly, once   | Fails closed (refuse) |
| Whole-body  | hash match + nothing was redacted + no fence   | Refused unless proven |

**Anchored is the default and the safe primitive.** The server searches the on-disk
body for `old`:

- exactly one match → splice in `new`
- zero matches → refuse: the anchor is stale, or it crosses redacted content
- two or more matches → refuse as ambiguous, ask for more surrounding context

An anchor taken from redacted text *cannot* match, so the dangerous case degrades
into a refusal rather than corruption. The exactly-one rule also removes the
"replaced the wrong paragraph" class of bug. This mirrors the edit primitive agents
already use well, and it keeps payloads small — relevant for long notes, where
re-emitting the whole body invites truncation and invented text.

**Whole-body is the escape hatch**, for genuine rewrites and for freshly created
drafts. It is allowed only when all three preconditions hold:

1. `base_hash` supplied and equal to `gitsafe.HashFile(path)` — the file has not
   moved since the agent read it;
2. `redact.Redact(currentBody)` reports **0** replacements — nothing was hidden, so
   what the agent saw *was* the truth;
3. the submitted body contains no delimiter fence — catches the common mistake of
   echoing back the `<data>` wrapper.

Precondition 2 is what makes the mode sound, and it is cheap: `Redact` already
returns a replacement count that the current call sites discard.

Why not a separate `replace_in_doc` tool? Two tools split one concept and force the
agent to choose before it knows whether a whole rewrite is even permitted. One tool
with a clear refusal message ("whole-body replacement is not available for this
document because part of it is redacted; use old/new") teaches the agent the rule at
the moment it matters.

## Front matter is off-limits

`update_doc` splits on the front-matter fence and rewrites only the body half. This
is a hard rule, not a default: taxonomy membership *is* the document's meaning in a
Linny notebook, and a prose rewrite must not be able to drop `customer:` or flip
`archived:`. Front-matter changes keep going through the surgical, order-preserving
`set_front_matter` / `unset_front_matter` / `archive` path.

It also sidesteps a trap — the whole-body mode would otherwise let an agent submit a
body that happens to start with `---`, silently turning prose into front matter.

## Optimistic concurrency, for real this time

The existing edit tools do:

```go
raw, hash, _, _, _ := w.loadForEdit(slug)   // read + hash
…
gitsafe.WriteIfUnchanged(path, new, hash)   // compare, microseconds later
```

The hash is produced inside the same call, so it guards a race that essentially
cannot happen and does nothing about the one that can: the agent read the note three
turns ago and `git-sync` has since pulled a remote edit. Surfacing `content_hash`
from `get_doc` and accepting it as `base_hash` closes the real window.

`content_hash` must be hashed **from the file**, not from the indexed copy: on the
deployed host a separate `lindexer watch` unit rebuilds the index, so the store can
lag the disk. Hashing the disk is also what `WriteIfUnchanged` compares against, so
the two agree by construction.

`base_hash` is optional in anchored mode (the exactly-one-match rule is already a
strong integrity check) and required in whole-body mode.

## Audit

The existing tools pass the entire new content as the audit `Diff`. For full-body
rewrites that would store two copies of every note that is ever edited. `update_doc`
records a unified diff instead. Existing tools are left alone — changing their audit
payload is a separate concern.

## Rejected alternatives

- **Whole-body replace with no preconditions.** Simplest, and the one that quietly
  destroys secrets. Rejected outright.
- **Un-redacting the body for writers holding `write:*`.** Would make whole-body
  always available, at the cost of turning the egress-redaction guarantee into a
  scope-dependent one. The project's secret-hygiene rule is absolute; redaction must
  not have an "unless" clause.
- **Line-range edits (`replace lines 10–14`).** Cheap to implement, brittle against a
  corpus that syncs underneath the agent, and meaningless once line numbers shift.
- **Section-addressed edits (`under heading X`).** Genuinely nice for a notebook, but
  it needs a markdown structure model. Worth a later change once anchored editing is
  proven.
