## Why

The JSON index format is already a de-facto contract between the indexer,
`linny.vim`, and `linnydroid` — but it has never been written down. Writing it down
is the precondition for an independent indexer: without it, the indexer epics have no
target and `verify_index` has nothing to diff against. The briefing (§3) is explicit:
produce the spec **before** any indexer code.

A stranded earlier spec exists on the project website (Linden Specification 0.1.2 and
0.2.0) but predates the 2021 Hugo/"Carl" port, which redesigned the index files and —
per the project's own "Project Reboot" post — never folded those redesigns back into
the spec. We use that document for **vocabulary and concepts only** and rewrite from
the ground up, documenting the format **as Hugo emits it today**.

## What Changes

- Add `docs/linden-index-spec.md`, **Linden Index Specification v0.3.0**, describing
  every index file the Hugo indexer emits today: filename convention, on-disk layout,
  exact JSON shape, field semantics, and which `linny.vim` feature consumes it.
- Document the front-matter model, the slug convention, and `lindenConfig`
  (`L1-CONF-TAX-*`, `L2-CONF-TAX-*-TRM-*`) as the source of term/taxonomy metadata.
- Record every ambiguity, deprecation, and apparently-accidental behaviour as an
  explicit **Open Questions** section rather than silently normalising.
- Resolve the spec version number to **v0.3.0** (continuing the project's numbering:
  0.1.2 → 0.2.0 → the unwritten Hugo redesign → 0.3.0).
- Correct the briefing's `$FORMAT`/`$INCLUDE` reference: those directives do not exist
  in the codebase; the corpus uses `[[...]]` **WikiTags**, which are an editor concern,
  not an index concern. This correction is captured in the spec.

## Capabilities

### New Capabilities
- `index-format`: the normative description of the Linden JSON index — the set of
  emitted files and their shapes — that the indexer MUST produce and `verify_index`
  MUST check. Each requirement is a machine-checkable clause the indexer epics
  implement.

### Modified Capabilities

## Impact

- New file: `docs/linden-index-spec.md`.
- New spec capability `index-format` becomes the contract for milestone 02
  (indexer) and the `verify` subcommand.
- Downstream: `linny.vim` / `linnydroid` compatibility is defined here.
