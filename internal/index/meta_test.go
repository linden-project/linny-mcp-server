package index

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadNotebookMetaMapsPluralToSingular pins the crossing that is easy to get
// wrong: front-matter keys are the plural taxonomy names, while L2 term configs
// are filed under the singular one.
func TestLoadNotebookMetaMapsPluralToSingular(t *testing.T) {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "content"))
	mkdirAll(t, filepath.Join(root, "config", "_default"))
	mkdirAll(t, filepath.Join(root, lindenConfigRel))

	write(t, filepath.Join(root, "config", "_default", "config.yaml"),
		"taxonomies:\n  tag: tags\n  customer: customer\n")
	write(t, filepath.Join(root, lindenConfigRel, "L2-CONF-TAX-tag-TRM-note.yml"), "title: Note\n")
	write(t, filepath.Join(root, lindenConfigRel, "L2-CONF-TAX-tag-TRM-Work Log.yml"), "title: Work Log\n")

	m := LoadNotebookMeta(root)

	if !m.Taxonomies["tags"] || !m.Taxonomies["customer"] {
		t.Fatalf("expected the plural keys as taxonomies, got %v", m.Taxonomies)
	}
	if !m.DeclaredTerms["tags"]["note"] {
		t.Errorf("term declared under the singular name not found for the plural key: %v", m.DeclaredTerms["tags"])
	}
	// Declared terms are stored normalized, so a config filename with a space
	// matches the form the indexer will compare against.
	if !m.DeclaredTerms["tags"]["work-log"] {
		t.Errorf("declared term not normalized: %v", m.DeclaredTerms["tags"])
	}
	if len(m.DeclaredTerms["customer"]) != 0 {
		t.Errorf("customer declares no terms, got %v", m.DeclaredTerms["customer"])
	}
}

func TestTermValuesShape(t *testing.T) {
	cases := []struct {
		name  string
		in    any
		terms []string
		ok    bool
	}{
		{"string", "acme", []string{"acme"}, true},
		{"list of strings", []any{"a", "b"}, []string{"a", "b"}, true},
		{"empty list", []any{}, []string{}, true},
		{"number", 3, nil, false},
		{"bool", true, nil, false},
		{"nil", nil, nil, false},
		{"mixed list", []any{"a", 3}, nil, false},
	}
	for _, c := range cases {
		terms, ok := TermValues(c.in)
		if ok != c.ok {
			t.Errorf("%s: ok = %v, want %v", c.name, ok, c.ok)
			continue
		}
		if len(terms) != len(c.terms) {
			t.Errorf("%s: terms = %v, want %v", c.name, terms, c.terms)
		}
	}
}

func TestNormalizeTermExported(t *testing.T) {
	if got := NormalizeTerm("Work Log"); got != "work-log" {
		t.Fatalf("NormalizeTerm = %q, want %q", got, "work-log")
	}
}

func mkdirAll(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, p, s string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}
