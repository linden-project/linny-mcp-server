# mcp-write-tools Specification

## Purpose
TBD - created by archiving change mcp-write-tools. Update Purpose after archive.

## Requirements

### Requirement: Write tools exist and require write scope

The server SHALL expose MCP write tools `create_doc`, `append_to_doc`,
`set_front_matter`, `unset_front_matter`, and `archive`. `create_doc` SHALL require
`write:inbox` (or `write:*`); modifying an existing document SHALL require `write:*`,
or `write:inbox` when the document is a quarantined draft. A write to a document the
caller cannot read SHALL be reported as not-found.

#### Scenario: create_doc forbidden without write scope

- **WHEN** a caller with only `read:*` calls `create_doc`
- **THEN** the tool refuses and no file is written

#### Scenario: modifying an existing doc requires write:*

- **WHEN** a caller with only `read:*` calls `set_front_matter` on a readable document
- **THEN** the tool refuses

### Requirement: Agent-created documents are quarantined

`create_doc` SHALL place the new document in the quarantine term by default, and the
result SHALL report the document as quarantined.

#### Scenario: New doc is quarantined

- **WHEN** `create_doc` succeeds
- **THEN** the created document is a member of the quarantine term and the result's
  membership reflects it

### Requirement: Writes are refused while degraded

Write tools SHALL refuse to write while the git working tree is degraded
(conflicted / mid-operation), returning a retryable error, and SHALL leave the corpus
unchanged.

#### Scenario: Create refused during conflict

- **WHEN** the working tree contains committed conflict markers and `create_doc` is
  called
- **THEN** the write is refused and no file is created

### Requirement: Writes are atomic and optimistically concurrent

Writes SHALL be atomic (temp → fsync → rename). `create_doc` SHALL fail if the slug
already exists; edits SHALL fail with a retryable stale-write error if the file
changed since it was read.

#### Scenario: Create does not clobber an existing slug

- **WHEN** `create_doc` targets a slug that already exists
- **THEN** it fails and the existing file is unchanged

### Requirement: Front-matter edits preserve key order

`set_front_matter`, `unset_front_matter`, and `archive` SHALL edit the front matter
surgically (as a YAML node), preserving the order of untouched keys and comments.

#### Scenario: Setting a key keeps existing keys

- **WHEN** `set_front_matter` adds a key to a document
- **THEN** the previously-present keys remain in their original order

### Requirement: Writes reindex and return resulting membership

After a successful write the server SHALL reindex and return the document's resulting
term membership, so the agent observes the effect of its write.

#### Scenario: Setting a taxonomy key updates membership

- **WHEN** `set_front_matter` sets a taxonomy key to a term
- **THEN** the returned membership includes that (taxonomy, term)

### Requirement: Every write is audited

Every write attempt (success, denial, or error) SHALL be recorded to the append-only
audit log outside the corpus, including the tool, slug, and outcome.

#### Scenario: Successful write is audited

- **WHEN** `create_doc` succeeds
- **THEN** an audit entry with tool `create_doc` and outcome `ok` is appended

### Requirement: Front-matter values preserve the caller's type

`set_front_matter` SHALL write a value as the YAML node type corresponding to the
type the caller supplied: string, boolean, integer, float, null, or a sequence of
scalars. No value SHALL be rendered through generic string formatting. A JSON number
with no fractional part and within int64 range SHALL be written as an integer,
otherwise as a float. Types SHALL be taken from the caller and SHALL NOT be inferred
from a string's contents.

#### Scenario: A number stays a number

- **WHEN** `set_front_matter` sets a key to the JSON number `3`
- **THEN** the front matter contains an unquoted integer `3`, not the string `'3'`

#### Scenario: A string is never re-typed

- **WHEN** `set_front_matter` sets a key to the JSON string `"2026-09-14"`
- **THEN** the front matter contains a string value, not a YAML timestamp

#### Scenario: Null is representable

- **WHEN** `set_front_matter` sets a key to JSON `null`
- **THEN** the key is present with a null value and is not removed

