# synthetic-corpus Specification

## REMOVED Requirements

### Requirement: Deterministic generation

**Reason**: The synthetic corpus exists to test the engine and moves with it.

**Migration**: See the `synthetic-corpus` capability in `linden-index`.

### Requirement: Realistic flat corpus shape

**Reason**: Corpus shape is an engine test fixture concern.

**Migration**: See `linden-index`.

### Requirement: Matching lindenConfig and Hugo config

**Reason**: The fixture moves with the Hugo reference comparison.

**Migration**: See `linden-index`.

### Requirement: Edge-case records on demand

**Reason**: Edge-case fixtures move with the engine tests.

**Migration**: See `linden-index`.

### Requirement: Never derived from real data

**Reason**: Data-hygiene rule stays attached to the generator, which moves.

**Migration**: See `linden-index`. The rule is unchanged there.

### Requirement: Config files use the singular taxonomy name

**Reason**: A fixture rule owned by the generator.

**Migration**: See `linden-index`.
