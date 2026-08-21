## 1. Crop assets from the brand sheet

- [ ] 1.1 Measure the sheet regions (imagemagick) — hero banner, 512² icon, horizontal logo
- [ ] 1.2 Crop the hero banner (exclude its caption) → optimize to `assets/hero.webp` (<300 KB)
- [ ] 1.3 Crop the icon → `assets/icon.png`; crop the horizontal logo → `assets/logo-horizontal.webp`
- [ ] 1.4 Verify every `assets/*` file is < 1 MiB (no jj size override)

## 2. Brand reference

- [ ] 2.1 Move a size-reduced sheet to `docs/brand/brand-sheet.webp` (<1 MiB); drop the raw 2 MiB PNG
- [ ] 2.2 `docs/brand/README.md`: palette hexes + taglines + "regenerate crisp exports" note

## 3. README

- [ ] 3.1 Add the centered `assets/hero.webp` hero with descriptive alt, above the H1
- [ ] 3.2 Add tagline "Connect. Classify. Empower." + the Linnaeus/Carl lore line
- [ ] 3.3 Add a static `license: MIT` shields badge (no status badges yet)
- [ ] 3.4 Leave the secret-hygiene callout and the rest of the README unchanged

## 4. Verify

- [ ] 4.1 `nix flake check` green (assets/docs excluded — sanity only)
- [ ] 4.2 Eyeball the rendered README hero at ~820 px display width

## Follow-ups (separate changes)

- Native-resolution asset exports (replace the soft crops); the icon for the GitHub
  social preview (repo setting).
- `.github/workflows/ci.yml` running `nix flake check` → then add build/lint/tests
  badges.
- Optional dark-mode `<picture>` hero variant.
