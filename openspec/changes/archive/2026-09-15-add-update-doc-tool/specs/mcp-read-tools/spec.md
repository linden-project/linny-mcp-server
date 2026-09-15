## ADDED Requirements

### Requirement: get_doc reports a content hash and redaction status

`get_doc` SHALL return a `content_hash` for the document, computed from the file on
disk rather than from the index, and a `redacted` flag reporting whether the returned
body differs from the stored body because of egress redaction. Both fields are
additive; existing fields SHALL be unchanged.

#### Scenario: Hash is usable as a write base version

- **WHEN** `get_doc` returns a `content_hash` and `update_doc` is then called with it
  as `base_hash` while the file is untouched
- **THEN** the base version is accepted

#### Scenario: Hash tracks the file, not the index

- **WHEN** a document on disk has changed but the index has not yet been rebuilt
- **THEN** `get_doc` reports the hash of the current file

#### Scenario: Redaction is disclosed

- **WHEN** `get_doc` returns a body the redactor altered
- **THEN** `redacted` is true, signalling that whole-body replacement is unavailable
  for that document
