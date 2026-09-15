# MCP tool surface

> **Tool names are an API.** Models and saved prompts depend on them. Treat this file
> as the contract: add tools additively, and record any rename/removal here with the
> version it changed in.

- **Surface version:** `v1` (read/navigate tools)
- **Transport:** MCP streamable-HTTP at `/mcp`, behind static bearer auth.
- **Every response** is scope-filtered **in SQL** (deny-by-default; a denied document
  is indistinguishable from a missing one) and passed through **egress redaction**
  before it leaves the server.

## Read / navigate (v1, shipped)

| Tool | Arguments | Returns |
|-------------------|--------------------------------------|------------------------------------------------|
| `search`          | `query: string`, `limit?: int`       | `hits: [{filename, title, snippet, score}]`; FTS5, bm25-ranked, redacted snippets |
| `get_doc`         | `slug: string`                       | `{found, filename, title, props, body}`; redacted; `found:false` when denied/missing |
| `list_taxonomies` | none                                 | `{taxonomies: [string]}`; only taxonomies with a readable document |
| `terms`           | `taxonomy: string`                   | `{terms: [string]}`; only terms with a readable document |
| `docs_by_term`    | `taxonomy: string`, `term: string`   | `{docs: [filename]}`; readable members only |
| `starred_docs`    | none                                 | `{docs: [{filename, title}], scope_filtered}`; ordered by title |

## History (v1, shipped)

All history tools are scope-aware (history/diff of a denied document reports
not-found) and redact free-text output (commit subjects, diff hunks).

| Tool | Arguments | Returns |
|------------------|--------------------------------------|------------------------------------------------|
| `history`        | `slug: string`, `limit?: int`        | `{found, commits: [{hash, author, date, subject}]}`; newest first |
| `diff`           | `slug: string`, `ref: string`        | `{found, diff}`; ref-vs-working diff; `ref` must not begin with `-` |
| `changed_since`  | `since: string`                      | `{docs: [slug]}`; changed & readable; `since` must not begin with `-` |

## Write (v1, shipped)

All write tools run the safe-write pipeline: scope check (denied ⇒ not-found),
degraded-mode gate, atomic + optimistic-concurrent write, quarantine-by-default on
create, reindex, and an entry in the external append-only audit log. Each returns the
document's resulting term membership.

| Tool | Arguments | Returns |
|----------------------|--------------------------------------------|------------------------------------------|
| `create_doc`         | `title`, `front_matter?`, `body?`          | `{ok, slug, quarantined, membership}`; lands in `status: agent-draft`; needs `write:inbox`/`write:*` |
| `append_to_doc`      | `slug`, `text`                             | `{ok, slug, membership}` |
| `update_doc`         | `slug`, `old`+`new` \| `body`, `base_hash?`, `allow_empty?` | `{ok, slug, membership}`; body only, front matter untouched |
| `set_front_matter`   | `slug`, `key`, `value`                     | `{ok, slug, membership, new_terms?}`; order-preserving, typed |
| `unset_front_matter` | `slug`, `key`                              | `{ok, slug, membership}` |
| `archive`            | `slug`                                     | `{ok, slug, membership}`; sets `archived: true` |
| `add_term`           | `slug`, `taxonomy`, `term`                 | `{ok, slug, membership, new_terms?}`; idempotent |
| `remove_term`        | `slug`, `taxonomy`, `term`                 | `{ok, slug, membership}`; idempotent |

Modifying an existing document requires `write:*` (or `write:inbox` for a quarantined
draft).

### Starred documents

`starred_docs` lists the documents whose front matter carries `starred: true`, with
titles so the result reads as a list rather than a set of filenames. It takes no
arguments and orders by title.

Scope is applied inside the query, as everywhere else, so a starred document the
caller may not read is absent and indistinguishable from one that does not exist. The
result carries no count of what was excluded: that would disclose the existence of
documents the caller is denied. `scope_filtered` is a constant `true`, a statement
about the tool rather than a measurement, telling the caller not to report the list as
exhaustive.

Starred **taxonomies** and **terms** are indexed too (from `starred: true` in the L1
and L2 config files) but are not exposed. Both are declared in `lindenConfig` rather
than by documents, so either can exist with no readable members, and exposing them
needs its own answer to the existence-leak question.

### Changing a body with `update_doc`

`get_doc` does not return what is on disk: the body is passed through egress
redaction and wrapped in data delimiters. So `update_doc` enforces one rule, **you
may only replace text you have genuinely seen**, and offers two ways to prove it.

**Anchored (`old` / `new`), the default.** `old` must occur **exactly once** in the
on-disk body. Zero matches and several matches are both refused, and neither changes
the document. An anchor taken from redacted text cannot match, so the
dangerous case degrades into a refusal instead of corrupting the note.

**Whole-body (`body`), the escape hatch.** Allowed only when all three hold:

