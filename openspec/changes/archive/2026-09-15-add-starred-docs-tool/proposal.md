## Why

Starring is already a first-class thing in a Linny notebook, and the indexer already
records it. `build.go` collects every record whose front matter carries
`starred: true` into `Graph.StarredDocs`, `Populate` writes it to the `docs.starred`
column, and `emit.go` writes `_index_docs_starred.json` for `linny.vim`.

Nothing ever reads the column back:

```
front matter  starred: true ──▶ StarredDocs ──┬──▶ _index_docs_starred.json ──▶ linny.vim  ok
                                              └──▶ docs.starred (SQLite)    ──▶ nothing
```

`grep -n starred internal/index/query*.go internal/mcp/*.go` returns nothing. So the
editor can show your starred notes and an agent cannot ask for them. The state is
written on every index build and read by no query.

An agent is not entirely blind to it: `get_doc` returns the whole `props` map, so
`starred: true` is visible on a document it already holds. It cannot enumerate. The
property is visible but not queryable, which is the gap this closes.

The write half already works. Since typed front-matter values shipped,
`set_front_matter(slug, "starred", true)` writes a real YAML bool, which is exactly
what `build.go` type-asserts on. Only reading back is missing.

## What Changes

- Add an MCP read tool **`starred_docs`**, taking no arguments and returning
  `{docs: [{filename, title}], scope_filtered: true}`.
- Add `Store.StarredDocsScoped`, which selects `starred = 1` intersected with the
  caller's readable-filenames subquery **in SQL**, following `DocsByTermScoped`.
- Titles pass through the egress redactor, as on every other read tool.
- Results are ordered by title, because the tool exists to be read by a person
  through an agent.
- Reserve the name in `docs/tools.md` and document the tool.

## Capabilities

### Modified Capabilities
- `mcp-read-tools`: adds `starred_docs`.

## Impact

- Modified: `internal/index/query_scoped.go` (one query plus a small result type),
  `internal/mcp/tools.go` (input/output types, handler, registration),
  `docs/tools.md`. No schema change: `docs.starred` already exists and is already
  populated.
- No new dependencies, no new authorization vocabulary, no write-path changes.
- The query is a full scan of a small table. At notebook scale (thousands of records)
  that is not worth an index; one becomes worth adding only if starred counts ever
  grow into the same order as the corpus, which curation makes unlikely.

## Non-Goals

- **Starred taxonomies and terms.** Both are indexed (`terms.starred`,
  `Graph.StarredTaxonomies`) and both are derived from `lindenConfig` rather than from
  documents, so either can exist with zero readable members. Exposing them safely
  needs its own decision about the existence-leak rule (see `design.md`), and they
  serve agent orientation rather than the human-facing list this tool is for.
- **A `limit` argument.** Starring is deliberate curation, so the result stays small.
  Adding `limit` later is additive and breaks nothing.
- **A count of documents hidden by scope.** Rejected on security grounds; `design.md`
  records why.
