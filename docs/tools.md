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
| `set_front_matter`   | `slug`, `key`, `value`                     | `{ok, slug, membership}` — order-preserving |
| `unset_front_matter` | `slug`, `key`                              | `{ok, slug, membership}` |
| `archive`            | `slug`                                     | `{ok, slug, membership}` — sets `archived: true` |

Modifying an existing document requires `write:*` (or `write:inbox` for a quarantined
draft).

## Planned (not yet shipped)

Recorded so names are reserved and stable when implemented:

- `co_occurring_terms(taxonomy, term)` — terms frequently co-tagged with a given term.
- `related(doc)` — documents related by shared taxonomy membership.
- `due_this_week()` — documents with a due date in the current week.
- `open_items(project)` — open task-list items for a project.
- Operational: `sync_status()`, `verify_index()`.
- `delete` and bulk-retag — require out-of-band confirmation (policy already flags them).

## Change log

- **v1** (2026-08-20): initial read/navigate surface — `search`, `get_doc`,
  `list_taxonomies`, `terms`, `docs_by_term`; history tools — `history`, `diff`,
  `changed_since`; write tools — `create_doc`, `append_to_doc`, `set_front_matter`,
  `unset_front_matter`, `archive`.
