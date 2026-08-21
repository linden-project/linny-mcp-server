# project-branding Specification

## Purpose
TBD - created by archiving change add-readme-branding. Update Purpose after archive.
## Requirements
### Requirement: README opens with a hero image

`README.md` SHALL open with a centered hero image referencing a repo-relative asset,
with descriptive alt text, above the `# linny-mcp` H1 (which is retained).

#### Scenario: Hero present with alt text

- **WHEN** `README.md` is read
- **THEN** it contains a centered `<img>` whose `src` is a repo-relative `assets/` path
  and whose `alt` names the project, and the `# linny-mcp` H1 still follows

### Requirement: Brand assets are optimized and within the size limit

Each deliverable brand asset under `assets/` SHALL be under the repository's file-size
limit (1 MiB), stored as WebP (illustrations) or optimized PNG.

#### Scenario: Assets fit the gate

- **WHEN** the brand assets are added
- **THEN** every file under `assets/` is < 1 MiB and no jj snapshot-size override is
  needed

### Requirement: Tagline and license badge

The README SHALL show the tagline "Connect. Classify. Empower." and a static
`license: MIT` badge. It SHALL NOT display build/lint/tests status badges until a CI
workflow backs them.

#### Scenario: Tagline and license, no vaporware badges

- **WHEN** `README.md` is read
- **THEN** it contains the tagline and a MIT license badge, and no build/lint/tests
  status badge

### Requirement: Brand reference is retained, not the raw composite

The 2.0 MiB composite brand sheet SHALL NOT be committed as-is; a size-reduced
reference SHALL live under `docs/brand/`, alongside notes recording the palette and
taglines.

#### Scenario: Reference kept under the limit

- **WHEN** the brand reference is committed
- **THEN** it lives under `docs/brand/`, is < 1 MiB, and `docs/brand/` documents the
  palette and taglines

### Requirement: Secret-hygiene callout is preserved

The rework SHALL preserve the existing secret-hygiene `[!IMPORTANT]` callout verbatim.

#### Scenario: Callout intact

- **WHEN** the reworked README is read
- **THEN** the secret-hygiene callout (no token value in a Nix option) is present
  unchanged

