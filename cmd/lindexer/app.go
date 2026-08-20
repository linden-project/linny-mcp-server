// Command lindexer is the standalone Linden indexer.
//
// It is a sibling entrypoint to linny-mcp: it parses YAML front matter, builds
// the taxonomy graph, persists to SQLite+FTS5, and emits the JSON index files
// described in docs/linden-index-spec.md. At scaffold stage only `version` is
// implemented; `build`, `watch`, and `verify` are stubbed for later epics.
package main

import (
	"fmt"
	"io"

	"github.com/linden-project/linny-mcp-server/internal/buildinfo"
)

const usage = `lindexer - the standalone Linden indexer

Usage:
  lindexer <command>

Commands:
  build       Full rebuild of the index (not yet implemented)
  watch       Incrementally update the index via fsnotify (not yet implemented)
  verify      Diff our JSON index against Hugo's output (not yet implemented)
  version     Print the version and exit
`

// Run executes the CLI and returns a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[1] {
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "lindexer %s\n", buildinfo.Version)
		return 0
	case "build", "watch", "verify":
		fmt.Fprintf(stderr, "lindexer: %s is not implemented yet\n", args[1])
		return 1
	case "help", "--help", "-h":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		fmt.Fprintf(stderr, "lindexer: unknown command %q\n\n%s", args[1], usage)
		return 2
	}
}
