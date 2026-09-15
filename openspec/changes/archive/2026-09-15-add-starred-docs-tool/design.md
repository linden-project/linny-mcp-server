# Design: starred_docs

## Why this one is small

The hard part of "expose starred" is not documents. It is that `starred` names three
different things with three different sources:

| Kind      | Source                        | Can exist with zero readable docs |
|-----------|-------------------------------|-----------------------------------|
| document  | front matter `starred: true`  | no, it is itself a document       |
| taxonomy  | `L1-CONF-TAX-*.yml`           | yes                               |
| term      | `L2-CONF-TAX-*-TRM-*.yml`     | yes                               |

Documents are the easy case precisely because a starred document *is* a document: it
flows through `ReadableFilenamesSQL()` like every other row, and the existing
`DocsByTermScoped` is the template.

The config-derived kinds are not rows in the corpus. `build.go` notes that starred
terms carry no occurrence filter, matching Hugo, so a term can be starred while no
record uses it. Returning those would walk past the rule `ListTaxonomiesScoped`
already enforces, which hides a taxonomy with no readable documents "so a fully-denied
taxonomy's existence is not leaked". Scoping this change to documents sidesteps that
entirely, and leaves the question intact for whoever picks it up.

## Why there is no count of hidden documents

The tempting affordance is to tell the agent how much it is not seeing:

```
{docs: [...], hidden: 3}
```

It is rejected. `internal/authz` states its purpose as applying authorization "as a
query-time SQL predicate, never a post-filter, so document existence is never leaked
to an unauthorized caller". A count is a post-filter disclosure wearing different
clothes: a caller holding `deny:taxonomy:health` would learn that three starred health
notes exist. It would also undercut `get_doc`, which answers `found:false` for a
denied document specifically so that denied is indistinguishable from missing.

The architecture makes the leak inconvenient to build, which is a good sign. With the
scope applied as a predicate, the hidden count is not available as a side effect of
the query. Producing it would mean issuing a second, deliberately unscoped query. Any
implementation has to go out of its way, so the rule is easy to keep.

## What replaces it

The thing worth preserving from that idea is the agent knowing its list may be
partial, so it does not report a filtered list as exhaustive. A constant does that
without measuring anything:

```
{docs: [...], scope_filtered: true}
```

`scope_filtered` is always `true`. It is a statement about the tool, not about the
corpus, so it carries no information about what was removed. For a `read:*` token it
is simply uninformative rather than wrong. The alternative, leaving the caveat in the
tool description only, costs nothing at runtime and relies on the agent having read
the description; the field is cheap enough to prefer.

## Shape

```
starred_docs()  ──▶  { docs: [{filename, title}], scope_filtered: true }
```

**Filename and title, not bare filenames.** `docs_by_term` returns bare filenames, and
consistency argues for matching it. But this tool exists to answer "show me my starred
notes", and a list of filenames forces an immediate fan-out of `get_doc` calls purely
to render it. Titles are already redacted on every other read path, so the machinery
exists. `search`'s fuller shape was rejected: a snippet in a curated list is noise,
and it costs body text in every response.

**Ordered by title.** Filename order is an implementation detail of how records are
named. A human-facing list should read in a stable, meaningful order, and title order
is both. Deterministic ordering also keeps the tool testable.

**No arguments.** Like `list_taxonomies` and `verify_index`. Filtering starred notes by
taxonomy is a composition of `starred_docs` and `docs_by_term` that the agent can do
itself, and building it in would invite the argument creep this surface has so far
avoided.
