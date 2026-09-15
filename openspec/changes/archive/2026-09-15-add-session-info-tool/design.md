# Design: session_info

## One tool, because the question is one question

The metadata splits naturally into two subjects, "what is this server" and "who am I",
and separate tools would model that cleanly. They are merged anyway, because the
question an agent actually has is neither of those. It is "can I do the thing I am
about to try", and answering it needs both halves plus the live tree state. Splitting
them means an agent must make three calls and combine the results correctly, which is
exactly the step that goes wrong.

## Why a boolean cannot describe modify permission

`ensureModify` has three outcomes, not two:

| Scope held    | May modify                          | Reported as  |
|---------------|-------------------------------------|--------------|
| `write:*`     | anything readable                   | `all`        |
| `write:inbox` | only a quarantined draft            | `own-drafts` |
| neither       | nothing                             | `none`       |

A boolean has to round the middle case to either `true` or `false`, and both are
wrong in the way that matters: `true` invites an agent to attempt edits that will be
refused, `false` stops it creating drafts it is entitled to create. The middle case is
also the common one, since `write:inbox` is the conservative grant a cautious operator
hands out first.

The three values are computed from `CanWriteAll()` and `CanWriteInbox()`. Those are
the functions the write path itself calls, so there is no second implementation of the
rule to keep in step. If the rule changes, both move together or neither compiles.

## Why `can_write_now` exists

Four independent conditions produce the identical symptom:

```
   registration            guard                scope               target
        │                    │                     │                   │
 stateDir + audit      tree clean?          write:* / inbox      readable?
 + not read-only            │                     │                   │
        └────────────────── AND ──────────────────┴───────────────────┘
                             │
                             ▼
                    "I cannot write"
```

Exposing the conditions separately is not the same as answering the question. An agent
handed four fields has to know the conjunction, know which gate dominates, and phrase
it correctly. Handed `can_write_now: false` with a reason naming the failing gate, it
has nothing to get wrong. The separate fields stay in the response for a human reading
over its shoulder, but the verdict is the field that matters.

`reason` names one gate, the first that fails in the order above, rather than listing
every problem. A caller that fixes the named gate and asks again gets the next one.
That keeps the string short and keeps the tool from implying that fixing everything
listed is required.

## Single sources of truth

Two values in this response are computed elsewhere in the server, and both must be
read rather than recomputed:

- **Degraded state** comes through the same guard-backed accessor `sync_status` uses.
  Two tools reporting tree state from two computations is a drift bug waiting for a
  confusing afternoon.
- **Writes enabled** is currently an inline expression in `server.go` deciding whether
  to register the write tools: `Guard != nil && Audit != nil && !Guard.ForcedReadOnly()`.
  If the report recomputes it, a future edit to registration can silently make the tool
  lie about a server whose write tools are absent. Extracting it to one predicate that
  both registration and the report call removes the possibility.

## Reporting scopes is not a disclosure

On a project whose authorization package exists to avoid leaking what a caller may not
see, a tool that returns authorization state deserves an explicit argument rather than
an assumption.

It leaks nothing. The caller holds the token. Its scopes are a property of the
credential it already possesses, not of the corpus, and it can enumerate them today by
attempting operations and reading the refusals, which already name the missing scope
(`requires write:* (or write:inbox for a quarantined draft)`). The only thing the
current design achieves is making that enumeration expensive and writing a `denied`
audit entry for every probe, turning a capability question into a log of apparent
policy violations.

What must not appear here is anything about the corpus: no counts, no taxonomy names,
no document existence. Every field is about the connection, the credential, or the
server process. `notebook.name` is the configured label ("default", "business"), not a
path and not a fact about content.

## Rejected alternatives

- **`whoami` and `server_info` as two tools.** Cleaner modelling, worse ergonomics for
  the one question that motivates the change.
- **Interpreted permissions only.** Hides the scope vocabulary a person needs when they
  go to edit the token file, which is the next thing they do after reading this.
- **Raw scopes only.** Correct and minimal, but it puts the burden of understanding the
  vocabulary on the model at exactly the moment it is already confused.
- **Reporting paths, as `_indexer_info.json` does.** That file is consumed by tooling
  on the same host. This response goes to an agent that might relay it anywhere.
