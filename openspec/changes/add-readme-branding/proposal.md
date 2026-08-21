## Why

The repo has no visual identity, and bean `linny-mcp-server-dxwr` asks for README
branding — specifically a hero image. A brand sheet exists
(`docs/linny-mcp-server-branding.png`, 1536×1024) with a Linnaeus mascot ("Linny"), the
tagline **"Connect. Classify. Empower."**, a color palette, and mocked badges. But it is
a *contact sheet* (seven assets + captions on one canvas, 2.0 MiB) — not something you
can embed. This change slices the individual assets out of it and gives the README a
proper hero, tagline, and license badge.

## What Changes

- Crop the individual assets out of the brand sheet into `assets/`:
  - `assets/hero.webp` — the "GitHub Banner / Hero" region (target ~1280×384, caption
    excluded), optimized to WebP well under the 1 MiB jj limit.
  - `assets/icon.png` — the 512² GitHub icon region (placeholder quality; see design).
  - `assets/logo-horizontal.webp` — the horizontal wordmark.
- Rework the top of `README.md`: centered `hero.webp` (descriptive alt), keep the
  `# linny-mcp` H1, add the tagline and the Linnaeus/"Carl" lore line, and a static
  **license: MIT** badge. The secret-hygiene callout stays exactly as-is.
- Keep the brand sheet as a versioned reference under `docs/brand/` (shrunk under the
  size limit) rather than committing the raw 2.0 MiB composite.
- Record the color palette + taglines in `docs/brand/README.md` so future surfaces
  (social preview, a Pages site) stay on-brand.

Out of scope (noted, not done): the `build/lint/tests` badges the sheet mocks require a
CI workflow (`ci.yml` running `nix flake check`) that we deferred — those badges land
with CI. Setting the GitHub **social preview** image is a repo-settings action, not a
file change.

## Capabilities

### New Capabilities
- `project-branding`: README hero + brand assets, tagline, and the versioned brand
  reference.

## Impact

- New: `assets/hero.webp`, `assets/icon.png`, `assets/logo-horizontal.webp`,
  `docs/brand/` (shrunk sheet + palette notes). Modified: `README.md` (hero, tagline,
  license badge). No code, no gate impact (assets/docs are outside the build closure).
- Fidelity caveat: assets cropped from the composite are soft/low-res (the icon
  especially); crisp native exports are a follow-up (design decision 1).
