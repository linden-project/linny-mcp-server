# Linny brand

Brand assets for **linny-mcp**. The illustrations descend from the Linnaeus mascot:
Carl Linnaeus with a magnifying glass, inspecting a leaf whose veins branch into a
taxonomy graph (the corpus as a living classification).

## Assets

- [`../../assets/hero.webp`](../../assets/hero.webp) — the README hero banner.
- [`brand-sheet.webp`](brand-sheet.webp) — size-reduced reference of the full brand
  sheet (icon, hero, logos, mascot, palette, badge mocks). The raw 2 MiB PNG is not
  committed; regenerate crisp exports from the source composite when needed.

## Taglines

- **Connect. Classify. Empower.**
- The Linnaeus-inspired MCP server for structured knowledge.

## Palette

| Swatch    | Hex       | Role                                  |
|-----------|-----------|---------------------------------------|
| Navy      | `#0F172A` | Wordmark, dark backgrounds, icon field |
| Green     | `#2E7D32` | Primary — "MCP SERVER", accents        |
| Leaf      | `#A8D05E` | Secondary green, foliage highlights    |
| Amber     | `#F5C542` | Taxonomy nodes, warm accent            |
| Blue      | `#4AA3DF` | Links, the MIT license badge           |
| Purple    | `#7E57C2` | Taxonomy node accent                   |
| Cream     | `#F7F1E6` | Banner / page background               |
| Sand      | `#E6D7B8` | Muted background, botanical linework   |

## Regenerating crisp exports

The current `assets/hero.webp` is cropped from the composite brand sheet, so it is
soft at large sizes. For pixel-crisp deliverables — a native-resolution hero, a 512²
icon for the GitHub social preview, and horizontal/vertical logos — re-export each
element individually from the source artwork rather than cropping the sheet.
