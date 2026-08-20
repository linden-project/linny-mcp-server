## ADDED Requirements

### Requirement: Index trees are diffed, arrays as sets

`VerifyDirs` SHALL compare two index directories file-by-file, reporting files present
on only one side and files whose content differs. JSON arrays SHALL be compared as
sets (order-insensitive); objects SHALL be compared key-by-key.

#### Scenario: Identical trees have no discrepancies

- **WHEN** two byte-equivalent index trees are compared
- **THEN** no discrepancies are reported

#### Scenario: Reordered array is not a discrepancy

- **WHEN** two index files hold the same membership in a different order
- **THEN** they are considered equal

#### Scenario: A changed membership is reported

- **WHEN** a document is present in one tree's term membership and absent in the
  other's
- **THEN** a discrepancy is reported for that file

#### Scenario: A missing file is reported

- **WHEN** a file exists in one tree but not the other
- **THEN** a discrepancy naming that file is reported

### Requirement: indexer_info identity fields are ignored

Comparison of `_indexer_info.json` SHALL ignore the identity fields
(`product_name`, `product_version`, `hugo_version`) and any field whose value is the
literal `"TODO"`, so a conforming indexer does not diff against Hugo's placeholders.

#### Scenario: Different product identity is not a discrepancy

- **WHEN** two `_indexer_info.json` differ only in `product_version` or a `"TODO"`
  path field
- **THEN** no discrepancy is reported

### Requirement: verify CLI reports drift and exits non-zero

`lindexer verify --corpus <c> --reference <dir>` SHALL build the index from the corpus
and diff it against the reference tree, print each discrepancy, and exit non-zero when
any discrepancy exists.

#### Scenario: Drift fails the command

- **WHEN** the built index diverges from the reference
- **THEN** `verify` prints the discrepancies and exits non-zero
