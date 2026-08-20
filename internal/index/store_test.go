package index

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/corpus"
)

// populatedStore generates a synthetic corpus, builds the graph, and returns an
// open store populated from it, plus the db path.
func populatedStore(t *testing.T) (*Store, string) {
	t.Helper()
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 5, Count: 60, EnableEdgeCases: true}); err != nil {
		t.Fatalf("generate: %v", err)
	}
	g, _, err := Build(root)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	dbPath := filepath.Join(t.TempDir(), "index.sqlite")
	store, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := store.Populate(g); err != nil {
		t.Fatalf("populate: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store, dbPath
}

func TestOpenEmptyStore(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "empty.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()
	tax, err := store.ListTaxonomies()
	if err != nil {
		t.Fatal(err)
	}
	if len(tax) != 0 {
		t.Fatalf("fresh store should have no taxonomies, got %v", tax)
	}
}

func TestPopulateAndQuery(t *testing.T) {
	store, _ := populatedStore(t)

	tax, err := store.ListTaxonomies()
	if err != nil {
		t.Fatal(err)
	}
	if !contains(tax, "tags") || !contains(tax, "customer") {
		t.Fatalf("expected tags+customer taxonomies, got %v", tax)
	}

	docs, err := store.DocsByTerm("tags", "health")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(docs, "work_and_health.md") {
		t.Fatalf("expected work_and_health.md in tags/health, got %v", docs)
	}

	doc, ok, err := store.GetDoc("work_and_health.md")
	if err != nil || !ok {
		t.Fatalf("GetDoc ok=%v err=%v", ok, err)
	}
	if doc.Title != "Work And Health" || doc.Body == "" {
		t.Fatalf("unexpected doc %+v", doc)
	}
	if _, ok := doc.Props["tags"]; !ok {
		t.Fatalf("doc props missing tags: %v", doc.Props)
	}
}

func TestGetDocMissing(t *testing.T) {
	store, _ := populatedStore(t)
	if _, ok, err := store.GetDoc("does_not_exist.md"); ok || err != nil {
		t.Fatalf("missing doc: ok=%v err=%v, want false/nil", ok, err)
	}
}

func TestSearchRankedAndEmpty(t *testing.T) {
	store, _ := populatedStore(t)

	// "brain" appears in generated bodies ("second brain system").
	hits, err := store.Search("brain", 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected hits for 'brain'")
	}
	for _, h := range hits {
		if h.Filename == "" || h.Snippet == "" {
			t.Fatalf("hit missing fields: %+v", h)
		}
	}
	// Ranking: bm25 ascending (best first) => non-decreasing scores.
	for i := 1; i < len(hits); i++ {
		if hits[i].Score < hits[i-1].Score {
			t.Fatalf("hits not ordered by bm25: %v", hits)
		}
	}

	empty, err := store.Search("zzznotpresentquery", 5)
	if err != nil {
		t.Fatalf("empty search errored: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected no hits, got %v", empty)
	}
}

func TestIdempotentRepopulate(t *testing.T) {
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 2, Count: 40, EnableEdgeCases: true}); err != nil {
		t.Fatal(err)
	}
	g, _, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	store, err := OpenStore(filepath.Join(t.TempDir(), "idx.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()

	if err := store.Populate(g); err != nil {
		t.Fatal(err)
	}
	first, _ := store.DocsByTerm("tags", "health")

	if err := store.Populate(g); err != nil {
		t.Fatalf("re-populate: %v", err)
	}
	second, _ := store.DocsByTerm("tags", "health")

	if len(first) != len(second) {
		t.Fatalf("re-populate changed row count: %d -> %d (duplicates?)", len(first), len(second))
	}
}

func TestDeleteAndRebuildRecovers(t *testing.T) {
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 9, Count: 30, EnableEdgeCases: true}); err != nil {
		t.Fatal(err)
	}
	g, _, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}

	dbDir := t.TempDir()
	dbPath := filepath.Join(dbDir, "index.sqlite")

	build := func() []string {
		store, err := OpenStore(dbPath)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = store.Close() }()
		if err := store.Populate(g); err != nil {
			t.Fatal(err)
		}
		docs, _ := store.DocsByTerm("tags", "health")
		return docs
	}

	before := build()
	// Delete the DB file (and sidecars) and rebuild — the documented recovery.
	removeGlob(t, dbPath+"*")
	after := build()

	if len(before) != len(after) || !sameSet(before, after) {
		t.Fatalf("delete-and-rebuild changed results: %v vs %v", before, after)
	}
}

func removeGlob(t *testing.T, pattern string) {
	t.Helper()
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range matches {
		if err := os.Remove(m); err != nil {
			t.Fatalf("remove %s: %v", m, err)
		}
	}
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := map[string]int{}
	for _, x := range a {
		m[x]++
	}
	for _, x := range b {
		m[x]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}
