## 1. Read-side base version

- [ ] 1.1 `getDocOut` gains `content_hash` + `redacted`
- [ ] 1.2 hash the file on disk (`gitsafe.HashFile`), not the indexed copy
- [ ] 1.3 derive `redacted` from the replacement count `Redact` already returns

## 2. update_doc handler (internal/mcp/write.go)

- [ ] 2.1 `updateDocIn` (slug, old, new, body, base_hash, allow_empty); reject both
      modes or neither
- [ ] 2.2 anchored splice against the on-disk body: exactly-one-match rule, distinct
      refusals for zero and for N>1 matches
- [ ] 2.3 whole-body preconditions: base_hash matches, zero redactions, no delimiter
      fence; actionable refusal pointing at anchored mode
- [ ] 2.4 body-only rewrite via `splitFrontMatter`; front matter preserved verbatim
- [ ] 2.5 empty-body guard unless `allow_empty`
- [ ] 2.6 reuse `ensureModify`, `guard.EnsureWritable`, `WriteIfUnchanged`, `finish`
- [ ] 2.7 audit a unified diff rather than the full content
- [ ] 2.8 register the tool with a description that states the anchored default

## 3. Tests

- [ ] 3.1 anchored replace: single match applied, rest of body byte-identical
- [ ] 3.2 anchored refusals: zero matches; N>1 matches (message names the count)
- [ ] 3.3 anchor spanning redacted content refuses and writes no placeholder
- [ ] 3.4 whole-body: happy path with a valid base_hash on a non-redacted doc
- [ ] 3.5 whole-body refused — redacted doc / stale base_hash / missing base_hash /
      echoed delimiter fence
- [ ] 3.6 front matter byte-identical after a full body rewrite; a body starting with
      `---` is not promoted to front matter
- [ ] 3.7 empty body refused, then allowed with `allow_empty`
- [ ] 3.8 refused while degraded; refused for `write:inbox` on a non-draft; unreadable
      target reports not-found
- [ ] 3.9 membership returned; audit entry present with a diff payload
- [ ] 3.10 protocol-level e2e: `get_doc` → `update_doc` with the returned hash → read
      back

## 4. Docs & gate

- [ ] 4.1 `docs/tools.md`: add `update_doc`, document both modes and the refusal rules
- [ ] 4.2 `docs/future.md`: record typed front-matter values as the next follow-up
- [ ] 4.3 `nix flake check` green (gotest + lint + coverage >= 70%)
