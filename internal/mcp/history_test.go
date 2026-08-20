package mcp

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/authz"
	"github.com/linden-project/linny-mcp-server/internal/corpus"
	"github.com/linden-project/linny-mcp-server/internal/index"
	"github.com/linden-project/linny-mcp-server/internal/redact"
)

// gitBackedReader generates a synthetic corpus, commits it to a real git repo,
// builds the store, and returns a reader whose corpusPath is that repo.
func gitBackedReader(t *testing.T, scopes ...string) *reader {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 5, Count: 20, EnableEdgeCases: true}); err != nil {
		t.Fatal(err)
	}
	env := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=t@t")
	git := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	git("init", "-q")
	git("add", "-A")
	git("commit", "-qm", "initial corpus")

	g, _, err := index.Build(root)
	if err != nil {
		t.Fatal(err)
	}
	store, err := index.OpenStore(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.Populate(g); err != nil {
		t.Fatal(err)
	}
	ss, err := authz.Parse(scopes)
	if err != nil {
		t.Fatal(err)
	}
	return newReader(store, redact.New(), ss, root)
}

func TestHistoryToolScopeEnforced(t *testing.T) {
	rd := gitBackedReader(t, "read:*", "deny:taxonomy:tags:health")

	// work_and_health.md is health-tagged → denied → history reports not-found.
	_, denied, err := rd.history(context.Background(), nil, historyIn{Slug: "work_and_health.md"})
	if err != nil {
		t.Fatal(err)
	}
	if denied.Found {
		t.Fatal("history of a health-denied doc must be not-found")
	}

	// A readable doc returns its commit(s).
	readable := firstReadableSlug(t, rd)
	_, ok, err := rd.history(context.Background(), nil, historyIn{Slug: readable})
	if err != nil {
		t.Fatal(err)
	}
	if !ok.Found || len(ok.Commits) == 0 {
		t.Fatalf("expected history for readable doc %q, got %+v", readable, ok)
	}
}

func TestDiffToolRedactsWorkingTreeSecret(t *testing.T) {
	rd := gitBackedReader(t, "read:*")
	slug := firstReadableSlug(t, rd)

	// Add a fake secret to the working-tree file (uncommitted) so diff shows it.
	path := filepath.Join(rd.corpusPath, contentDir, slug)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(b, []byte("\nAKIAIOSFODNN7EXAMPLE\n")...), 0o644); err != nil {
		t.Fatal(err)
	}

	_, out, err := rd.diff(context.Background(), nil, diffIn{Slug: slug, Ref: "HEAD"})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Found {
		t.Fatal("diff of a readable doc should be found")
	}
	if strings.Contains(out.Diff, "AKIAIOSFODNN7EXAMPLE") {
		t.Fatalf("secret leaked through diff: %s", out.Diff)
	}
	if !strings.Contains(out.Diff, "[REDACTED:aws-access-key]") {
		t.Fatalf("expected redaction placeholder in diff, got:\n%s", out.Diff)
	}
}

func TestChangedSinceScopeFiltered(t *testing.T) {
	rd := gitBackedReader(t, "read:*", "deny:taxonomy:tags:health")
	_, out, err := rd.changedSince(context.Background(), nil, changedSinceIn{Since: "1970-01-01"})
	if err != nil {
		t.Fatal(err)
	}
	for _, slug := range out.Docs {
		if slug == "work_and_health.md" {
			t.Fatalf("health-denied doc leaked in changed_since: %v", out.Docs)
		}
	}
	if len(out.Docs) == 0 {
		t.Fatal("expected some readable changed docs")
	}
}

// firstReadableSlug returns a slug the reader can read (via list+docs_by_term).
func firstReadableSlug(t *testing.T, rd *reader) string {
	t.Helper()
	_, tax, err := rd.listTaxonomies(context.Background(), nil, emptyIn{})
	if err != nil || len(tax.Taxonomies) == 0 {
		t.Fatalf("no taxonomies: %v", err)
	}
	_, terms, err := rd.terms(context.Background(), nil, termsIn{Taxonomy: tax.Taxonomies[0]})
	if err != nil || len(terms.Terms) == 0 {
		t.Fatalf("no terms: %v", err)
	}
	_, docs, err := rd.docsByTerm(context.Background(), nil, docsByTermIn{Taxonomy: tax.Taxonomies[0], Term: terms.Terms[0]})
	if err != nil || len(docs.Docs) == 0 {
		t.Fatalf("no docs: %v", err)
	}
	return docs.Docs[0]
}
