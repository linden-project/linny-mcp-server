## ADDED Requirements

### Requirement: update_doc edits an existing document's body

The server SHALL expose an MCP write tool `update_doc` that replaces text in an
existing document's body. It SHALL support an anchored mode (`old`/`new`) and a
whole-body mode (`body`), and SHALL refuse a call that supplies both or neither.
`update_doc` SHALL require `write:*`, or `write:inbox` when the target is a
quarantined draft; a document the caller cannot read SHALL be reported as not-found.

#### Scenario: Anchored replacement succeeds

- **WHEN** `update_doc` is called with an `old` string occurring exactly once in the
  document body and a `new` string
- **THEN** that occurrence is replaced, the rest of the body is byte-identical, and
  the result reports success

#### Scenario: Both modes supplied is refused

- **WHEN** `update_doc` is called with both `body` and `old`
- **THEN** the call is refused and the document is unchanged

#### Scenario: Modifying an existing doc requires write:*

- **WHEN** a caller holding only `read:*` and `write:inbox` calls `update_doc` on a
  document that is not a quarantined draft
- **THEN** the tool refuses and the document is unchanged

### Requirement: Anchored edits fail closed

An anchored `update_doc` SHALL match `old` against the document's on-disk body and
SHALL apply the edit only when there is **exactly one** match. Zero matches SHALL be
refused as a stale or unmatched anchor, and two or more matches SHALL be refused as
ambiguous. Both refusals SHALL leave the document unchanged and SHALL state how to
proceed.

#### Scenario: Anchor not found

- **WHEN** `old` does not occur in the on-disk body
- **THEN** the write is refused, the document is unchanged, and the message advises
  re-reading the document and choosing a different anchor

#### Scenario: Ambiguous anchor

- **WHEN** `old` occurs more than once in the body
- **THEN** the write is refused, the document is unchanged, and the message reports
  the number of matches and asks for more surrounding context

#### Scenario: Anchor drawn from redacted text cannot corrupt

- **WHEN** the body contains a region the redactor replaced in the read response and
  an anchor covering that region is submitted
- **THEN** the anchor does not match the on-disk body, the write is refused, and no
  redaction placeholder is written to the corpus

### Requirement: Whole-body replacement requires proof the agent saw the truth

`update_doc` in whole-body mode SHALL be permitted only when all of the following
hold: a `base_hash` is supplied and equals the current on-disk content hash; the
current body produces zero redactions; and the submitted body contains no data
delimiter fence. If any precondition fails the write SHALL be refused, the document
SHALL be unchanged, and the message SHALL direct the caller to anchored mode.

#### Scenario: Whole-body refused for a redacted document

- **WHEN** whole-body `update_doc` targets a document whose body the redactor alters
- **THEN** the write is refused and the message directs the caller to use `old`/`new`

#### Scenario: Whole-body refused on a stale base hash

- **WHEN** whole-body `update_doc` supplies a `base_hash` that no longer matches the
  file
- **THEN** the write is refused as a retryable stale write and the document is
  unchanged

#### Scenario: Whole-body refused without a base hash

- **WHEN** whole-body `update_doc` is called with no `base_hash`
- **THEN** the write is refused and the document is unchanged

#### Scenario: Echoed delimiter fence is rejected

- **WHEN** the submitted body contains the data delimiter fence used on read
  responses
- **THEN** the write is refused and no fence is written to the corpus

### Requirement: update_doc never modifies front matter

`update_doc` SHALL rewrite only the body portion of a document. The front-matter
block SHALL be preserved byte-for-byte, including key order and comments, regardless
of the mode used or the content submitted.

#### Scenario: Front matter survives a body rewrite

- **WHEN** `update_doc` replaces the entire body of a document carrying several
  front-matter keys and a comment
- **THEN** the front-matter block is byte-identical afterwards

#### Scenario: Submitted body that looks like front matter is not promoted

- **WHEN** a submitted body begins with a `---` fence
- **THEN** it is written as body content and the document's existing front matter
  remains the only front matter

### Requirement: Emptying a document is explicit

`update_doc` SHALL refuse a resulting body that is empty or whitespace-only unless
the caller passes `allow_empty: true`.

#### Scenario: Accidental wipe refused

- **WHEN** `update_doc` submits an empty body without `allow_empty`
- **THEN** the write is refused and the document is unchanged

#### Scenario: Deliberate wipe allowed

- **WHEN** the same call passes `allow_empty: true`
- **THEN** the body is emptied and the front matter is preserved

### Requirement: update_doc runs the standard safe-write pipeline

`update_doc` SHALL be refused while the git working tree is degraded, SHALL write
atomically against the read-time content hash, SHALL reindex and return the
document's resulting term membership, and SHALL append an audit entry for every
attempt recording the tool, slug, outcome, and a unified diff of the change.

#### Scenario: Refused while degraded

- **WHEN** the working tree is conflicted and `update_doc` is called
- **THEN** the write is refused with a retryable error and the document is unchanged

#### Scenario: Successful update is audited as a diff

- **WHEN** `update_doc` succeeds
- **THEN** an audit entry with tool `update_doc` and outcome `ok` is appended, whose
  payload is a unified diff rather than the full document

#### Scenario: Membership is returned after the update

- **WHEN** `update_doc` succeeds
- **THEN** the result reports the document's resulting term membership
