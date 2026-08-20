// Package index parses YAML front matter, builds the Linden taxonomy graph,
// persists it to SQLite+FTS5, and emits the JSON index files described in
// docs/linden-index-spec.md. It is the heart of the standalone indexer
// (cmd/lindexer) and is imported by the server for query and reindex.
package index