### Requirement: Sequences are first-class front-matter values

`set_front_matter` SHALL accept a list of scalars and write it as a YAML sequence.
The resulting term membership SHALL include every element of the sequence.

#### Scenario: Setting a taxonomy key to several terms

- **WHEN** `set_front_matter` sets a declared taxonomy key to `["acme", "globex"]`
- **THEN** the front matter contains a two-element sequence and the returned
  membership includes both terms

#### Scenario: No stringified list ever reaches the corpus

- **WHEN** any list value is written
- **THEN** the file contains no value of the form `'[…]'` produced by formatting a
  list as a string

### Requirement: Nested structures are refused

`set_front_matter` SHALL refuse a value that is a mapping, or a list containing
anything other than scalars, and SHALL leave the document unchanged. Linny front
matter is flat.

#### Scenario: Nested mapping refused

- **WHEN** `set_front_matter` sets a key to a JSON object
- **THEN** the write is refused and the document is unchanged

### Requirement: Writes that the indexer would ignore are refused

When the key is a taxonomy declared by the notebook, `set_front_matter` SHALL refuse
a value that is neither a string nor a sequence of strings, because such a value
produces no term membership. The refusal SHALL name the taxonomy. Keys that are not
declared taxonomies SHALL accept any supported scalar type.

#### Scenario: Number on a taxonomy key is refused

- **WHEN** `set_front_matter` sets a declared taxonomy key to the number `3`
- **THEN** the write is refused, the message names the taxonomy, and the document is
  unchanged

#### Scenario: Number on an ordinary property is accepted

- **WHEN** `set_front_matter` sets a key that is not a declared taxonomy to `3`
- **THEN** the write succeeds and the value is an integer

### Requirement: add_term and remove_term manage taxonomy membership

The server SHALL expose `add_term` and `remove_term`, each taking a slug, a taxonomy,
and a term. Both SHALL be idempotent. `add_term` SHALL create the key as a sequence
when absent, promote an existing scalar value to a sequence before appending, and
make no change when the term is already a member. `remove_term` SHALL drop the term,
and SHALL remove the key entirely when the last term is removed. Membership
comparison SHALL use the indexer's normalized term form while the caller's spelling
is what gets written.

#### Scenario: Adding to an absent key

- **WHEN** `add_term` targets a document with no such taxonomy key
- **THEN** the key is created as a sequence containing the term

#### Scenario: Adding to a scalar value promotes it

- **WHEN** `add_term` targets a document whose taxonomy key holds a single scalar
- **THEN** the value becomes a sequence containing the original value and the new term

#### Scenario: Adding an existing term changes nothing

- **WHEN** `add_term` adds a term already present, differing only in case or spacing
- **THEN** the document is unchanged and the result reports success

#### Scenario: Removing the last term removes the key

- **WHEN** `remove_term` removes the only term of a taxonomy key
- **THEN** the key is absent from the front matter and no empty sequence remains

#### Scenario: Removing an absent term changes nothing

- **WHEN** `remove_term` removes a term that is not present
- **THEN** the document is unchanged and the result reports success

### Requirement: Newly coined terms are reported

A term with no declared term config SHALL still be written, and the result SHALL list
it under `new_terms` so that coining a term is visible to the caller.

#### Scenario: Undeclared term is written and reported

- **WHEN** `add_term` adds a term that has no `L2-CONF-TAX-<tax>-TRM-<term>.yml`
- **THEN** the write succeeds and the result lists the term under `new_terms`

#### Scenario: Declared term is not reported as new

- **WHEN** `add_term` adds a term that has a declared term config
- **THEN** `new_terms` is empty

### Requirement: Typed edits remain surgical

Writing a typed or sequence value SHALL preserve the order of untouched front-matter
keys and any comments, and SHALL leave the document body unchanged.

#### Scenario: Key order and comments survive a sequence write

- **WHEN** a sequence value is written to a document whose front matter carries
  several keys and a comment
- **THEN** the untouched keys keep their original order and the comment is retained

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
