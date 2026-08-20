# Future work (out of scope for the PoC)

Recorded so the alpha's deliberate boundaries are explicit and retrievable. None of
these are implemented; each is a conscious deferral.

## Indexer / format

- **Hugo/index drift is a known, surfaced divergence.** `lindexer verify --hugo`
  runs the real Hugo reference (vendored layouts in `internal/hugoref`) and diffs ours
  against it; every load-bearing file matches, and the one accepted divergence is the
  L1 term-config index — Hugo emits `{}` for the singular≠plural taxonomies (`tag`,
  `project`) where its data-key lookup fails, while ours embeds the L2 config. A future
  step could reconcile that lookup (spec §13 Q7) or teach the synthetic generator to
  produce Hugo-resolvable L2 config filenames. A `verify_index`-style report over
  `--hugo` in CI is already wired via the coverage check.
- **Incremental updates via `fsnotify`.** Today every write does a full rebuild
  (cheap at ~5k notes). A watch mode would update incrementally. (Epic E0204.)
- **Multi-content-dir support.** The tools assume `content/`; resolve it from the
  notebook config instead.
- **Open questions in the index spec** (`docs/linden-index-spec.md` §13): nested vs.
  flat filenames, title-less/invalid-front-matter handling, the vestigial per-page
  JSON, multi-valued taxonomy semantics, task-count regex, and the source of the
  taxonomy list.

## Writes

- **Surgical, `fred`-style front-matter editing for all value types.** `set`/`unset`
  already edit the YAML node order-preservingly; a full editor would mirror `fred`'s
  `set_bool_val`/`replace_key`/`toggle_bool_val` and validate values against
  `lindenConfig` before writing.
- **Full `lindenConfig` validation.** Reject writes that introduce unknown taxonomies
  or malformed term values, rather than only ensuring the file still parses.
- **`delete` and bulk-retag with out-of-band confirmation.** The quarantine policy
  already flags these as confirmation-required; the confirmation channel and the tools
  are not built.
- **Semantic YAML front-matter git merge driver.** A custom merge driver that parses
  front matter and falls back to a real conflict rather than line-merging taxonomy
  lists — so `projects: [Acme]` vs `project: acme` never silently merges wrong.

## Retrieval

- **Local embeddings.** If ever added: **local only** (ONNX/MiniLM-class + `sqlite-vec`).
  Cloud embedding APIs are rejected outright — they would ship the entire private
  corpus to a third party. Out of scope for v1.
- **Remaining navigate tools:** `co_occurring_terms`, `related`, `due_this_week`,
  `open_items` (names reserved in `docs/tools.md`).

## Auth

- **`OIDCAuthenticator`.** The `Authenticator` interface exists with
  `StaticTokenAuthenticator` as the only implementation; an OIDC one can be added
  without changing callers. Deliberately not built (briefing §5.3).
