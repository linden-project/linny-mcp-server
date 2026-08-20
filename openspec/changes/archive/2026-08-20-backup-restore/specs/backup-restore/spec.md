## ADDED Requirements

### Requirement: Snapshot backup of the source-of-truth data

`Backup` SHALL write a `tar.gz` snapshot of the corpus's source-of-truth data (the
content directory and `lindenConfig`), excluding the disposable index/state and VCS
directories.

#### Scenario: Backup contains content and config

- **WHEN** a corpus is backed up
- **THEN** the archive contains the content records and the lindenConfig files, and
  not the index/state or `.git` directories

### Requirement: Verified restore round-trip

`Restore` SHALL extract a snapshot back into the corpus so that a record deleted or
mutated after the backup is recovered byte-for-byte.

#### Scenario: Deleted record is recovered

- **WHEN** a corpus is backed up, a record is then deleted, and the snapshot is
  restored
- **THEN** the deleted record exists again with its original content

#### Scenario: Mutated record is recovered

- **WHEN** a record is modified after backup and the snapshot is restored
- **THEN** the record's content matches the backed-up version

### Requirement: Restore rejects path traversal

`Restore` SHALL reject archive entries whose paths escape the target directory
(absolute paths or `..` traversal), so a malicious archive cannot write outside the
corpus.

#### Scenario: Traversing entry rejected

- **WHEN** an archive entry has a path containing `..` that escapes the target
- **THEN** restore fails and writes nothing outside the target