1. `base_hash` is supplied and still matches the file (use `get_doc`'s `content_hash`);
2. the stored body produces no redactions, so what you read *was* the truth;
3. the submitted body carries no data-delimiter fence.

Any failed precondition refuses the write and points back at anchored mode.

Either way the front-matter block is preserved byte-for-byte: a prose rewrite can
never drop a document's classification, and a body that itself begins with `---` is
written as body. A resulting empty body is refused unless `allow_empty` is passed.
The audit log records a unified diff of the body rather than a second copy of the
document.

`get_doc` returns `content_hash` (hashed from the file, not the index, which can lag
behind it) and `redacted`, which tells you up front whether whole-body mode is
available for that document.

### Front-matter value types

`set_front_matter` writes the value with the type the caller sent, rather than
stringifying it. Types are never inferred from a string's contents, so a string that
looks like a date stays a string.

| Caller sends (JSON)   | Written as            |
|-----------------------|-----------------------|
| `"acme"`              | string                |
| `true`                | bool                  |
| `3`                   | int                   |
| `3.5`                 | float                 |
| `null`                | null                  |
| `["acme","globex"]`   | block sequence        |
| object / nested list  | refused               |

On a **declared taxonomy** key the value must be a string or a list of strings.
Anything else is refused: the indexer reads only those two shapes, so a number would
write cleanly and then produce no term membership at all.

### Managing terms

`add_term` and `remove_term` are the operations to reach for when classifying a
document. They avoid a read-modify-write of the whole list against a corpus that may
be syncing underneath you. Both are idempotent:

- `add_term` creates the key as a list when absent, promotes a single existing value
  to a list, and does nothing when the term is already a member. Membership is
  compared the way the indexer normalizes terms, so adding `Acme` to a document
  already carrying `acme` is a no-op.
- `remove_term` drops the term, and removes the key entirely when the last term goes.

A term with no `L2-CONF-TAX-<tax>-TRM-<term>.yml` is still written, because terms are
open-ended, but it is reported back in `new_terms`, so coining a term is visible
rather than accidental.

## Operational (v1, shipped)

| Tool | Arguments | Returns |
|---------------|-----------|------------------------------------------|
| `sync_status`  | none | `{degraded, conflicted, conflicts, in_progress, detached, read_only, reason}`; live git-safety state |
| `verify_index` | none | `{in_sync, corpus_docs, store_docs, missing_from_store, stale_in_store, conflicted}`; served index vs corpus |
| `session_info` | none | `{server, notebook, caller, status, can_write_now, reason}`; what this connection is |

### Knowing what you may do

`session_info` answers the question a refusal only answers after the fact: what is this
server, who am I, and can I write right now. It takes no arguments and describes the
connection, never the corpus, so it carries no counts, no taxonomy or term names, no
slugs and no filesystem paths.

`caller.can_modify` has three values, because modifying has three outcomes:

| Scope held    | can_modify   | May modify                |
|---------------|--------------|---------------------------|
| `write:*`     | `all`        | anything readable         |
| `write:inbox` | `own-drafts` | only a quarantined draft  |
| neither       | `none`       | nothing                   |

They are computed from the same predicates the write path enforces with, so the report
cannot drift from what actually happens.

`can_write_now` is the conjunction of every gate a write must pass: the write tools
being registered at all, the working tree not being degraded, and the caller holding a
write scope. Any one of them failing produces the same symptom, so the tool combines
them rather than leaving that to the caller, and `reason` names the first gate that
fails. Creating counts as writing: a `write:inbox` token gets `can_write_now: true`
with a reason saying existing documents need `write:*`.

Reporting a caller its own scopes discloses nothing. It holds the token already, and
it can enumerate its scopes today by attempting writes and reading the refusals. The
only thing the old behaviour achieved was making that discovery expensive and writing
a `denied` audit entry for every probe.

## Planned (not yet shipped)

Recorded so names are reserved and stable when implemented:

- `co_occurring_terms(taxonomy, term)`: terms frequently co-tagged with a given term.
- `related(doc)`: documents related by shared taxonomy membership.
- `due_this_week()`: documents with a due date in the current week.
- `open_items(project)`: open task-list items for a project.
(Note: the Hugo-reference JSON diff ships as the `lindexer verify` CLI; the `verify_index` MCP tool checks the served index against the corpus on disk.)
- `delete` and bulk-retag require out-of-band confirmation (policy already flags them).

## Change log

- **v1** (2026-08-20): initial read/navigate surface (`search`, `get_doc`,
  `list_taxonomies`, `terms`, `docs_by_term`); history tools (`history`, `diff`,
  `changed_since`); write tools (`create_doc`, `append_to_doc`, `set_front_matter`,
  `unset_front_matter`, `archive`); operational (`sync_status`, `verify_index`).
- **v1.1**: typed front-matter values (lists and real scalars, replacing the
  stringifying writer) and the `add_term` / `remove_term` tools.
- **v1.2**: `update_doc` (anchored body edits, guarded whole-body replacement);
  `get_doc` gained `content_hash` and `redacted`.
- **v1.3**: `starred_docs` and `session_info`.
