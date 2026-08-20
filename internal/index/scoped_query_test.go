package index_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/authz"
	"github.com/linden-project/linny-mcp-server/internal/index"
)

// controlledStore writes a tiny, deterministic corpus and returns a populated
// store. It contains exactly the fixtures the scope-intersection test needs:
//   - both.md    tagged tags:[work, health]
//   - workonly.md tagged tags:[work]
//   - health.md   tagged tags:[health]
func controlledStore(t *testing.T) *index.Store {
	t.Helper()
	root := t.TempDir()
	content := filepath.Join(root, "content")
	if err := os.MkdirAll(content, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(content, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("both.md", "---\ntitle: Both\ntags: [work, health]\n---\nsecret backup notes\n")
	write("workonly.md", "---\ntitle: Work Only\ntags: work\n---\nwork backup notes\n")
	write("health.md", "---\ntitle: Health\ntags: health\n---\nhealth backup notes\n")

	// Declare `tags` as a taxonomy so Build recognizes it and creates memberships.
	cfg := filepath.Join(root, "lindenConfig")
	if err := os.MkdirAll(cfg, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg, "L1-CONF-TAX-tags.yml"), []byte("title: Tags\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	g, _, err := index.Build(root)
	if err != nil {
		t.Fatal(err)
	}
	store, err := index.OpenStore(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Populate(g); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func readSQL(t *testing.T, scopes ...string) (string, []any) {
	t.Helper()
	ss, err := authz.Parse(scopes)
	if err != nil {
		t.Fatal(err)
	}
	return ss.ReadableFilenamesSQL()
}

func TestScopeInterisectionWorkHealth(t *testing.T) {
	store := controlledStore(t)
	sub, args := readSQL(t, "read:*", "deny:taxonomy:tags:health")

	// both.md (work+health) must be excluded because health is denied, even
	// though read:* would otherwise grant it — deny is evaluated across ALL terms.
	health, err := store.DocsByTermScoped("tags", "health", sub, args)
	if err != nil {
		t.Fatal(err)
	}
	if len(health) != 0 {
		t.Fatalf("no health docs should be visible, got %v", health)
	}

	// work-only remains visible; both.md must not appear under work either.
	work, err := store.DocsByTermScoped("tags", "work", sub, args)
	if err != nil {
		t.Fatal(err)
	}
	if !containsStr(work, "workonly.md") {
		t.Fatalf("workonly.md should be visible, got %v", work)
	}
	if containsStr(work, "both.md") {
		t.Fatalf("both.md is health-tagged and must be excluded, got %v", work)
	}

	// Search is filtered in SQL: "backup" is in every body, but only workonly.md
	// survives the scope.
	hits, err := store.SearchScoped("backup", 10, sub, args)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]bool{}
	for _, h := range hits {
		files[h.Filename] = true
	}
	if !files["workonly.md"] || files["both.md"] || files["health.md"] {
		t.Fatalf("scoped search leaked denied docs: %v", files)
	}
}

func TestDenyByDefaultReadsNothing(t *testing.T) {
	store := controlledStore(t)
	sub, args := readSQL(t) // no scopes at all

	hits, err := store.SearchScoped("backup", 10, sub, args)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("deny-by-default should return no hits, got %v", hits)
	}
	tax, err := store.ListTaxonomiesScoped(sub, args)
	if err != nil {
		t.Fatal(err)
	}
	if len(tax) != 0 {
		t.Fatalf("deny-by-default should list no taxonomies, got %v", tax)
	}
}

func TestGetDocScopedDeniedIsMissing(t *testing.T) {
	store := controlledStore(t)
	sub, args := readSQL(t, "read:*", "deny:taxonomy:tags:health")

	// both.md is denied → indistinguishable from a nonexistent doc.
	if _, ok, err := store.GetDocScoped("both.md", sub, args); ok || err != nil {
		t.Fatalf("denied doc: ok=%v err=%v, want false/nil", ok, err)
	}
	// A truly missing doc behaves identically.
	if _, ok, err := store.GetDocScoped("nope.md", sub, args); ok || err != nil {
		t.Fatalf("missing doc: ok=%v err=%v, want false/nil", ok, err)
	}
	// workonly.md is readable.
	if _, ok, err := store.GetDocScoped("workonly.md", sub, args); !ok || err != nil {
		t.Fatalf("workonly.md: ok=%v err=%v, want true/nil", ok, err)
	}
}

func TestReadTaxonomyLimitsVisibility(t *testing.T) {
	store := controlledStore(t)
	// Only allow reads via the tags taxonomy term "work".
	sub, args := readSQL(t, "read:taxonomy:tags:work")

	doc, ok, err := store.GetDocScoped("workonly.md", sub, args)
	if err != nil || !ok || doc.Filename != "workonly.md" {
		t.Fatalf("workonly.md should be readable: ok=%v err=%v", ok, err)
	}
	// health.md has no work term → not readable.
	if _, ok, _ := store.GetDocScoped("health.md", sub, args); ok {
		t.Fatal("health.md should not be readable under read:taxonomy:tags:work")
	}
}

func containsStr(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
