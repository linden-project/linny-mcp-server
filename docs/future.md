# Future work (out of scope for the PoC)

Recorded so the alpha's deliberate boundaries are explicit and retrievable. None of
these are implemented; each is a conscious deferral.

## Indexer / format

- **`verify_index` — diff our JSON against Hugo.** The safety net for the "own index"
  decision: run Hugo's indexer and ours over the same corpus and diff the JSON
  (arrays as sets; ignore Hugo's `"TODO"` placeholders and `product_*`). The synthetic
  corpus already emits a Hugo config for this. (Epic E0203 / E0504.)
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
