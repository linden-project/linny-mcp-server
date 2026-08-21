## ADDED Requirements

### Requirement: Config files use the singular taxonomy name

The generator SHALL name lindenConfig files with the singular taxonomy name
(`L1-CONF-TAX-<singular>.yml`, `L2-CONF-TAX-<singular>-TRM-<term>.yml`) and declare the
matching `singular: plural` map in the Hugo config, so the reference indexer resolves
the L2 config for every taxonomy.

#### Scenario: Plural taxonomy config named by singular

- **WHEN** a corpus uses the `tags` taxonomy (singular `tag`)
- **THEN** its config files are named `L1-CONF-TAX-tag.yml` and
  `L2-CONF-TAX-tag-TRM-<term>.yml`
