// Command lindexer is the standalone Linden indexer.
//
// It parses YAML front matter, builds the taxonomy graph, and emits the JSON
// index files described in docs/linden-index-spec.md. `watch` and `verify` are
// stubbed for later epics.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/linden-project/linny-mcp-server/internal/buildinfo"
	"github.com/linden-project/linny-mcp-server/internal/index"
)

// storeFile is the SQLite database filename inside a state directory.
const storeFile = "index.sqlite"

const usage = `lindexer - the standalone Linden indexer

Usage:
  lindexer <command> [flags]

Commands:
  build       Full rebuild of the index from a corpus
  search      Full-text search a persisted index store
  verify      Diff our JSON index against a reference (Hugo) index tree
  watch       Rebuild the index on corpus changes (fsnotify, debounced)
  version     Print the version and exit

Run 'lindexer <command> -h' for a command's flags.
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
	case "search":
		return searchCmd(args[2:], stdout, stderr)
	case "verify":
		return verifyCmd(args[2:], stdout, stderr)
	case "watch":
		return watchCmd(args[2:], stdout, stderr)
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
	indexRoot := fs.String("index", "lindenIndex", "JSON index output directory")
	stateDir := fs.String("state-dir", "", "if set, also persist the SQLite+FTS5 store to this directory")
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
	if *stateDir != "" {
		if err := persistStore(g, *stateDir); err != nil {
			fmt.Fprintf(stderr, "lindexer: persist failed: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "persisted SQLite+FTS5 store to %s\n", filepath.Join(*stateDir, storeFile))
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

// persistStore writes the graph to a SQLite+FTS5 store under stateDir.
func persistStore(g *index.Graph, stateDir string) error {
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return err
	}
	store, err := index.OpenStore(filepath.Join(stateDir, storeFile))
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()
	return store.Populate(g)
}

func searchCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(stderr)
	stateDir := fs.String("state-dir", "", "directory holding the persisted store (required)")
	limit := fs.Int("limit", 10, "maximum number of hits")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *stateDir == "" {
		fmt.Fprintln(stderr, "lindexer: search requires --state-dir")
		return 2
	}
	query := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if query == "" {
		fmt.Fprintln(stderr, "lindexer: search requires a query")
		return 2
	}

	store, err := index.OpenStore(filepath.Join(*stateDir, storeFile))
	if err != nil {
		fmt.Fprintf(stderr, "lindexer: opening store: %v\n", err)
		return 1
	}
	defer func() { _ = store.Close() }()

	hits, err := store.Search(query, *limit)
	if err != nil {
		fmt.Fprintf(stderr, "lindexer: search failed: %v\n", err)
		return 1
	}
	if len(hits) == 0 {
		fmt.Fprintf(stdout, "no matches for %q\n", query)
		return 0
	}
	for _, h := range hits {
		title := h.Title
		if title == "" {
			title = "(untitled)"
		}
		fmt.Fprintf(stdout, "%s\t%s\n    %s\n", h.Filename, title, h.Snippet)
	}
	return 0
}

func verifyCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(stderr)
	corpus := fs.String("corpus", ".", "corpus root")
	reference := fs.String("reference", "", "reference index directory to diff against (e.g. Hugo's lindenIndex) (required)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *reference == "" {
		fmt.Fprintln(stderr, "lindexer: verify requires --reference (a reference index directory)")
		return 2
	}

	g, _, err := index.Build(*corpus)
	if err != nil {
		fmt.Fprintf(stderr, "lindexer: build failed: %v\n", err)
		return 1
	}
	ours, err := os.MkdirTemp("", "linny-verify-*")
	if err != nil {
		fmt.Fprintf(stderr, "lindexer: %v\n", err)
		return 1
	}
	defer func() { _ = os.RemoveAll(ours) }()
	if err := index.Emit(g, ours); err != nil {
		fmt.Fprintf(stderr, "lindexer: emit failed: %v\n", err)
		return 1
	}

	discrepancies, err := index.VerifyDirs(ours, *reference)
	if err != nil {
		fmt.Fprintf(stderr, "lindexer: verify failed: %v\n", err)
		return 1
	}
	if len(discrepancies) == 0 {
		fmt.Fprintln(stdout, "verify: no discrepancies; our index matches the reference")
		return 0
	}
	for _, d := range discrepancies {
		fmt.Fprintf(stdout, "DRIFT %s: %s\n", d.File, d.Detail)
	}
	fmt.Fprintf(stderr, "lindexer: %d discrepancy(ies) vs reference\n", len(discrepancies))
	return 1
}

func watchCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("watch", flag.ContinueOnError)
	fs.SetOutput(stderr)
	corpus := fs.String("corpus", ".", "corpus root")
	stateDir := fs.String("state-dir", "", "state dir to persist the store (required)")
	indexRoot := fs.String("index", "", "if set, also re-emit JSON here on each refresh")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *stateDir == "" {
		fmt.Fprintln(stderr, "lindexer: watch requires --state-dir")
		return 2
	}

	refresh := func() error {
		g, report, err := index.Build(*corpus)
		if err != nil {
			return err
		}
		if err := persistStore(g, *stateDir); err != nil {
			return err
		}
		if *indexRoot != "" {
			if err := index.Emit(g, *indexRoot); err != nil {
				return err
			}
		}
		fmt.Fprintf(stdout, "refreshed index: %d records\n", report.RecordCount)
		return nil
	}

	if err := refresh(); err != nil { // build once up front
		fmt.Fprintf(stderr, "lindexer: initial build failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "watching %s for changes (Ctrl-C to stop)\n", *corpus)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err := index.Watch(ctx, *corpus, 0, func() {
		if err := refresh(); err != nil {
			fmt.Fprintf(stderr, "lindexer: refresh failed: %v\n", err)
		}
	})
	if err != nil && err != context.Canceled {
		fmt.Fprintf(stderr, "lindexer: watch error: %v\n", err)
		return 1
	}
	return 0
}
