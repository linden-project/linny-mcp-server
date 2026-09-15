## Why

The write surface can create documents and append to them, but it cannot **change**
anything already written. `append_to_doc` is the only body mutation there is, so an
agent cannot fix a typo, rewrite a paragraph, restructure a note, or correct its own
earlier draft. In practice every "update this note" request fails, and the agent
reports that it cannot write.

The obvious fix — "add a whole-body replace" — is unsafe in this architecture, and
that is the real reason this tool does not exist yet:

- `get_doc` returns the body **redacted** (`redact.Redactor.Redact`) and **wrapped in
  data delimiters** (`defense.Delimit`). The text an agent holds is therefore *not*
  the text on disk.
- A naive write-back would bake `[REDACTED]` placeholders and the delimiter fence
  into the corpus permanently. That is silent, unrecoverable data loss in a notebook
  whose whole point is being the user's memory.
- The corpus is also live: `git-sync` pulls remote edits every 30 s on the deployed
  host, so a body read several turns ago may already be stale.

So `update_doc` has to be designed around one invariant: **an agent may only replace
text it has genuinely seen.** Anchored editing enforces that and fails closed;
whole-body replacement can only be permitted when it is provable that the agent saw
the true content.

## What Changes

- Add an MCP write tool **`update_doc`** with two mutually exclusive modes:
  - **Anchored (default)** — `old` / `new`. The server locates `old` in the
    **on-disk** body and requires **exactly one** match. Zero matches or several
    matches are refused with an actionable message. Because the anchor is matched
    against real file content, an anchor drawn from redacted or stale text simply
    fails to match: it fails closed, never destructively.
  - **Whole-body** — `body`. Permitted **only** when all three hold: a `base_hash` is
    supplied and still matches the file, the current body redacts to itself (nothing
    was hidden from the agent), and the submitted text carries no delimiter fence.
    Any failed precondition refuses the write and points the agent at anchored mode.
- **Front matter is never touched.** `update_doc` rewrites the body only; taxonomy
  and state transitions stay with `set_front_matter` / `unset_front_matter` /
  `archive`, so an agent rewriting prose cannot destroy a document's classification.
- **Guard against wipes**: a body that is empty or whitespace-only is refused unless
  `allow_empty: true` is passed explicitly.
- **Real optimistic concurrency**: `get_doc` gains `content_hash` (hashed from the
  file on disk, not from the index) and `redacted`, so an agent can carry a base
  version across turns. Today's write path re-reads and hashes immediately before
  writing, which only closes a microsecond-wide race, not a multi-turn one.
- `update_doc` otherwise reuses the existing safe-write pipeline unchanged: scope
  check (denied == not-found), degraded-mode gate, atomic `WriteIfUnchanged`,
  reindex, resulting term membership, and an audit entry — recording a unified diff
  rather than the full new content.
- Scopes are unchanged: modifying an existing document needs `write:*`, or
  `write:inbox` when the target is the agent's own quarantined draft.

## Capabilities

### Modified Capabilities
- `mcp-write-tools`: adds `update_doc` (anchored + guarded whole-body), the
  front-matter-preservation rule, and the empty-body guard.
- `mcp-read-tools`: `get_doc` returns `content_hash` and `redacted`.

## Impact

- Modified: `internal/mcp/write.go` (the `update_doc` handler and body-splice
  helpers), `internal/mcp/tools.go` (`getDocOut` gains two fields; hash the file on
  disk), `docs/tools.md`. No new dependencies; `gitsafe.HashFile`,
  `gitsafe.WriteIfUnchanged`, `splitFrontMatter`, and `redact.Redact`'s existing
  replacement count supply everything needed.
- `getDocOut` gains fields only — additive, so existing clients are unaffected.
- Reading a document now stats and hashes one file per `get_doc` call. Negligible
  against a ~5k-note corpus, and it is what makes the base version meaningful.

## Non-Goals

Deliberately out of scope; each deserves its own change:

- **Typed front-matter values.** `set_front_matter` writes scalars only —
  `scalarNode()` stringifies anything else, so `tags: ["a","b"]` lands as the string
  `"[a b]"`. An agent therefore still cannot tag or re-classify a document. This is
  the single most valuable follow-up (already flagged in `docs/future.md`).
- Delete and rename/move tools (`authz` parses `delete:*`; nothing consumes it).
- Section-aware editing addressed by heading.
- Promoting a document out of quarantine as a first-class transition.
