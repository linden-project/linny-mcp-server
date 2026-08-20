# synthetic-corpus Specification

## Purpose
TBD - created by archiving change synthetic-corpus-generator. Update Purpose after archive.
## Requirements
### Requirement: Deterministic generation

The generator SHALL be deterministic: given the same seed and options it SHALL
produce a byte-identical corpus. It SHALL NOT read the wall clock or use an unseeded
random source.

#### Scenario: Reproducible

- **WHEN** the generator runs twice with the same seed into two directories
- **THEN** the two directories are byte-identical

### Requirement: Realistic flat corpus shape

The generator SHALL write a flat content directory of `.md` records whose filenames
follow the slug convention, each with YAML front matter carrying a `title`, a
`crdate` (`"YYYY-MM-DD"`), an optional `starred` boolean, and one or more taxonomy
memberships across at least the taxonomies `tags`, `projects`, `customer`, `type`,
`subject`. Some records SHALL contain Markdown task lists.

#### Scenario: Records are well-formed and flat

- **WHEN** a corpus of N normal records is generated
- **THEN** every record sits directly in the content directory (no subdirectories)
  and parses as valid front matter + body

### Requirement: Matching lindenConfig and Hugo config

The generator SHALL emit a `lindenConfig/` directory with `L1-CONF-TAX-<tax>.yml` and
`L2-CONF-TAX-<tax>-TRM-<term>.yml` files for the taxonomies/terms it uses (including
some `starred: true`), and a Hugo `config.yaml` declaring the same taxonomies, so the
reference (Hugo) indexer can build the same corpus for `verify`.

#### Scenario: Config is consistent with content

- **WHEN** a corpus is generated
- **THEN** every taxonomy used in front matter has an L1 config, and Hugo's
  `taxonomies:` map lists the same taxonomies

### Requirement: Edge-case records on demand

When edge cases are enabled, the generator SHALL include: a unicode-heavy record, a
record with very long front matter, a record with an empty body, a record with
malformed YAML front matter, a record containing committed git conflict markers, and
at least one record containing **fake** secrets (e.g. an AWS-key-shaped string, a
private-key header) for the redaction filter to catch.

#### Scenario: Conflict-marker record present

- **WHEN** edge cases are enabled
- **THEN** at least one record contains a line beginning with `<<<<<<<`

#### Scenario: Fake-secret record present

- **WHEN** edge cases are enabled
- **THEN** at least one record contains a credential-shaped string that is not a real
  secret

### Requirement: Never derived from real data

The generator SHALL synthesize all content from fixed word lists and the seed. It
SHALL NOT read, import, or reference the real `secondbrain` corpus or any external
private source.

#### Scenario: Self-contained

- **WHEN** the generator runs
- **THEN** it requires no network and no external corpus input

