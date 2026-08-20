package index

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/corpus"
)

// emitCorpus generates a corpus and emits its index into a fresh dir.
func emitCorpus(t *testing.T, seed int64) string {
	t.Helper()
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: seed, Count: 40, EnableEdgeCases: true}); err != nil {
		t.Fatal(err)
	}
	g, _, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "idx")
	if err := Emit(g, out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestVerifyIdenticalTrees(t *testing.T) {
	a := emitCorpus(t, 5)
	b := emitCorpus(t, 5) // same seed => same corpus => same index
	d, err := VerifyDirs(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 0 {
		t.Fatalf("identical trees should have no discrepancies, got %+v", d)
	}
}

func TestVerifyReorderedArrayIsEqual(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	writeJSONFile(t, filepath.Join(a, "tags", "note", "index.json"), []string{"x.md", "y.md", "z.md"})
	writeJSONFile(t, filepath.Join(b, "tags", "note", "index.json"), []string{"z.md", "x.md", "y.md"})
	d, err := VerifyDirs(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 0 {
		t.Fatalf("reordered array should be equal, got %+v", d)
	}
}

func TestVerifyReportsChangedMembershipAndMissingFile(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	// changed membership
	writeJSONFile(t, filepath.Join(a, "tags", "note", "index.json"), []string{"x.md", "y.md"})
	writeJSONFile(t, filepath.Join(b, "tags", "note", "index.json"), []string{"x.md"})
	// missing file (only in a)
	writeJSONFile(t, filepath.Join(a, "_index_taxonomies.json"), []string{"tags"})

	d, err := VerifyDirs(a, b)
	if err != nil {
		t.Fatal(err)
	}
	byFile := map[string]string{}
	for _, x := range d {
		byFile[x.File] = x.Detail
	}
	if byFile["tags/note/index.json"] != "content differs" {
		t.Fatalf("expected content-differs for membership, got %+v", d)
	}
	if _, ok := byFile["_index_taxonomies.json"]; !ok {
		t.Fatalf("expected missing-file discrepancy, got %+v", d)
	}
}

func TestVerifyIgnoresIndexerInfoIdentity(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	writeJSONFile(t, filepath.Join(a, "_indexer_info.json"), map[string]any{
		"product_name": "lindexer", "product_version": "1.2.3",
		"content_dir": "/n/content", "config_dir": "/n/lindenConfig", "index_dir": "/n/idx",
	})
	writeJSONFile(t, filepath.Join(b, "_indexer_info.json"), map[string]any{
		"product_name": "hugo-lindex", "hugo_version": "0.120",
		"content_dir": "/n/content", "config_dir": "/n/lindenConfig", "index_dir": "TODO",
	})
	d, err := VerifyDirs(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 0 {
		t.Fatalf("indexer_info identity/TODO fields should be ignored, got %+v", d)
	}
}

func writeJSONFile(t *testing.T, path string, v any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(v)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyIgnoreReferenceOnlyAndBuiltins(t *testing.T) {
	ours := t.TempDir()
	ref := t.TempDir()

	// Shared props file: ours lacks Hugo's built-ins; ref has them.
	writeJSONFile(t, filepath.Join(ours, "_index_docs_with_props.json"),
		map[string]any{"a.md": map[string]any{"title": "A", "tags": "note"}})
	writeJSONFile(t, filepath.Join(ref, "_index_docs_with_props.json"),
		map[string]any{"a.md": map[string]any{"title": "A", "tags": "note", "draft": false, "iscjklanguage": false}})

	// Reference-only extra file + an unparseable per-page file (Hugo emits raw
	// newlines in some per-page summaries).
	writeJSONFile(t, filepath.Join(ref, "extra_page", "index.json"), map[string]any{"x": 1})
	if err := os.MkdirAll(filepath.Join(ref, "bad_page"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ref, "bad_page", "index.json"), []byte("{\"summary\":\"line1\nline2\"}"), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := VerifyDirsWithOpts(ours, ref, VerifyOpts{IgnoreReferenceOnly: true})
	if err != nil {
		t.Fatalf("tolerant verify errored: %v", err)
	}
	if len(d) != 0 {
		t.Fatalf("expected no discrepancies (built-ins normalized, ref-only ignored), got %+v", d)
	}

	// Strict mode: the reference-only file is reported (and the unparseable file errors).
	dStrict, err := VerifyDirsWithOpts(ours, ref, VerifyOpts{})
	if err == nil && len(dStrict) == 0 {
		t.Fatal("strict mode should report reference-only files or fail on the unparseable one")
	}
}
