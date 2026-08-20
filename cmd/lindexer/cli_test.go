package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/corpus"
)

// TestBuildAndSearchCLI exercises the build + search commands end-to-end,
// covering buildCmd, persistStore, and searchCmd.
func TestBuildAndSearchCLI(t *testing.T) {
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 4, Count: 20, EnableEdgeCases: true}); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(root, "state")
	index := filepath.Join(root, "lindenIndex")

	var out, errOut bytes.Buffer
	code := Run([]string{"lindexer", "build", "--corpus", root, "--index", index, "--state-dir", state}, &out, &errOut)
	if code != 0 {
		t.Fatalf("build exit=%d stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "indexed") {
		t.Fatalf("build output = %q", out.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"lindexer", "search", "--state-dir", state, "--limit", "5", "brain"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("search exit=%d stderr=%q", code, errOut.String())
	}
	// "brain" appears in generated bodies; expect at least one hit line.
	if strings.TrimSpace(out.String()) == "" {
		t.Fatalf("search produced no output")
	}
}

func TestSearchRequiresStateDir(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Run([]string{"lindexer", "search", "brain"}, &out, &errOut); code != 2 {
		t.Fatalf("search without --state-dir exit=%d, want 2", code)
	}
}

func TestVerifyCLIMatches(t *testing.T) {
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 6, Count: 15, EnableEdgeCases: true}); err != nil {
		t.Fatal(err)
	}
	ref := filepath.Join(root, "reference-index")

	var out, errOut bytes.Buffer
	// Emit a reference index from the same corpus.
	if code := Run([]string{"lindexer", "build", "--corpus", root, "--index", ref}, &out, &errOut); code != 0 {
		t.Fatalf("build ref exit=%d: %s", code, errOut.String())
	}
	out.Reset()
	errOut.Reset()
	// verify our freshly-built index against that reference => no drift.
	if code := Run([]string{"lindexer", "verify", "--corpus", root, "--reference", ref}, &out, &errOut); code != 0 {
		t.Fatalf("verify exit=%d, want 0; stdout=%q stderr=%q", code, out.String(), errOut.String())
	}
	if !strings.Contains(out.String(), "no discrepancies") {
		t.Fatalf("verify output = %q", out.String())
	}
}

func TestVerifyCLIRequiresReference(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Run([]string{"lindexer", "verify", "--corpus", "."}, &out, &errOut); code != 2 {
		t.Fatalf("verify without --reference exit=%d, want 2", code)
	}
}

func TestWatchRequiresStateDir(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Run([]string{"lindexer", "watch", "--corpus", "."}, &out, &errOut); code != 2 {
		t.Fatalf("watch without --state-dir exit=%d, want 2", code)
	}
}
