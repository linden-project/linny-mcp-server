## Context

`docs/linny-mcp-server-branding.png` is a 1536×1024 **brand sheet**, not a hero: it
lays out a GitHub icon (512²), a "GitHub Banner / Hero (1280×384)", horizontal/vertical
logos, a mascot-and-elements strip, a color palette (greens/yellow/blue/purple/cream),
and mocked repository badges — each with a caption baked in. jj already refused to
snapshot it (2.0 MiB > the 1 MiB default), which is the right signal: a README hero
should be ~100–300 KB.

The mascot is Carl Linnaeus; the project's Hugo indexer predecessor was codenamed
"Carl", and our indexer is `lindexer` — so "the Linnaeus-inspired MCP server for
structured knowledge / Connect. Classify. Empower." is coherent lore, not decoration.

## Goals / Non-Goals

**Goals:** a real README hero + tagline + license badge; individual brand assets under
`assets/`; the sheet retained as a versioned reference; nothing over the file-size gate.

**Non-Goals:** the `build/lint/tests` badges (need CI — separate change); setting the
GitHub social-preview image (repo setting); a full Pages site.

## Decisions

- **D1 — Crop from the sheet now; regenerate crisp exports later.** The maintainer chose
  to crop rather than block on clean exports. Trade-off recorded: the sheet regions are
  downscaled and the banner has a caption to exclude, so cropped assets are soft — the
  512² icon especially, since its region on the sheet is far smaller than 512². Good
  enough for a README hero at ~820 px display width; a follow-up should replace them
  with native-resolution exports (and is the source of truth for the social preview).
- **D2 — WebP for photographic assets, PNG where needed.** The hero/logo are rich
  illustrations → WebP at high quality lands well under 1 MiB with no jj bump. The icon
  stays PNG. Cropping/encoding uses `imagemagick` + `cwebp` from nixpkgs
  (`nix run nixpkgs#imagemagick` / `#libwebp`), run at implementation time.
- **D3 — `assets/` for deliverables, `docs/brand/` for the reference.** Deliverable
  assets live in `assets/` (README-relative `src="assets/…"`, idiomatic for GitHub).
  The sheet is kept under `docs/brand/` **shrunk** under the size limit, plus a
  `docs/brand/README.md` capturing the palette hexes and taglines.
- **D4 — Keep the H1.** The hero carries the wordmark, but `# linny-mcp` stays for
  accessibility, the GitHub TOC, and package/registry listings. Alt text is descriptive
  ("Linny — the Linnaeus-inspired MCP server for structured knowledge").
- **D5 — License badge now, status badges with CI.** A static shields.io `license: MIT`
  badge is safe today; `build/lint/tests` badges wait for `ci.yml` running
  `nix flake check` (deferred in `add-release-process` D5). This change does not fake
  them.
- **D6 — One opaque hero, no dark/light split (for now).** The banner is a light-cream
  rectangle; on GitHub's dark theme it reads as an intentional self-contained banner. A
  `<picture>` + `prefers-color-scheme` dark variant is a possible refinement, deferred.

## Risks / Trade-offs

- **Cropped-asset fidelity (D1).** Soft assets, and the icon is not truly 512². Mitigate
  by treating these as v1 and tracking a "native exports" follow-up; the README hero at
  display width is the only high-visibility use.
- **Repo bloat.** Even shrunk, binaries live in git history forever. Keep them small
  (WebP, optimized) and few (hero, icon, logo, one reference sheet).
- **Relative `src` off-GitHub.** `assets/…` paths break if the README is rendered
  outside the repo (rare). A `raw.githubusercontent.com/.../main/...` URL is the
  fallback; not adopted (pins to `main`, uglier).

## Open Questions

- Do we want the dark-mode `<picture>` hero variant (D6) now or later?
- Should the CI badges (and thus `ci.yml`) be bundled with this, or shipped separately
  first so the badges are honest on day one?
