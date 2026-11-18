# indexer Specification

## REMOVED Requirements

### Requirement: Full rebuild from a corpus

**Reason**: The index engine moves to the `linden-index` repository as the public
`index` package.

**Migration**: Import `github.com/linden-project/linden-index/index` and call the
exported build function. Behaviour is unchanged.

### Requirement: Front matter is parsed directly

**Reason**: Front matter parsing is part of the extracted engine.

**Migration**: Use the `linden-index` parser. No MCP server code parses front
matter after this change.

### Requirement: Title gating and task counting match the spec

**Reason**: Title gating and task counting are engine behaviour and move with it.

**Migration**: See the `indexer` capability in `linden-index`; the rules are
unchanged.

### Requirement: lindenConfig drives L1 and starred indexes

**Reason**: `lindenConfig` handling belongs to the engine.

**Migration**: See `linden-index`. Notebook configuration files keep their format.

### Requirement: Malformed front matter never crashes the build

**Reason**: Defensive parsing moves with the parser.

**Migration**: See `linden-index`. The MCP server no longer builds indexes itself.

### Requirement: Committed conflict markers are reported

**Reason**: Conflict-marker detection is part of the build.

**Migration**: See `linden-index`. `lindexer build` still reports conflicting files.

### Requirement: Config is loaded and keyed by filename

**Reason**: Configuration loading is part of the engine.

**Migration**: See `linden-index`.

### Requirement: Starred indexes are config-derived with no occurrence filter

**Reason**: Starred index generation is engine behaviour.

**Migration**: See `linden-index`. Output shape is unchanged.
