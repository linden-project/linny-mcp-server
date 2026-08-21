package hugoref_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/corpus"
	"github.com/linden-project/linny-mcp-server/internal/hugoref"
	"github.com/linden-project/linny-mcp-server/internal/index"
)

// TestHugoRoundTrip runs the real Hugo reference indexer over a synthetic corpus
// and diffs our index against it, asserting ZERO drift: our indexer reproduces
// Hugo's output exactly (including its singular-keyed L1 term-config lookup).
func TestHugoRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("hugo"); err != nil {
		t.Skip("hugo not available")
	}
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 1, Count: 12, EnableEdgeCases: false}); err != nil {
		t.Fatal(err)
	}

	ref, cleanup, err := hugoref.BuildReference(root)
	if err != nil {
		t.Fatalf("build hugo reference: %v", err)
	}
	defer cleanup()

	g, _, err := index.Build(root)
	if err != nil {
		t.Fatal(err)
	}
	ours := filepath.Join(t.TempDir(), "idx")
	if err := index.Emit(g, ours); err != nil {
		t.Fatal(err)
	}

	d, err := index.VerifyDirsWithOpts(ours, ref, index.VerifyOpts{IgnoreReferenceOnly: true})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if len(d) != 0 {
		for _, x := range d {
			t.Errorf("drift: %s — %s", x.File, x.Detail)
		}
		t.Fatalf("expected zero drift vs the Hugo reference, got %d discrepancy(ies)", len(d))
	}
}
