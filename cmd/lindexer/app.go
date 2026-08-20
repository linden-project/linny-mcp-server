// Command lindexer is the standalone Linden indexer.
//
// It parses YAML front matter, builds the taxonomy graph, and emits the JSON
// index files described in docs/linden-index-spec.md. `watch` and `verify` are
// stubbed for later epics.
package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/linden-project/linny-mcp-server/internal/buildinfo"
	"github.com/linden-project/linny-mcp-server/internal/index"
)

const usage = `lindexer - the standalone Linden indexer

Usage:
  lindexer <command> [flags]

Commands:
  build       Full rebuild of the index from a corpus
  watch       Incrementally update the index via fsnotify (not yet implemented)
  verify      Diff our JSON index against Hugo's output (not yet implemented)
  version     Print the version and exit

Run 'lindexer build -h' for build flags.
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
	case "build":
		return buildCmd(args[2:], stdout, stderr)
	case "watch", "verify":
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

func buildCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(stderr)
	corpus := fs.String("corpus", ".", "corpus root (contains the content dir and lindenConfig)")
	indexRoot := fs.String("index", "lindenIndex", "index output directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	g, report, err := index.Build(*corpus)
	if err != nil {
		fmt.Fprintf(stderr, "lindexer: build failed: %v\n", err)
		return 1
	}
	if err := index.Emit(g, *indexRoot); err != nil {
		fmt.Fprintf(stderr, "lindexer: emit failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "indexed %d records into %s\n", report.RecordCount, *indexRoot)
	for _, m := range report.Malformed {
		fmt.Fprintf(stderr, "  WARN malformed front matter: %s (%s)\n", m.File, m.Detail)
	}
	for _, c := range report.Conflicted {
		fmt.Fprintf(stderr, "  ALERT committed conflict markers: %s (%s)\n", c.File, c.Detail)
	}
	if report.HasProblems() {
		fmt.Fprintf(stderr, "lindexer: corpus contains %d conflicted file(s); the server should treat the tree as degraded\n", len(report.Conflicted))
	}
	return 0
}
