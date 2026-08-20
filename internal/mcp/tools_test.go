package mcp

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/authz"
	"github.com/linden-project/linny-mcp-server/internal/corpus"
	"github.com/linden-project/linny-mcp-server/internal/index"
	"github.com/linden-project/linny-mcp-server/internal/redact"
)

func testReader(t *testing.T, scopes ...string) *reader {
	t.Helper()
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 5, Count: 40, EnableEdgeCases: true}); err != nil {
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

func TestGetDocRedactsSecrets(t *testing.T) {
	rd := testReader(t, "read:*")
	_, out, err := rd.getDoc(context.Background(), nil, getDocIn{Slug: "fake_secrets.md"})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Found {
		t.Fatal("fake_secrets.md should be readable under read:*")
	}
	for _, secret := range []string{"AKIAIOSFODNN7EXAMPLE", "ghp_0123456789", "BEGIN RSA PRIVATE KEY", "NL91ABNA0417164300"} {
		if strings.Contains(out.Body, secret) {
			t.Errorf("secret %q leaked through get_doc: body=%q", secret, out.Body)
		}
	}
}

func TestScopedToolsDenyHealth(t *testing.T) {
	rd := testReader(t, "read:*", "deny:taxonomy:tags:health")

	// work_and_health.md is tagged both; denying health must exclude it everywhere.
	_, doc, err := rd.getDoc(context.Background(), nil, getDocIn{Slug: "work_and_health.md"})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Found {
		t.Fatal("work_and_health.md must read as not-found when health is denied")
	}

	_, dbt, err := rd.docsByTerm(context.Background(), nil, docsByTermIn{Taxonomy: "tags", Term: "health"})
	if err != nil {
		t.Fatal(err)
	}
	if len(dbt.Docs) != 0 {
		t.Fatalf("no health docs should be visible, got %v", dbt.Docs)
	}

	// "health" must not appear in the terms of tags either.
	_, terms, err := rd.terms(context.Background(), nil, termsIn{Taxonomy: "tags"})
	if err != nil {
		t.Fatal(err)
	}
	for _, term := range terms.Terms {
		if term == "health" {
			t.Fatalf("denied term 'health' leaked in terms(): %v", terms.Terms)
		}
	}
}

func TestDenyByDefaultToolsEmpty(t *testing.T) {
	rd := testReader(t) // no scopes

	_, s, err := rd.search(context.Background(), nil, searchIn{Query: "brain", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Hits) != 0 {
		t.Fatalf("deny-by-default search should be empty, got %d hits", len(s.Hits))
	}
	_, tax, err := rd.listTaxonomies(context.Background(), nil, emptyIn{})
	if err != nil {
		t.Fatal(err)
	}
	if len(tax.Taxonomies) != 0 {
		t.Fatalf("deny-by-default list_taxonomies should be empty, got %v", tax.Taxonomies)
	}
}

func TestSearchScopedReturnsHits(t *testing.T) {
	rd := testReader(t, "read:*")
	_, s, err := rd.search(context.Background(), nil, searchIn{Query: "brain", Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Hits) == 0 {
		t.Fatal("expected hits for 'brain' under read:*")
	}
	if s.Hits[0].Filename == "" {
		t.Fatal("hit missing filename")
	}
}
