package index

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/corpus"
)

// buildCorpus generates a synthetic corpus, builds and emits the index, and
// returns the corpus root, index root, and build report.
func buildCorpus(t *testing.T, opts corpus.Options) (root, indexRoot string, g *Graph, report *BuildReport) {
	t.Helper()
	root = t.TempDir()
	if err := corpus.Generate(root, opts); err != nil {
		t.Fatalf("generate: %v", err)
	}
	var err error
	g, report, err = Build(root)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	indexRoot = filepath.Join(t.TempDir(), "lindenIndex")
	if err := Emit(g, indexRoot); err != nil {
		t.Fatalf("emit: %v", err)
	}
	return root, indexRoot, g, report
}

func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
}

func TestBuildProducesAllHomeFiles(t *testing.T) {
	_, indexRoot, _, _ := buildCorpus(t, corpus.Options{Seed: 1, Count: 60, EnableEdgeCases: true})
	for _, name := range []string{
		fileTaxonomies, fileDocsStarred, fileDocsWithProps, fileDocsWithTitle,
		fileDocsTasksCount, fileIndexerInfo, fileTaxonomiesStarred, fileTermsStarred,
	} {
		var v any
		readJSON(t, filepath.Join(indexRoot, name), &v)
	}
}

func TestMultiTermMembership(t *testing.T) {
	_, indexRoot, _, _ := buildCorpus(t, corpus.Options{Seed: 2, Count: 20, EnableEdgeCases: true})
	for _, term := range []string{"work", "health"} {
		var members []string
		readJSON(t, filepath.Join(indexRoot, "tags", term, "index.json"), &members)
		if !contains(members, "work_and_health.md") {
			t.Fatalf("expected work_and_health.md in tags/%s membership, got %v", term, members)
		}
	}
}

func TestTitleGatingAndTasks(t *testing.T) {
	root := t.TempDir()
	write := func(name, content string) {
		if err := os.MkdirAll(filepath.Join(root, "content"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "content", name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("titled.md", "---\ntitle: Has Title\ntags: note\n---\n- [ ] a\n- [x] b\n")
	write("untitled.md", "---\ntags: note\n---\nno title here\n")

	g, _, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	indexRoot := filepath.Join(t.TempDir(), "idx")
	if err := Emit(g, indexRoot); err != nil {
		t.Fatal(err)
	}

	var props map[string]any
	readJSON(t, filepath.Join(indexRoot, fileDocsWithProps), &props)
	if _, ok := props["titled.md"]; !ok {
		t.Error("titled.md should be in docs_with_props")
	}
	if _, ok := props["untitled.md"]; ok {
		t.Error("untitled.md must be excluded from docs_with_props (no title)")
	}

	var tasks map[string]TaskCount
	readJSON(t, filepath.Join(indexRoot, fileDocsTasksCount), &tasks)
	if got := tasks["titled.md"]; got.Open != 1 || got.Closed != 1 || got.Total != 2 {
		t.Errorf("titled.md tasks = %+v, want open1 closed1 total2", got)
	}
	if _, ok := tasks["untitled.md"]; ok {
		t.Error("untitled.md has no tasks and must be excluded from tasks_count")
	}
}

func TestL1ConfigEmbeddedAndStarred(t *testing.T) {
	_, indexRoot, _, _ := buildCorpus(t, corpus.Options{Seed: 3, Count: 40, EnableEdgeCases: true})

	var l1 map[string]map[string]any
	readJSON(t, filepath.Join(indexRoot, "customer", "index.json"), &l1)
	eric, ok := l1["eric"]
	if !ok {
		t.Fatalf("expected term 'eric' in customer L1 index, got keys %v", keysOf(l1))
	}
	if eric["title"] != "Eric" {
		t.Errorf("customer/eric title = %v, want Eric", eric["title"])
	}

	// terms_starred entries must be well-formed {taxonomy, term}.
	var starred []StarredTerm
	readJSON(t, filepath.Join(indexRoot, fileTermsStarred), &starred)
	for _, s := range starred {
		if s.Taxonomy == "" || s.Term == "" {
			t.Fatalf("malformed starred term entry: %+v", s)
		}
	}
}

func TestMalformedSkippedConflictReported(t *testing.T) {
	_, indexRoot, _, report := buildCorpus(t, corpus.Options{Seed: 4, Count: 10, EnableEdgeCases: true})

	if len(report.Malformed) == 0 {
		t.Error("expected malformed_yaml.md to be reported")
	}
	if !report.HasProblems() {
		t.Error("expected conflict markers to be reported (HasProblems)")
	}
	// The malformed record must not appear in the props index.
	var props map[string]any
	readJSON(t, filepath.Join(indexRoot, fileDocsWithProps), &props)
	if _, ok := props["malformed_yaml.md"]; ok {
		t.Error("malformed_yaml.md must not appear in docs_with_props")
	}
}

func TestIdempotentRebuild(t *testing.T) {
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 8, Count: 30, EnableEdgeCases: true}); err != nil {
		t.Fatal(err)
	}
	emitOnce := func() map[string]string {
		g, _, err := Build(root)
		if err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(t.TempDir(), "idx")
		if err := Emit(g, out); err != nil {
			t.Fatal(err)
		}
		files := map[string]string{}
		_ = filepath.Walk(out, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			// _indexer_info.json embeds absolute paths that vary by temp dir; skip it.
			if filepath.Base(p) == fileIndexerInfo {
				return nil
			}
			b, _ := os.ReadFile(p)
			rel, _ := filepath.Rel(out, p)
			files[rel] = string(b)
			return nil
		})
		return files
	}
	a, b := emitOnce(), emitOnce()
	if len(a) != len(b) {
		t.Fatalf("file count differs: %d vs %d", len(a), len(b))
	}
	for name, ca := range a {
		if b[name] != ca {
			t.Fatalf("file %s differs between rebuilds", name)
		}
	}
}

func contains(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}

func keysOf[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
