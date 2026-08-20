package hugoref_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/corpus"
	"github.com/linden-project/linny-mcp-server/internal/hugoref"
	"github.com/linden-project/linny-mcp-server/internal/index"
)

// TestHugoRoundTrip runs the real Hugo reference indexer over a synthetic corpus
// and diffs our index against it. It asserts every load-bearing file matches and
// that the only accepted drift is the L1 term-config indexes (Hugo emits {} for
// the singular≠plural taxonomies; ours embeds the config — spec §13 Q7).
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

	// Every discrepancy must be a top-level <tax>/index.json (L1 config) file —
	// the one accepted divergence. Anything else is unexpected drift.
	flagged := map[string]bool{}
	for _, x := range d {
		flagged[x.File] = true
		isL1 := strings.HasSuffix(x.File, "/index.json") && strings.Count(x.File, "/") == 1
		if !isL1 {
			t.Errorf("unexpected drift in %s: %s", x.File, x.Detail)
		}
	}

	// Load-bearing files must match Hugo exactly.
	mustMatch := []string{
		"_index_taxonomies.json",
		"_index_docs_tasks_count.json",
		"_index_docs_with_props.json",
		"_index_docs_starred.json",
		"_index_terms_starred.json",
		"tags/note/index.json", // an L2 membership
	}
	for _, f := range mustMatch {
		if flagged[f] {
			t.Errorf("load-bearing file %s must match Hugo, but it was flagged", f)
		}
	}
}
