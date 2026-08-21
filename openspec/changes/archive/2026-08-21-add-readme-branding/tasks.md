## 1. Crop assets from the brand sheet

- [x] 1.1 Measure the sheet regions (imagemagick) — hero banner, 512² icon, horizontal logo
- [x] 1.2 Crop the hero banner (exclude its caption) → optimize to `assets/hero.webp` (<300 KB)
- [x] 1.3 Icon + horizontal logo deferred — sheet crops are too soft to ship; moved to the
  native-export follow-up (they are not README deliverables; the spec requires only the hero)
- [x] 1.4 Verify every `assets/*` file is < 1 MiB (no jj size override) — `hero.webp` is 84 KB

## 2. Brand reference

- [x] 2.1 Move a size-reduced sheet to `docs/brand/brand-sheet.webp` (<1 MiB); drop the raw 2 MiB PNG
- [x] 2.2 `docs/brand/README.md`: palette hexes + taglines + "regenerate crisp exports" note

## 3. README

- [x] 3.1 Add the centered `assets/hero.webp` hero with descriptive alt, above the H1
- [x] 3.2 Add tagline "Connect. Classify. Empower." + the Linnaeus/Carl lore line
- [x] 3.3 Add a static `license: MIT` shields badge (no status badges yet)
- [x] 3.4 Leave the secret-hygiene callout and the rest of the README unchanged

## 4. Verify

- [x] 4.1 `nix flake check` green (assets/docs excluded — sanity only)
- [x] 4.2 Eyeball the rendered README hero at ~820 px display width

## Follow-ups (separate changes)

- Native-resolution asset exports (replace the soft hero crop; add a 512² `icon.png` and
  horizontal/vertical logos); the icon for the GitHub social preview (repo setting).
- `.github/workflows/ci.yml` running `nix flake check` → then add build/lint/tests
  badges.
- Optional dark-mode `<picture>` hero variant.
