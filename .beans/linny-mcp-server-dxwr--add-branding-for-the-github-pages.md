---
# linny-mcp-server-dxwr
title: add branding for the github pages
status: completed
type: task
priority: normal
created_at: 2026-08-21T12:32:10Z
updated_at: 2026-08-21T14:26:58Z
---

I mean the github readme.md Specially the hero image



**OpenSpec change:** `add-readme-branding` (proposal ready; implementation = crop assets from the brand sheet + wire the README hero/tagline/license badge).


## Summary of Changes

Added Linnaeus-mascot branding to the README.

- `assets/hero.webp` (84 KB) — hero banner cropped from the brand sheet and optimized to WebP.
- README: centered hero above the H1 with descriptive alt, the tagline "Connect. Classify. Empower.", a Carl-Linnaeus lore line, and a static MIT license badge. Secret-hygiene callout and H1 preserved verbatim.
- `docs/brand/`: size-reduced `brand-sheet.webp` reference (146 KB) + `README.md` recording the palette hexes and taglines. Raw 2 MiB composite PNG kept gitignored as the crop source.
- CHANGELOG `[Unreleased]` gains a Branding bullet.

Icon + horizontal/vertical logos deferred to a native-export follow-up (sheet crops too soft; not README deliverables). Shipped via `scripts/ship-change.sh`; `nix flake check` green; OpenSpec change archived as `2026-08-21-add-readme-branding`.
