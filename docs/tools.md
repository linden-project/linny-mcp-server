# MCP tool surface

> **Tool names are an API.** Models and saved prompts depend on them. Treat this file
> as the contract: add tools additively, and record any rename/removal here with the
> version it changed in.

- **Surface version:** `v1` (read/navigate tools)
- **Transport:** MCP streamable-HTTP at `/mcp`, behind static bearer auth.
- **Every response** is scope-filtered **in SQL** (deny-by-default; a denied document
  is indistinguishable from a missing one) and passed through **egress redaction**
  before it leaves the server.

## Read / navigate (v1 — shipped)

| Tool | Arguments | Returns |
|-------------------|--------------------------------------|------------------------------------------------|
| `search`          | `query: string`, `limit?: int`       | `hits: [{filename, title, snippet, score}]` — FTS5, bm25-ranked, redacted snippets |
| `get_doc`         | `slug: string`                       | `{found, filename, title, props, body}` — redacted; `found:false` when denied/missing |
| `list_taxonomies` | —                                    | `{taxonomies: [string]}` — only taxonomies with a readable document |
| `terms`           | `taxonomy: string`                   | `{terms: [string]}` — only terms with a readable document |
| `docs_by_term`    | `taxonomy: string`, `term: string`   | `{docs: [filename]}` — readable members only |

## History (v1 — shipped)

All history tools are scope-aware (history/diff of a denied document reports
not-found) and redact free-text output (commit subjects, diff hunks).

| Tool | Arguments | Returns |
|------------------|--------------------------------------|------------------------------------------------|
| `history`        | `slug: string`, `limit?: int`        | `{found, commits: [{hash, author, date, subject}]}` — newest first |
| `diff`           | `slug: string`, `ref: string`        | `{found, diff}` — ref-vs-working diff; `ref` must not begin with `-` |
| `changed_since`  | `since: string`                      | `{docs: [slug]}` — changed & readable; `since` must not begin with `-` |

## Write (v1 — shipped)

All write tools run the safe-write pipeline: scope check (denied ⇒ not-found),
degraded-mode gate, atomic + optimistic-concurrent write, quarantine-by-default on
create, reindex, and an entry in the external append-only audit log. Each returns the
document's resulting term membership.

| Tool | Arguments | Returns |
|----------------------|--------------------------------------------|------------------------------------------|
| `create_doc`         | `title`, `front_matter?`, `body?`          | `{ok, slug, quarantined, membership}` — lands in `status: agent-draft`; needs `write:inbox`/`write:*` |
| `append_to_doc`      | `slug`, `text`                             | `{ok, slug, membership}` |
| `set_front_matter`   | `slug`, `key`, `value`                     | `{ok, slug, membership, new_terms?}` — order-preserving, typed |
| `unset_front_matter` | `slug`, `key`                              | `{ok, slug, membership}` |
| `archive`            | `slug`                                     | `{ok, slug, membership}` — sets `archived: true` |
| `add_term`           | `slug`, `taxonomy`, `term`                 | `{ok, slug, membership, new_terms?}` — idempotent |
| `remove_term`        | `slug`, `taxonomy`, `term`                 | `{ok, slug, membership}` — idempotent |

Modifying an existing document requires `write:*` (or `write:inbox` for a quarantined
draft).

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
document — they avoid a read-modify-write of the whole list against a corpus that may
be syncing underneath you. Both are idempotent:

- `add_term` creates the key as a list when absent, promotes a single existing value
  to a list, and does nothing when the term is already a member. Membership is
  compared the way the indexer normalizes terms, so adding `Acme` to a document
  already carrying `acme` is a no-op.
- `remove_term` drops the term, and removes the key entirely when the last term goes.

A term with no `L2-CONF-TAX-<tax>-TRM-<term>.yml` is still written — terms are
open-ended — but it is reported back in `new_terms`, so coining a term is visible
rather than accidental.

## Operational (v1 — shipped)

| Tool | Arguments | Returns |
|---------------|-----------|------------------------------------------|
| `sync_status`  | — | `{degraded, conflicted, conflicts, in_progress, detached, read_only, reason}` — live git-safety state |
| `verify_index` | — | `{in_sync, corpus_docs, store_docs, missing_from_store, stale_in_store, conflicted}` — served index vs corpus |

## Planned (not yet shipped)

Recorded so names are reserved and stable when implemented:

- `co_occurring_terms(taxonomy, term)` — terms frequently co-tagged with a given term.
- `related(doc)` — documents related by shared taxonomy membership.
- `due_this_week()` — documents with a due date in the current week.
- `open_items(project)` — open task-list items for a project.
(Note: the Hugo-reference JSON diff ships as the `lindexer verify` CLI; the `verify_index` MCP tool checks the served index against the corpus on disk.)
- `delete` and bulk-retag — require out-of-band confirmation (policy already flags them).

## Change log

- **v1** (2026-08-20): initial read/navigate surface — `search`, `get_doc`,
  `list_taxonomies`, `terms`, `docs_by_term`; history tools — `history`, `diff`,
  `changed_since`; write tools — `create_doc`, `append_to_doc`, `set_front_matter`,
  `unset_front_matter`, `archive`; operational — `sync_status`, `verify_index`.
- **v1.1**: typed front-matter values (lists and real scalars, replacing the
  stringifying writer) and the `add_term` / `remove_term` tools.
