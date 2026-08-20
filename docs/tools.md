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

## Planned (not yet shipped)

Recorded so names are reserved and stable when implemented:

- `co_occurring_terms(taxonomy, term)` — terms frequently co-tagged with a given term.
- `related(doc)` — documents related by shared taxonomy membership.
- `due_this_week()` — documents with a due date in the current week.
- `open_items(project)` — open task-list items for a project.
- Write tools (quarantine-default): `create_doc`, `set_front_matter`,
  `unset_front_matter`, `append_to_doc`, `archive`.
- History: `history(doc)`, `diff(doc, ref)`, `changed_since(date)`.
- Operational: `sync_status()`, `verify_index()`.

## Change log

- **v1** (2026-08-20): initial read/navigate surface — `search`, `get_doc`,
  `list_taxonomies`, `terms`, `docs_by_term`.
